package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// Md5 MD5
func Md5(str string) []byte {
	hash := md5.New()
	_, _ = hash.Write([]byte(str))
	return hash.Sum(nil)
}

func Md5String(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has) //将[]byte转成16进制
}

// MD5Encode16 返回一个16位md5加密后的字符串
func MD5Encode16(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	bys := h.Sum(nil)
	res := make([]byte, 0)
	for i := 0; i < len(bys); i++ {
		if i%2 != 0 {
			res = append(res, bys[i])
		}
	}
	return hex.EncodeToString(res)
}
