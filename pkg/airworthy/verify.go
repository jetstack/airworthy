package airworthy

import (
	"fmt"
	"io"

	"github.com/jetstack/airworthy/pkg/gnupg"
	"golang.org/x/crypto/openpgp"
)

func (a *Airworthy) verify(flags *Flags, reader io.Reader) (*openpgp.Entity, error) {

	// set up gnupg keyring
	keyring, err := gnupg.TrustedKeyring()
	if err != nil {
		return nil, fmt.Errorf("error building keyring: %s", err)
	}
	for _, key := range keyring {
		a.log.Debugf("keyring contains: %s", gnupg.KeyToString(key))
	}

	// download signature
	var signatureURL string
	var armored bool
	if flags.SignatureArmored != "" {
		armored = true
		signatureURL = flags.SignatureArmored
	} else if flags.SignatureBinary != "" {
		signatureURL = flags.SignatureBinary
	} else {
		return nil, fmt.Errorf("no signature URL given")
	}

	signatureReader, err := a.Download(signatureURL)
	if err != nil {
		return nil, fmt.Errorf("error getting signature from %s: %s", signatureURL, err)
	}
	defer signatureReader.Reader.Close()

	if armored {
		return openpgp.CheckArmoredDetachedSignature(
			keyring,
			reader,
			signatureReader,
		)
	}

	return openpgp.CheckDetachedSignature(
		keyring,
		reader,
		signatureReader,
	)
}
