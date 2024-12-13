package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/mwinyimoha/card-validator-utils/errors"
)

const NonceSize = 12

type NewCipherFunc func(key []byte) (cipher.Block, error)

type NewGCMFunc func(block cipher.Block) (cipher.AEAD, error)

type IOFullReaderFunc func(reader io.Reader, buffer []byte) (int, error)

type DataEncryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type Encryptor struct {
	SecretKey    []byte
	NewCipher    NewCipherFunc
	NewGCM       NewGCMFunc
	IOFullReader IOFullReaderFunc
}

func NewEncryptor(secretKey string) (DataEncryptor, error) {
	key := []byte(secretKey)

	if len(key) != 32 {
		return nil, errors.NewErrorf(errors.BadRequest, "secret key must be 32 bytes long")
	}

	return &Encryptor{
		SecretKey:    key,
		NewCipher:    aes.NewCipher,
		NewGCM:       cipher.NewGCM,
		IOFullReader: io.ReadFull,
	}, nil
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	block, err := e.NewCipher(e.SecretKey)
	if err != nil {
		return "", errors.WrapError(err, errors.Internal, "could not create cypher block")
	}

	gcm, err := e.NewGCM(block)
	if err != nil {
		return "", errors.WrapError(err, errors.Internal, "could not create GCM block cipher")
	}

	nonce := make([]byte, NonceSize)
	if _, err := e.IOFullReader(rand.Reader, nonce); err != nil {
		return "", errors.WrapError(err, errors.Internal, "could not create nonce")
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	block, err := e.NewCipher(e.SecretKey)
	if err != nil {
		return "", errors.WrapError(err, errors.Internal, "could not create cypher block")
	}

	gcm, err := e.NewGCM(block)
	if err != nil {
		return "", errors.WrapError(err, errors.Internal, "could not create GCM block cipher")
	}

	ciphertextBytes, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.WrapError(err, errors.Internal, "could not decode ciphertext")
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return "", errors.NewErrorf(errors.Internal, "ciphertext too short")
	}

	nonce, ciphertextBytes := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", errors.NewErrorf(errors.Internal, "could not decrypt ciphertext")
	}

	return string(plaintext), nil
}
