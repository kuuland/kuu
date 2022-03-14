package kuu

import (
	"bytes"
	"math/rand"
)

var keyMap = "bcefghjkmpqrtvwxyBCEFGHJKMPQRTVWXY2346789"

// GenPassword Gen random password
func GenPassword(passwordSize ...int) string {
	var (
		size int
	)
	if len(passwordSize) > 0 {
		size = passwordSize[0]
	}
	if size == 0 {
		size = 6
	}

	var buf bytes.Buffer
	for {
		if buf.Len() >= size {
			break
		}
		i := rand.Intn(41)
		buf.Write([]byte{keyMap[i]})
	}
	return buf.String()
}
