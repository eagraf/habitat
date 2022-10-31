package identity

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"math/big"
)

func commonNameObjectIdentifier() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{
		2, 5, 4, 3,
	}
}

func uidObjectIdentifier() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{
		0, 9, 2342, 19200300, 100, 1, 1,
	}
}

func generateSerial() (*big.Int, error) {
	serialBytes := make([]byte, 20)
	_, err := rand.Read(serialBytes)
	if err != nil {
		return nil, err
	}

	s := new(big.Int)
	s.SetBytes(serialBytes)

	return s, nil
}

func GetCertFromPEM(certPEMBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEMBytes)
	certBytes := block.Bytes

	return x509.ParseCertificate(certBytes)
}

func GetCertFormats(certBytes []byte, privKey *rsa.PrivateKey) (certificate *x509.Certificate, certPemBytes []byte, keyPemBytes []byte, err error) {
	certPem := new(bytes.Buffer)
	pem.Encode(certPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	keyPem := new(bytes.Buffer)
	pem.Encode(keyPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	// The x509 libraries interface is pretty stupid, making us reparse a struct we already encoded
	// to get proper behavior
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return
	}

	certificate = cert
	certPemBytes = certPem.Bytes()
	keyPemBytes = keyPem.Bytes()
	return
}

func GetUIDFromName(name *pkix.Name) (string, error) {
	for _, n := range name.Names {
		if n.Type.Equal(uidObjectIdentifier()) {
			if s, ok := n.Value.(string); ok {
				return s, nil
			} else {
				return "", errors.New("matching object identifier for UID does not give a string value")
			}
		}
	}

	return "", errors.New("object identifier for UID not found")
}
