package identity

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserIdentityStorage(t *testing.T) {
	tempDir := t.TempDir()

	user, err := GenerateNewUserCert("bob_ross", "uuid")
	assert.Nil(t, err)

	uuid, err := GetUIDFromName(&user.Cert.Subject)
	assert.Nil(t, err)
	assert.Equal(t, "uuid", uuid)

	err = StoreUserIdentity(tempDir, user, []byte("likes_painting"))
	assert.Nil(t, err)

	reloadedUser, err := LoadUserIdentity(tempDir, "bob_ross", []byte("likes_painting"))
	assert.Nil(t, err)
	assert.NotNil(t, reloadedUser)
	assert.Equal(t, "bob_ross", reloadedUser.Username)

	assert.Equal(t, true, bytes.Equal(user.CertBytes, reloadedUser.CertBytes))
	assert.Equal(t, true, bytes.Equal(user.PrivateKeyBytes, reloadedUser.PrivateKeyBytes))

	assert.Equal(t, "bob_ross", reloadedUser.Cert.Subject.CommonName)

	uuid, err = GetUIDFromName(&reloadedUser.Cert.Subject)
	assert.Nil(t, err)
	assert.Equal(t, "uuid", uuid)
}
