package auth

import (
	"crypto/sha512"
	"fmt"
)

func GetHash() {
	s := "1234567890"

	h := sha512.New()

	h.Write([]byte(s))

	bs := h.Sum(nil)

	fmt.Println(s)
	fmt.Printf("%x\n", bs)
}
