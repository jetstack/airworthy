package airworthy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cavaliercoder/grab"
	"golang.org/x/crypto/openpgp"

	"github.com/jetstack/airworthy/pkg/cacertificates"
	"github.com/jetstack/airworthy/pkg/gnupg"
)

type Flags struct {
	ChecksumSHA256 string
	Signature      string
	URL            string
	Destination    string

	Unarchive *bool
}

type Airworthy struct {
	log *logrus.Entry
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
	if flags.Destination == "" {
		flags.Destination = filepath.Base(flags.URL)
		a.log.WithField("destination", flags.Destination).Debug("set destination")
	}

	return nil

}

func (a *Airworthy) checkDestination(flags *Flags) error {
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
	if stat, err := os.Stat(flags.Destination); os.IsNotExist(err) {
		exists = false
		// TODO: check if the directory exists
	} else if err != nil {
		return err
	} else if stat.Mode().IsRegular() {
		exists = true
	} else {
		return fmt.Errorf("not a regular file: %s", flags.Destination)
	}

	// gnupg
	keyring, err := gnupg.TrustedKeyring()
	if err != nil {
		return fmt.Errorf("error building keyring: %s", err)
	}
	for _, key := range keyring {
		a.log.Debugf("keyring contains: %s", gnupg.KeyToString(key))
	}

	// setup http client with built in CA
	rootCAs, err := cacertificates.Roots()
	if err != nil {
		return err
	}

	// download signature
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: rootCAs,
		},
	}
	httpClient := &http.Client{
		Transport: httpTransport,
	}
	sigReq, err := http.NewRequest("GET", flags.Signature, nil)
	if err != nil {
		return err
	}
	sigResp, err := httpClient.Do(sigReq)
	if err != nil {
		return err
	}
	defer sigResp.Body.Close()

	if sigResp.StatusCode > 400 {
		return fmt.Errorf("unexpected response code %d from signature (%s)", sigResp.StatusCode, flags.Signature)
	}

	if !exists {
		// download binary
		grabClient := grab.NewClient()
		grabClient.HTTPClient.Transport = httpTransport

		req, err := grab.NewRequest(flags.Destination, flags.URL)
		if err != nil {
			return err
		}

		// start download
		a.log.Infof("downloading %v...\n", req.URL())
		resp := grabClient.Do(req)
		a.log.Infof("  %v\n", resp.HTTPResponse.Status)

		// start UI loop
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()

	Loop:
		for {
			select {
			case <-t.C:
				a.log.Debugf("  transferred %v / %v bytes (%.2f%%)\n",
					resp.BytesComplete(),
					resp.Size,
					100*resp.Progress())

			case <-resp.Done:
				// download is complete
				break Loop
			}
		}
		// check for errors
		if err := resp.Err(); err != nil {
			return err
		}

		a.log.Infof("download saved to ./%v \n", resp.Filename)
	}

	file, err := os.Open(flags.Destination)
	if err != nil {
		return err
	}
	defer file.Close()

	signer, err := openpgp.CheckArmoredDetachedSignature(
		keyring,
		file,
		sigResp.Body,
	)
	if err != nil {
		return err
	}

	a.log.Infof("successfully signed by %s", gnupg.KeyToString(signer))

	return nil
}
