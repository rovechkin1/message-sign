package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"os"
	"strconv"
)

func main() {
	numRecords := 100
	var err error
	if len(os.Args) > 1 {
		if os.Args[1] == "-h" ||
			os.Args[1] == "--help" {
			fmt.Printf("Usage: key-generator [num_record]\n")
			fmt.Printf("\t num_record default is 100\n")
			return
		} else {
			numRecords, err = strconv.Atoi(os.Args[1])
			if err != nil {
				panic(err)
			}
		}
	}

	keys, err := generateKeys(numRecords)
	if err != nil {
		log.Fatal("can't generate keys")
	}

	f, err := os.Create("keys.csv")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	count := 0
	for _, v := range keys {
		_, err2 := f.WriteString(fmt.Sprintf("%s,%s\n", v.KeyId, v.pk))

		if err2 != nil {
			log.Fatal(err2)
		}
		count += 1
	}

	fmt.Printf("done, inserted %v records\n", count)
}

// SigningKey contains key id, public key and private key
type SigningKey struct {
	KeyId string
	pk    string
}

func generateKeys(nKeys int) (map[string]SigningKey, error) {
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
