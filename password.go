package kuu

import (
	"bytes"
	uuid "github.com/satori/go.uuid"
	"math/big"
	"strconv"
)

var keyMap = map[int]string{
	0: "0123456789",
	1: "abcdefghijklmnopqrstuvwxyz",
	2: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
}

// GenPassword
func GenPassword(passwordSize ...int) string {
	var (
		size   int
		hexStr = MD5(uuid.NewV4().String())
	)
	if len(passwordSize) > 0 {
		size = passwordSize[0]
	}
	if size > len(hexStr) {
		size = len(hexStr)
	}
	if size == 0 {
		size = 6
	}
	hexStr = hexStr[:size-1]
	var numList []int
	bi := big.NewInt(0)
	for _, num := range hexStr {
		bi.SetString(string(num), 16)
		v, _ := strconv.Atoi(bi.String())
		numList = append(numList, v)
	}

	var buf bytes.Buffer
	lastIdx2 := 0
	for index, num := range numList {
		if index < 3 {
			idx1 := index
			idx2 := num % len(keyMap[idx1])
			buf.WriteString(keyMap[idx1][idx2 : idx2+1])
		} else {
			idx1 := index % len(keyMap)
			idx2 := (num + lastIdx2) % len(keyMap[idx1])
			lastIdx2 = idx2
			buf.WriteString(keyMap[idx1][idx2 : idx2+1])
		}
	}
	return buf.String()
}
