package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"time"
)

// A MemberNodeIdentity identifies a node participating in a community, belonging to a user.
// It consists of a certificate signed by the user's private key, along with a corresponding private key.
type MemberNodeIdentity struct {
	NodeID  string
	Cert    *x509.Certificate
	PrivKey *rsa.PrivateKey

	CertBytes       []byte
	PrivateKeyBytes []byte
}

// GenerateMemberNodeKeypair is called on the node
func GenerateMemberNodeKeypair() ([]byte, *rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	pubKey := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)

	return pubKey, privKey, nil
}

// GenerateMemberNodeCertificate is called on the client side
func GenerateMemberNodeCertificate(nodeID string, userIdentity *UserIdentity, publicKey *rsa.PublicKey) ([]byte, error) {
	serial, err := generateSerial()
	if err != nil {
		return nil, err
	}

	// TODO determine if its important that we do this on the node side rather than the client side.
	rawSubject, err := getMemberNodeDistinguishedNames(nodeID)
	if err != nil {
		return nil, err
	}

	rawIssuer, err := getUserDistinguishedNames(userIdentity.Username, userIdentity.UUID)
	if err != nil {
		return nil, err
	}

	cert := &x509.Certificate{
		SerialNumber:          serial,
		RawSubject:            rawSubject,
		RawIssuer:             rawIssuer,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(100, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	return x509.CreateCertificate(rand.Reader, cert, userIdentity.Cert, publicKey, userIdentity.PrivKey)
}

func getMemberNodeDistinguishedNames(uuid string) ([]byte, error) {
	rdnSequence := pkix.RDNSequence{
		pkix.RelativeDistinguishedNameSET{
			pkix.AttributeTypeAndValue{
				Type:  uidObjectIdentifier(),
				Value: uuid,
			},
		},
	}

	return asn1.Marshal(rdnSequence)
}
