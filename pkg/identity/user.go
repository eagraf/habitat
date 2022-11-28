package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/scrypt"
)

type UserIdentity struct {
	Username string
	UUID     string
	Cert     *x509.Certificate
	PrivKey  *rsa.PrivateKey

	CertBytes       []byte
	PrivateKeyBytes []byte
}

func usernameAllowed(username string) (bool, rune) {
	allowed := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890-_"
	for _, c := range username {
		if !strings.Contains(allowed, string(c)) {
			return false, c
		}
	}
	return true, ' '
}

func GenerateNewUserCert(username, uuid string) (*UserIdentity, error) {

	if allowed, c := usernameAllowed(username); !allowed {
		return nil, fmt.Errorf("username %s contains an invalid character: %c", username, c)
	}

	serial, err := generateSerial()
	if err != nil {
		return nil, err
	}

	rawSubject, err := getUserDistinguishedNames(username, uuid)
	if err != nil {
		return nil, err
	}

	ca := &x509.Certificate{
		RawSubject:            rawSubject,
		SerialNumber:          serial,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(100, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	cert, certPemBytes, keyPemBytes, err := GetCertFormats(caBytes, caPrivKey)
	if err != nil {
		return nil, err
	}

	return &UserIdentity{
		Username: username,
		UUID:     uuid,

		Cert:    cert,
		PrivKey: caPrivKey,

		CertBytes:       certPemBytes,
		PrivateKeyBytes: keyPemBytes,
	}, nil
}

func StoreUserIdentity(path string, user *UserIdentity, password []byte) error {

	// encrypt the private key using a password, and store it

	key, salt, err := deriveKey(password, nil)
	if err != nil {
		return err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, user.PrivateKeyBytes, nil)
	ciphertext = append(ciphertext, salt...)

	certPath, privKeyPath := keyPaths(path, user.Username)

	err = ioutil.WriteFile(privKeyPath, ciphertext, 0600)
	if err != nil {
		return err
	}

	// write the cert out in plaintext

	err = ioutil.WriteFile(certPath, user.CertBytes, 0600)
	if err != nil {
		return err
	}

	return nil
}

func LoadUserIdentity(path, username string, password []byte) (*UserIdentity, error) {
	certPath, privKeyPath := keyPaths(path, username)
	certPEMBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(certPEMBytes)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, err
	}

	salt := data[len(data)-32:]
	data = data[:len(data)-32]

	key, _, err := deriveKey([]byte(password), salt)
	if err != nil {
		return nil, err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	privKeyBytes, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	block, _ = pem.Decode(privKeyBytes)

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &UserIdentity{
		Username: username,
		Cert:     cert,
		PrivKey:  privKey,

		CertBytes:       certPEMBytes,
		PrivateKeyBytes: privKeyBytes,
	}, nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	if password == nil {
		return nil, nil, errors.New("received nil password byte array")
	}

	key, err := scrypt.Key(password, salt, 1048576, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

func keyPaths(path, username string) (string, string) {
	certPath := filepath.Join(path, fmt.Sprintf("%s.cert", username))
	privKeyPath := filepath.Join(path, fmt.Sprintf("%s.key", username))
	return certPath, privKeyPath
}

func getUserDistinguishedNames(username string, uuid string) ([]byte, error) {
	rdnSequence := pkix.RDNSequence{
		pkix.RelativeDistinguishedNameSET{
			pkix.AttributeTypeAndValue{
				Type:  commonNameObjectIdentifier(),
				Value: username,
			},
			pkix.AttributeTypeAndValue{
				Type:  uidObjectIdentifier(),
				Value: uuid,
			},
		},
	}

	return asn1.Marshal(rdnSequence)
}
