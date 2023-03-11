package dex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Super simple test to verify HTTP over Unix Domain Socket
func TestServer(t *testing.T) {
	sockPath := "/tmp/test-dex-server.sock"
	driver := &NoopDriver{
		MockGetRes: &GetResult{
			Body: []byte("{}"),
		},
		MockSchemaRes: &SchemaResult{
			Schema: []byte("{}"),
		},
		MockInterfaceRes: &InterfaceResult{
			Hash: "",
			Interface: &Interface{
				SchemaHash:  "123",
				Description: "desc",
			},
		},
		MockImplementationsRes: &ImplementationsResult{
			Implementations: &Implementations{
				InterfaceHash: "abc",
				Map:           map[string]string{"ipfs": "ipfs_impl"},
			},
		},
	}

	server, err := NewServer(sockPath, driver)
	if err != nil {
		t.Error(err)
	}

	go server.Start()

	time.Sleep(1 * time.Second)

	client, err := NewClient(sockPath)
	assert.Nil(t, err)

	getRes, err := client.Get("abc")
	assert.Nil(t, err)
	assert.Equal(t, "{}", string(getRes.Body))

	schemaRes, err := client.Schema("abc")
	assert.Nil(t, err)
	assert.Equal(t, "{}", string(schemaRes.Schema))

	interfaceRes, err := client.Interface("abc")
	assert.Nil(t, err)
	assert.Equal(t, "desc", interfaceRes.Interface.Description)
	assert.Equal(t, "123", string(interfaceRes.Interface.SchemaHash))

	implementationsRes, err := client.Implementations("abc")
	assert.Nil(t, err)
	assert.Equal(t, "abc", implementationsRes.Implementations.InterfaceHash)
	assert.Equal(t, "ipfs_impl", implementationsRes.Implementations.Map["ipfs"])
}
