package crypto

import (
	"crypto/md5"
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
