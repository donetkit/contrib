package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
)

// PKCS7 填充模式
func pKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	latest := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, latest...)
}

// 填充的反向操作，删除填充字符串
func pKCS7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("data is empty")
	} else {
		padding := int(data[length-1])
		return data[:(length - padding)], nil
	}
}

// aesEncrypt 加密
func aesEncrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	data = pKCS7Padding(data, blockSize)
	blocMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encrypted := make([]byte, len(data))
	blocMode.CryptBlocks(encrypted, data)
	return encrypted, nil
}

// aesDeCrypt 实现解密
func aesDeCrypt(data []byte, key []byte) (string, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(data))
	blockMode.CryptBlocks(origData, data)
	origData, err = pKCS7UnPadding(origData)
	if err != nil {
		return "", err
	}
	return string(origData), err
}

// EnCrypt 加密
func EnCrypt(str, key string) (string, error) {
	result, err := aesEncrypt([]byte(str), []byte(key))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(result), nil
}

// DeCrypt 解密
func DeCrypt(str, key string) (string, error) {
	temp, _ := hex.DecodeString(str)
	res, err := aesDeCrypt(temp, []byte(key))
	if err != nil {
		return "", err
	}
	return res, nil
}
