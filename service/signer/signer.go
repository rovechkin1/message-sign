package signer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// SigningKey contains key id, public key and private key
type SigningKey struct {
	KeyId string
	pk    string
}

// KeyStore store of public/private key pairs
type KeyStore interface {
	// GetKeyById returns key by key id
	GetKeyById(keyId string) (*SigningKey, error)
	// GetKeyIds returns key ids
	GetKeyIds() ([]string, error)
}

func (c *SigningKey) Sign(msg string) (string, error) {
	data := []byte(msg)
	hash := crypto.Keccak256Hash(data)
	privateKey, err := crypto.HexToECDSA(c.pk)
	if err != nil {
		return "", err
	}

	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}
	signatureHex := hexutil.Encode(signature)
	return signatureHex, nil
}
