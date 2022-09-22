package gstring

// https://github.com/Zoelov/IDToCode

import (
	"container/list"
	"errors"
	"fmt"
	"math"
	"strconv"
)

var baseCodeStr = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// GenCodeBase34 -- 将id转换成6位长度的code
func GenCodeBase34(id uint64) []byte {
	num := id
	mod := uint64(0)
	l := list.New()

	baseCodeByte := []byte(baseCodeStr)

	for num != 0 {
		mod = num % 34
		num = num / 34

		l.PushFront(baseCodeByte[int(mod)])
	}

	listLen := l.Len()

	var res []byte
	if listLen >= 6 {
		res = make([]byte, 0, listLen)
		for i := l.Front(); i != nil; i = i.Next() {
			res = append(res, i.Value.(byte))
		}
	} else {
		res = make([]byte, 0, 6)
		for i := 0; i < 6; i++ {
			if i < 6-listLen {
				res = append(res, baseCodeByte[0])
			} else {
				res = append(res, l.Front().Value.(byte))
				l.Remove(l.Front())
			}
		}
	}

	return res
}

// CodeToIDBase34 -- 将code逆向转换成原始id
func CodeToIDBase34(idByte []byte) (uint64, error) {
	baseCodeByte := []byte(baseCodeStr)
	baseMap := make(map[byte]int)
	for i, v := range baseCodeByte {
		baseMap[v] = i
	}

	if idByte == nil || len(idByte) == 0 {
		return 0, errors.New("param id nil or empyt")
	}

	var res uint64
	var r uint64

	for i := len(idByte) - 1; i >= 0; i-- {
		v, ok := baseMap[idByte[i]]
		if !ok {
			return 0, errors.New("param contain illegle character")
		}

		var b uint64 = 1
		for j := uint64(0); j < r; j++ {
			b *= 34
		}

		res += b * uint64(v)
		r++
	}

	return res, nil
}

// Reverse 字符串翻转
func Reverse(x int64) int64 {
	result := ""
	abs := int(math.Abs(float64(x)))
	str := strconv.Itoa(abs)
	for i := len(str) - 1; i >= 0; i-- {
		result = result + string(str[i])
	}

	if x < 0 {
		result = "-" + result
	}

	resultInt, err := strconv.Atoi(result)
	if err != nil || resultInt == 0 {
		return x
	}

	return int64(resultInt)
}

// GenCodeReverse 将id转换成n位长度的code
func GenCodeReverse(id int64) string {
	code := GenCodeBase34(uint64(Reverse(id)))
	return string(code)
}

// CodeToIDReverse -- 将code逆向转换成原始id
func CodeToIDReverse(code string) (int64, error) {
	if len(code) == 0 {
		return 0, fmt.Errorf("code is not empty")
	}
	id := []byte(code)
	baseCodeByte := []byte(baseCodeStr)
	baseMap := make(map[byte]int)
	for i, v := range baseCodeByte {
		baseMap[v] = i
	}

	if id == nil || len(id) == 0 {
		return 0, errors.New("param id nil or empyt")
	}

	var res int64
	var r int64

	for i := len(id) - 1; i >= 0; i-- {
		v, ok := baseMap[id[i]]
		if !ok {
			return 0, errors.New("param contain illegle character")
		}

		var b int64 = 1
		for j := int64(0); j < r; j++ {
			b *= 34
		}

		res += b * int64(v)
		r++
	}
	res = Reverse(res)
	return res, nil
}
