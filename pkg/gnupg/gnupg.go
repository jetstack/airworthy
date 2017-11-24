package gnupg

//go:generate sh -c "exec go-bindata -pkg $GOPACKAGE -o public_keys_bindata.go *.asc"

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

type armoredKey func() ([]byte, error)

func TrustedKeyring() (openpgp.EntityList, error) {

	var keyRing openpgp.EntityList

	for _, key := range []armoredKey{jetstackReleasesAscBytes, hashicorpAscBytes} {
		data, err := key()
		if err != nil {
			return nil, err
		}
		block, err := armor.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		entity, err := openpgp.ReadEntity(packet.NewReader(block.Body))
		if err != nil {
			return nil, err
		}
		keyRing = append(keyRing, entity)
	}

	return keyRing, nil
}

func KeyToString(key *openpgp.Entity) string {
	var identities []string
	for _, val := range key.Identities {
		identities = append(identities, val.Name)
	}
	return fmt.Sprintf("id=%s (%s)", key.PrimaryKey.KeyIdString(), strings.Join(identities, ", "))
}
