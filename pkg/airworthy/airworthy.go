package airworthy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"golang.org/x/crypto/openpgp"

	"github.com/jetstack/airworthy/pkg/gnupg"
)

type Flags struct {
	ChecksumSHA256 string
	Signature      string
	URL            string
	Output         string

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
	if flags.Signature == "" && flags.ChecksumSHA256 == "" {
		flags.Signature = fmt.Sprintf("%s.asc", flags.URL)
		a.log.WithField("signature", flags.Signature).Debug("set signature to URL")
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

	// gnupg
	keyring, err := gnupg.TrustedKeyring()
	if err != nil {
		return fmt.Errorf("error building keyring: %s", err)
	}
	for _, key := range keyring {
		a.log.Debugf("keyring contains: %s", gnupg.KeyToString(key))
	}

	// download signature
	signatureReader, err := a.Download(flags.Signature)
	if err != nil {
		return fmt.Errorf("error getting signature: %s", err)
	}
	defer signatureReader.Reader.Close()

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

		fileOutput, err := os.OpenFile(flags.Output, os.O_RDWR|os.O_CREATE, flags.filePermission)
		if err != nil {
			return err
		}
		defer fileOutput.Close()

		// start download
		a.log.Infof("downloading %s", flags.URL)
		downloadReader, err := a.Download(flags.URL)
		if err != nil {
			return err
		}
		defer downloadReader.Reader.Close()

		// copy into file
		if _, err := io.Copy(fileOutput, downloadReader.Reader); err != nil {
			return fmt.Errorf("error during copy: %s", err)
		}

		fileOutput.Close()
		a.log.Infof("downloaded to %s", flags.Output)

	}

	file, err := os.Open(flags.Output)
	if err != nil {
		return err
	}
	defer file.Close()

	signer, err := openpgp.CheckArmoredDetachedSignature(
		keyring,
		file,
		signatureReader,
	)
	if err != nil {
		return err
	}

	a.log.Infof("successfully signed by %s", gnupg.KeyToString(signer))

	return nil
}
