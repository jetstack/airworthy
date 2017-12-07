package airworthy

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/jetstack/airworthy/pkg/gnupg"
)

func (a *Airworthy) checkSHA256Sum(hash []byte, reader io.Reader) error {
	h := sha256.New()

	_, err := io.Copy(h, reader)
	if err != nil {
		return err
	}

	hashIs := h.Sum([]byte{})

	if !reflect.DeepEqual(hashIs, hash) {
		return fmt.Errorf("sha256sum mismatch expected=%x actual=%x", hash, hashIs)
	}

	return nil
}

func (a *Airworthy) getSHA256Sum(flags *Flags) ([]byte, error) {
	filename := filepath.Base(flags.URL)

	sha256sumsReader, err := a.Download(flags.SHA256Sums)
	if err != nil {
		return []byte{}, fmt.Errorf("error getting sha256sums: %s", err)
	}

	// buffer sha256sums locally
	sha256sumsBuffer, err := ioutil.ReadAll(sha256sumsReader)
	if err != nil {
		return nil, err
	}
	sha256sumsBufferReader := bytes.NewReader(sha256sumsBuffer)

	signer, err := a.verify(flags, sha256sumsBufferReader)
	if err != nil {
		return []byte{}, fmt.Errorf("error checking signature for %s: %s", flags.SHA256Sums, err)
	}
	a.log.Infof("sha256sums successfully signed by %s", gnupg.KeyToString(signer))

	_, err = sha256sumsBufferReader.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(sha256sumsBufferReader)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "  ")
		if len(parts) != 2 {
			continue
		}
		if filename == parts[1] {
			return hex.DecodeString(parts[0])
		}
	}

	return []byte{}, fmt.Errorf("filename '%s' not found in %s", filename, flags.SHA256Sums)
}
