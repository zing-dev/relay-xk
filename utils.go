package relay

import (
	"bytes"
	"fmt"
)

func ByteBinary(b byte) []byte {
	return bytes.Map(func(r rune) rune {
		return r - 48
	}, []byte(fmt.Sprintf("%08b", b)))
}

func BinaryByte(b []byte) byte {
	sum := byte(0)
	l := len(b) - 1
	for k, v := range b {
		sum += v << byte(l-k)
	}
	return sum
}
