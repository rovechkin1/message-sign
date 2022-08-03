package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"os"
)

func main() {
	keys, err := generateKeys()
	if err != nil {
		log.Fatal("can't generate keys")
	}

	f, err := os.Create("keys.csv")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	for _, v := range keys {
		_, err2 := f.WriteString(fmt.Sprintf("%s,%s\n", v.KeyId, v.pk))

		if err2 != nil {
			log.Fatal(err2)
		}
	}

	fmt.Println("done")
}

// SigningKey contains key id, public key and private key
type SigningKey struct {
	KeyId string
	pk    string
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

		publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

		key := SigningKey{
			KeyId: hexutil.Encode(publicKeyBytes),
			pk:    privateKeyHex,
		}

		keys[key.KeyId] = key
	}
	return keys, nil
}
