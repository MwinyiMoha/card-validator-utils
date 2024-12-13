package encryption

import (
	"crypto/cipher"
	"io"
	"testing"

	"github.com/mwinyimoha/card-validator-utils/errors"
	"github.com/stretchr/testify/assert"
)

var secretKey = "39b04101cac8b8f8c24f4780fd5f1950"

func TestNewEncryptor(t *testing.T) {
	t.Run("Valid Key", func(t *testing.T) {
		encryptor, err := NewEncryptor(secretKey)
		assert.NoError(t, err)
		assert.NotNil(t, encryptor)
	})

	t.Run("Invalid Key Length", func(t *testing.T) {
		shortSecretKey := "shortkey"
		encryptor, err := NewEncryptor(shortSecretKey)
		assert.Nil(t, encryptor)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret key must be 32 bytes long")
	})
}

func TestEncryptor_Encrypt(t *testing.T) {
	encryptor, _ := NewEncryptor(secretKey)

	t.Run("Encrypt Valid Plaintext", func(t *testing.T) {
		ciphertext, err := encryptor.Encrypt("Hello, World!")
		assert.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
	})

	t.Run("Encrypt Empty Plaintext", func(t *testing.T) {
		ciphertext, err := encryptor.Encrypt("")
		assert.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
	})
}

func TestEncryptor_Decrypt(t *testing.T) {
	encryptor, _ := NewEncryptor(secretKey)

	t.Run("Valid Ciphertext", func(t *testing.T) {
		plaintext := "Hello, World!"
		ciphertext, err := encryptor.Encrypt(plaintext)
		assert.NoError(t, err)

		decryptedText, err := encryptor.Decrypt(ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decryptedText)
	})

	t.Run("Invalid Base64 Ciphertext", func(t *testing.T) {
		invalidCiphertext := "invalid$$$-base64"
		decryptedText, err := encryptor.Decrypt(invalidCiphertext)
		assert.Error(t, err)
		assert.Empty(t, decryptedText)
		assert.Contains(t, err.Error(), "could not decode ciphertext")
	})

	t.Run("Short Ciphertext", func(t *testing.T) {
		shortCiphertext := "c2hvcnQK"
		decryptedText, err := encryptor.Decrypt(shortCiphertext)
		assert.Error(t, err)
		assert.Empty(t, decryptedText)
		assert.Contains(t, err.Error(), "ciphertext too short")
	})

	t.Run("Tampered Ciphertext", func(t *testing.T) {
		plaintext := "Hello, World!"
		ciphertext, _ := encryptor.Encrypt(plaintext)
		tamperedCiphertext := ciphertext[:len(ciphertext)-1]

		decryptedText, err := encryptor.Decrypt(tamperedCiphertext)
		assert.Error(t, err)
		assert.Empty(t, decryptedText)
		assert.Contains(t, err.Error(), "could not decrypt ciphertext")
	})
}

func TestEncryptor_ErrorScenarios(t *testing.T) {
	t.Run("Cipher Creation Error", func(t *testing.T) {
		encryptor, _ := NewEncryptor(secretKey)

		mockNewCipher := func(key []byte) (cipher.Block, error) {
			return nil, errors.NewErrorf(errors.Internal, "mock cipher creation error")
		}
		concreteEncryptor := encryptor.(*Encryptor)
		concreteEncryptor.NewCipher = mockNewCipher

		t.Run("Encrypt", func(t *testing.T) {
			_, err := concreteEncryptor.Encrypt("test plaintext")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock cipher creation error")
		})

		t.Run("Decrypt", func(t *testing.T) {
			_, err := concreteEncryptor.Decrypt("test-ciphertext")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock cipher creation error")
		})
	})

	t.Run("GCM Creation Error", func(t *testing.T) {
		encryptor, _ := NewEncryptor(secretKey)

		mockNewGCM := func(block cipher.Block) (cipher.AEAD, error) {
			return nil, errors.NewErrorf(errors.Internal, "mock GCM creation error")
		}
		concreteEncryptor := encryptor.(*Encryptor)
		concreteEncryptor.NewGCM = mockNewGCM

		t.Run("Encrypt", func(t *testing.T) {
			_, err := concreteEncryptor.Encrypt("test plaintext")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock GCM creation error")
		})

		t.Run("Decrypt", func(t *testing.T) {
			_, err := concreteEncryptor.Decrypt("test-ciphertext")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "mock GCM creation error")
		})
	})

	t.Run("Nonce Creation Error", func(t *testing.T) {
		encryptor, _ := NewEncryptor(secretKey)

		mockFullRead := func(reader io.Reader, buffer []byte) (int, error) {
			return 0, errors.NewErrorf(errors.Internal, "mock io full read error")
		}
		concreteEncryptor := encryptor.(*Encryptor)
		concreteEncryptor.IOFullReader = mockFullRead

		_, err := concreteEncryptor.Encrypt("test plaintext")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock io full read error")
	})
}
