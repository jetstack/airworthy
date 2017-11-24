package cacertificates

//go:generate go-bindata -pkg $GOPACKAGE -o cacertificates_bindata.go ca-certificates.crt
import (
	"crypto/x509"
)

func Roots() (*x509.CertPool, error) {
	roots := x509.NewCertPool()

	rootsPEM, err := caCertificatesCrtBytes()
	if err != nil {
		return nil, err
	}
	roots.AppendCertsFromPEM(rootsPEM)

	return roots, nil
}
