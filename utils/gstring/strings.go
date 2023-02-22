package gstring

import (
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	// NonceSymbols 随机字符串可用字符集
	NonceSymbols = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// NonceLength 随机字符串的长度
	NonceLength = 32
)

// GenerateNonce 生成一个长度为 NonceLength 的随机字符串（只包含大小写字母与数字）
func GenerateNonce() (string, error) {
	bytes := make([]byte, NonceLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	symbolsByteLength := byte(len(NonceSymbols))
	for i, b := range bytes {
		bytes[i] = NonceSymbols[b%symbolsByteLength]
	}
	return string(bytes), nil
}

// GenerateNonceLength 生成一个长度为 Length 的随机字符串（只包含大小写字母与数字）
func GenerateNonceLength(Length int) (string, error) {
	bytes := make([]byte, Length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	symbolsByteLength := byte(len(NonceSymbols))
	for i, b := range bytes {
		bytes[i] = NonceSymbols[b%symbolsByteLength]
	}
	return string(bytes), nil
}

// RandomString fun
func RandomString(length int) string {
	b := []byte(NonceSymbols)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, b[r.Intn(len(b))])
	}
	return string(result)
}

// Join concatenates the elements of its first argument to create a single string. The separator
// string sep is placed between elements in the resulting string.
func Join(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(elems[0])
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(s)
	}
	return b.String()
}

// NewLine new line
func NewLine() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// Sub fun
func Sub(str string, start int, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

// ContainsString fun
func ContainsString(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

// Contains fun
func Contains(target interface{}, obj interface{}) bool {
	targetValue := reflect.ValueOf(target)
	var v = reflect.TypeOf(target).Kind()
	switch v {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}
	return false
}

// PadLeft fun
func PadLeft(s string, pad string, plength int) string {
	for i := len(s); i < plength; i++ {
		s = pad + s
	}
	return s
}

// UppercaseFirst fun
func UppercaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// LowercaseFirst fun
func LowercaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// Len 字符串长度，包括控制码，一个汉字长度1
func Len(str string) int {
	return utf8.RuneCountInString(str)
}

func SubString(str string, begin, length int) string {
	rs := []rune(str)
	lth := len(rs)
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length

	if end > lth {
		end = lth
	}
	return string(rs[begin:end])
}

func StringToIntArray(arr []string) []int {
	var result = make([]int, len(arr))
	for index, val := range arr {
		result[index], _ = strconv.Atoi(val)
	}
	return result
}

func IntToStringArray(arr []int) []string {
	var result = make([]string, len(arr))
	for index, val := range arr {
		result[index] = strconv.Itoa(val)
	}
	return result
}

// RemoveRepeatSliceInt 元素去重
func RemoveRepeatSliceInt(slc []int) []int {
	if len(slc) <= 0 {
		return slc
	}
	if len(slc) < 1024 {
		// 切片长度小于1024的时候，循环来过滤
		return removeDuplicatesAndEmptyInt(slc)
	} else {
		// 大于的时候，通过map来过滤
		return removeDuplicateInt(slc)
	}
}

// 去除重复字符串和空格
func removeDuplicatesAndEmptyInt(a []int) (ret []int) {
	alen := len(a)
	for i := 0; i < alen; i++ {
		if i > 0 && a[i-1] == a[i] {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}

// 通过map主键唯一的特性过滤重复元素
func removeDuplicateInt(arr []int) []int {
	resArr := make([]int, 0)
	tmpMap := make(map[int]interface{})
	for _, val := range arr {
		//判断主键为val的map是否存在
		if _, ok := tmpMap[val]; !ok {
			resArr = append(resArr, val)
			tmpMap[val] = nil
		}
	}
	return resArr
}

// RemoveRepeatSlice 元素去重
func RemoveRepeatSlice(slc []string) []string {
	if len(slc) <= 0 {
		return slc
	}
	if len(slc) < 1024 {
		// 切片长度小于1024的时候，循环来过滤
		return removeDuplicatesAndEmpty(slc)
	} else {
		// 大于的时候，通过map来过滤
		return removeDuplicate(slc)
	}
}

// 去除重复字符串和空格
func removeDuplicatesAndEmpty(a []string) (ret []string) {
	alen := len(a)
	for i := 0; i < alen; i++ {
		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}

// 通过map主键唯一的特性过滤重复元素
func removeDuplicate(arr []string) []string {
	resArr := make([]string, 0)
	tmpMap := make(map[string]interface{})
	for _, val := range arr {
		if len(val) == 0 {
			continue
		}
		//判断主键为val的map是否存在
		if _, ok := tmpMap[val]; !ok {
			resArr = append(resArr, val)
			tmpMap[val] = nil
		}
	}
	return resArr
}
