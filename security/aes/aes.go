package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

func NewBlock(key []byte) (block cipher.Block, err error) {
	if block, err = aes.NewCipher(key); err != nil {
		err = errors.Wrapf(err, "[aes.NewBlock] aes.NewCipher failed")
	}
	return
}

func Encrypt(key []byte, block cipher.Block, org []byte) (ser []byte, err error) {
	if block == nil {
		return nil, errors.Errorf("[aes.Encrypt] block is nil")
	}
	if len(key) <= 0 {
		return nil, errors.Errorf("[aes.Encrypt] key is empty")
	}
	if len(org) <= 0 {
		return nil, errors.Errorf("[aes.Encrypt] org is empty")
	}

	blockSize := block.BlockSize()
	org = pkcs7Padding(org, blockSize)
	ser = make([]byte, len(org))

	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	blockMode.CryptBlocks(ser, org)
	return ser, nil
}

func Decrypt(key []byte, block cipher.Block, ser []byte) (org []byte, err error) {
	if len(key) <= 0 {
		return nil, errors.Errorf("[aes.Decrypt] key is empty")
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	org = make([]byte, len(ser))
	blockMode.CryptBlocks(org, ser)
	return pkcs7UnPadding(org)
}

func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	ciphertext = append(ciphertext, padText...)
	return ciphertext
}

func pkcs7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	unPadding := int(origData[length-1])
	if unPadding <= 0 || unPadding > length {
		return nil, errors.Errorf("[aes.pkcs7UnPadding] length error")
	}
	return origData[:(length - unPadding)], nil
}
