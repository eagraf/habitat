package ipfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPFSInit(t *testing.T) {
	tempDir := t.TempDir()

	i := &IPFSInstance{
		IPFSPath: tempDir,
	}
	err := i.Init()
	assert.Nil(t, err)

	// read config
	config, err := i.Config()
	assert.Nil(t, err)

	config.Bootstrap = []string{"hello"}

	err = i.Configure(config)
	assert.Nil(t, err)

	config, err = i.Config()
	assert.Nil(t, err)
	assert.Equal(t, "hello", config.Bootstrap[0])
}
