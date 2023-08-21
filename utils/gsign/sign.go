package gsign

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"strings"
)

// 签名算法方式
const (
	SignTypeMD5        = `MD5`
	SignTypeHMACSHA256 = `HMAC-SHA256`
)

// CalculateSign 计算签名
func CalculateSign(content, signType, key string) (string, error) {
	var h hash.Hash
	if signType == SignTypeHMACSHA256 {
		h = hmac.New(sha256.New, []byte(key))
	} else {
		h = md5.New()
	}

	if _, err := h.Write([]byte(content)); err != nil {
		return ``, err
	}
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil))), nil
}

// ParamSign 计算所传参数的签名
func ParamSign(p map[string]string, key string) (string, error) {
	bizKey := "&key=" + key
	str := OrderParam(p, bizKey)

	var signType string
	switch p["sign_type"] {
	case SignTypeMD5, SignTypeHMACSHA256:
		signType = p["sign_type"]
	case ``:
		signType = SignTypeMD5
	default:
		return ``, errors.New(`invalid sign_type`)
	}

	return CalculateSign(str, signType, key)
}
