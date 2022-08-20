package signer

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rovechkin1/message-sign/service/config"
	"log"
	"os"
	"path"
	"strings"
)

type fileKeyStore struct {
	keys map[string]SigningKey
}

func NewFileKeyStore() (KeyStore, error) {
	content, err := os.ReadFile(path.Join(config.GetKeysDir(), "keys.csv"))
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(content), "\n")
	keys := map[string]SigningKey{}
	for _, l := range lines {
		ks := strings.Split(l, ",")
		if len(ks) < 2 {
			continue
		}
		keys[ks[0]] = SigningKey{
			KeyId: ks[0],
			pk:    ks[1][2:],
		}
	}
	return &fileKeyStore{
		keys: keys,
	}, nil
}

func generateKeys() (map[string]SigningKey, error) {
	nKeys := 100
	keys := make(map[string]SigningKey)
	for iKey := 0; iKey < nKeys; iKey += 1 {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
		privateKeyHex := hexutil.Encode(crypto.FromECDSA(privateKey))

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, err
		}

		address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

		key := SigningKey{
			KeyId: address,
			pk:    privateKeyHex,
		}

		keys[key.KeyId] = key
	}
	return keys, nil
}

func (c *fileKeyStore) GetKeyById(keyId string) (*SigningKey, error) {
	if key, ok := c.keys[keyId]; ok {
		return &key, nil
	}
	return nil, fmt.Errorf("Cannot find key")
}

func (c *fileKeyStore) GetKeyIds() ([]string, error) {
	var keys []string
	for k, _ := range c.keys {
		keys = append(keys, k)
	}
	return keys, nil
}
