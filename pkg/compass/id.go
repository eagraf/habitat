package compass

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/rs/zerolog/log"
)

func PeerIDPrivKeyPath() string {
	return filepath.Join(HabitatPath(), "libp2p-priv.key")
}

func PeerIDPubKeyPath() string {
	return filepath.Join(HabitatPath(), "libp2p-pub.key")
}

func PeerID() peer.ID {
	_, pub := GetPeerIDKeyPair()
	id, err := peer.IDFromPublicKey(pub)
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("LOCAL PEER ID %s", id)
	return id
}

func GetPeerIDKeyPair() (crypto.PrivKey, crypto.PubKey) {
	privKeyPath := PeerIDPrivKeyPath()
	pubKeyPath := PeerIDPubKeyPath()

	var privKeyExists, pubKeyExists bool = true, true
	_, err := os.Stat(privKeyPath)
	if errors.Is(err, os.ErrNotExist) {
		privKeyExists = false
	}

	_, err = os.Stat(pubKeyPath)
	if errors.Is(err, os.ErrNotExist) {
		pubKeyExists = false
	}

	// generate keypair if none exists
	if !privKeyExists && !pubKeyExists {
		err := generateKeyPair()
		if err != nil {
			panic(err)
		}
	}

	// xor privKeyExists and pubKeyExists
	if privKeyExists != pubKeyExists {
		panic(errors.New("incomplete libp2p keypair - exiting"))
	}

	privKeyBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		panic(err)
	}
	privKey, err := crypto.UnmarshalEd25519PrivateKey(privKeyBytes)
	if err != nil {
		panic(err)
	}

	pubKeyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		panic(err)
	}
	pubKey, err := crypto.UnmarshalEd25519PublicKey(pubKeyBytes)
	if err != nil {
		panic(err)
	}

	// Generate keypair if it does not exist
	return privKey, pubKey
}

func generateKeyPair() error {
	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	if err != nil {
		return err
	}

	privKeyPath := PeerIDPrivKeyPath()
	privKeyFile, err := os.OpenFile(privKeyPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer privKeyFile.Close()

	privKeyBytes, err := privKey.Raw()
	if err != nil {
		return err
	}

	_, err = privKeyFile.Write(privKeyBytes)
	if err != nil {
		return err
	}

	pubKeyPath := PeerIDPubKeyPath()
	pubKeyFile, err := os.OpenFile(pubKeyPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer pubKeyFile.Close()

	pubKeyBytes, err := pubKey.Raw()
	if err != nil {
		return err
	}

	_, err = pubKeyFile.Write(pubKeyBytes)
	if err != nil {
		return err
	}

	return nil
}
