package tools

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/hmac"
    "crypto/md5"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "io"
    "os"
    "strings"

	"golang.org/x/crypto/ripemd160"
)

func Md5Sum(inputs ...string) string {
	message := strings.Join(inputs, "")
	md5str := md5.Sum([]byte(message))
	return hex.EncodeToString(md5str[:])
}

func Sha256Sum(inputs ...string) string {
	message := strings.Join(inputs, "")
	sha256str := sha256.Sum256([]byte(message))
	return hex.EncodeToString(sha256str[:])
}

func Rmd160Sum(inputs ...string) string {
	message := strings.Join(inputs, "")
	rmd160 := ripemd160.New()
	_, _ = io.WriteString(rmd160, message)
	return hex.EncodeToString(rmd160.Sum(nil))
}

func GetFileSha256Hash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		Errorf("error on opening file %v\n%v", filePath, err)
		return "", err
	}
	defer f.Close()

	h256 := sha256.New()
	if _, err := io.Copy(h256, f); err != nil {
		Errorf("error on calculating file %v hash\n%v", filePath, err)
		return "", err
	}

	hash := hex.EncodeToString(h256.Sum(nil))
	Debugf("file %v SHA256 hash %v", filePath, hash)

	return hash, nil
}

func GetFileMd5Hash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		Errorf("error on opening file %v\n%v", filePath, err)
		return "", err
	}
	defer f.Close()

	md5str := md5.New()
	if _, err := io.Copy(md5str, f); err != nil {
		Errorf("error on calculating file %v hash\n%v", filePath, err)
		return "", err
	}

	hash := hex.EncodeToString(md5str.Sum(nil))
	Debugf("file %v MD5 hash %v", filePath, hash)

	return hash, nil
}

func SignWithSecretKey(key string, inputs ...string) (string, error) {
	if len(key) < 1 {
		Error("empty secret key")
		return "", errors.New("empty secret key")
	}

	keyHash := sha256.Sum256([]byte(key))
	mac := hmac.New(sha256.New, keyHash[:])

	message := strings.Join(inputs, "")
	_, err := mac.Write([]byte(message))
	if err != nil {
		Errorf("error on hmac inputs %v\n%v", message, err)
		return "", err
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

func GCMEncrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(passphrase))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		Errorf("error on encrypt\n%v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		Errorf("error on encrypt\n%v", err)
	}
	ciphered := gcm.Seal(nonce, nonce, data, nil)
	return ciphered
}

func GCMDecrypt(data []byte, passphrase string) []byte {
	key := []byte(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		Errorf("error on decrypt\n%v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		Errorf("error on decrypt\n%v", err)
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphered := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphered, nil)
	if err != nil {
		Errorf("error on decrypt\n%v", err)
	}
	return plaintext
}
