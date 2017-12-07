package airworthy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"github.com/jetstack/airworthy/pkg/gnupg"
)

type Flags struct {
	SHA256Sums       string
	SignatureBinary  string
	SignatureArmored string

	URL    string
	Output string

	dirPermission  os.FileMode
	filePermission os.FileMode

	Unarchive *bool
}

type Airworthy struct {
	log *logrus.Entry

	httpClient *http.Client
}

func New(log *logrus.Entry) *Airworthy {
	return &Airworthy{
		log: log,
	}
}

func (a *Airworthy) Must(err error) {
	if err != nil {
		a.log.Fatal(err)
	}
}

func (a *Airworthy) initFlags(flags *Flags) error {

	if flags.URL == "" {
		return errors.New("no URL supplied")
	}

	// by default download signature file from same url and only append .asc
	if flags.SignatureArmored == "" && flags.SignatureBinary == "" {
		if flags.SHA256Sums == "" {
			flags.SignatureArmored = fmt.Sprintf("%s.asc", flags.URL)
		} else {
			flags.SignatureArmored = fmt.Sprintf("%s.asc", flags.SHA256Sums)
		}
		a.log.WithField("signature-armored", flags.SignatureArmored).Debug("guessed signature URL")
	}

	// by default set to base name of URL
	if flags.Output == "" {
		flags.Output = filepath.Base(flags.URL)
		a.log.WithField("output", flags.Output).Debug("set output")
	}

	flags.dirPermission = 0755
	flags.filePermission = 0755

	return nil

}

func (a *Airworthy) Run(flags *Flags) error {

	a.log.WithField("flags", flags).Debug()

	err := a.initFlags(flags)
	if err != nil {
		return err
	}

	// get checksum file, if neccessary
	var sha256sum []byte
	checkSHA256sum := flags.SHA256Sums != ""
	if checkSHA256sum {
		sha256sum, err = a.getSHA256Sum(flags)
		if err != nil {
			return err
		}
	}

	// check if exists
	var exists bool
	if stat, err := os.Stat(flags.Output); os.IsNotExist(err) {
		exists = false
		// TODO: check if the directory exists
	} else if err != nil {
		return err
	} else if stat.Mode().IsRegular() {
		exists = true
	} else {
		return fmt.Errorf("not a regular file: %s", flags.Output)
	}

	var contents io.ReadSeeker

	if !exists {
		dirPath := filepath.Dir(flags.Output)
		if dirPath != "" {
			//ensure dir
			if stat, err := os.Stat(dirPath); os.IsNotExist(err) {
				err := os.MkdirAll(dirPath, flags.dirPermission)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if !stat.Mode().IsDir() {
				return fmt.Errorf("output directory %s is not a directory", dirPath)
			}
		}

		// start download
		a.log.Infof("downloading %s", flags.URL)
		downloadReader, err := a.Download(flags.URL)
		if err != nil {
			return err
		}
		defer downloadReader.Reader.Close()

		// copy contents into buffer
		downloadBuffer, err := ioutil.ReadAll(downloadReader.Reader)
		if err != nil {
			return fmt.Errorf("error during download: %s", err)
		}
		a.log.Infof("downloaded to %s", flags.Output)

		contents = bytes.NewReader(downloadBuffer)
	}

	if exists {
		file, err := os.Open(flags.Output)
		if err != nil {
			return err
		}
		defer file.Close()
		contents = file
	}

	if checkSHA256sum {
		if err := a.checkSHA256Sum(sha256sum, contents); err != nil {
			return err
		}
		a.log.Infof("contents match sha256_hash=%x", sha256sum)
	} else {
		signer, err := a.verify(flags, contents)
		if err != nil {
			return err
		}
		a.log.Infof("contents successfully signed by %s", gnupg.KeyToString(signer))
	}

	// write to file if not existed
	if !exists {
		fileOutput, err := os.OpenFile(flags.Output, os.O_RDWR|os.O_CREATE, flags.filePermission)
		if err != nil {
			return err
		}
		defer fileOutput.Close()

		if _, err := contents.Seek(0, 0); err != nil {
			return err
		}

		if _, err := io.Copy(fileOutput, contents); err != nil {
			return err
		}

	}

	return nil
}
