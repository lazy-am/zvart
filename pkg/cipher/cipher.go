package cipher

import (
	"crypto/aes"
	"crypto/ed25519"
	rand2 "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"math/rand"

	"github.com/lazy-am/zvart/pkg/random"
)

func GenerateSesKey() []byte {
	return GetSHA256([]byte(random.RandStringBytes(10)))
}

func RSAEncrypt(data []byte, key *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand2.Reader, key, data, nil)
}

func RSADecrypt(data []byte, privKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand2.Reader, privKey, data, nil)
}

func GetSHA256(src []byte) []byte {
	h := sha256.New()
	h.Write(src)
	return h.Sum(nil)
}

func AESEncript(key, data []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Calculate how many bytes should be added to the full block
	ad := byte(aes.BlockSize - (len(data) % aes.BlockSize))

	// We add until there is one byte left to the block
	for (len(data) % aes.BlockSize) != (aes.BlockSize - 1) {
		data = append(data, byte(rand.Intn(255)))
	}

	// Add as the first element the number of bytes to be removed during the decryption
	data = append([]byte{ad}, data...)

	// Encode all
	encoded := make([]byte, len(data))
	for j := 1; (j * aes.BlockSize) <= len(data); j++ {
		i := (j - 1) * aes.BlockSize
		cipher.Encrypt(encoded[i:], data[i:])
	}

	return encoded, nil

}

func AESDecript(key, data []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if (len(data) % aes.BlockSize) != 0 {
		return nil, errors.New("no multiplicity of aes block")
	}

	//Decrypt all data
	decoded := make([]byte, len(data))
	for j := 1; (j * aes.BlockSize) <= len(data); j++ {
		i := (j - 1) * aes.BlockSize
		cipher.Decrypt(decoded[i:], data[i:])
	}
	ad := decoded[0]
	if ad > (aes.BlockSize + 1) {
		return nil, errors.New("weird headline")
	}

	//Delete unnecessary stuff
	decoded = decoded[1:]
	decoded = decoded[:(len(decoded) - int(ad-1))]
	return decoded, nil
}

func GeneratePrivEd25519() (ed25519.PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}
	return priv, nil
}
