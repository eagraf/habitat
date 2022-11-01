package identity

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemberNodeCertVerification(t *testing.T) {
	user, err := GenerateNewUserCert("bob_ross", "uuid")
	assert.Nil(t, err)

	pub, priv, err := GenerateMemberNodeKeypair()
	assert.Nil(t, err)

	pubKey, err := x509.ParsePKCS1PublicKey(pub)
	assert.Nil(t, err)

	certBytes, err := GenerateMemberNodeCertificate("bobs_node", user, pubKey)
	assert.Nil(t, err)

	cert, _, _, err := GetCertFormats(certBytes, priv)
	assert.Nil(t, err)

	uuid, err := GetUIDFromName(&cert.Subject)
	assert.Nil(t, err)
	assert.Equal(t, "bobs_node", uuid)

	userUuid, err := GetUIDFromName(&cert.Issuer)
	assert.Nil(t, err)
	assert.Equal(t, "uuid", userUuid)

	pool := x509.NewCertPool()
	pool.AddCert(user.Cert)

	chains, err := cert.Verify(x509.VerifyOptions{
		Roots: pool,
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageAny,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(chains))
	assert.Equal(t, 2, len(chains[0]))
}
