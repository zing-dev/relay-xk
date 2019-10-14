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
