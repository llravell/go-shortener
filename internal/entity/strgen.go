package entity

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// RandomStringGenerator генерирует случайные строки в формате base64.
type RandomStringGenerator struct {
	hashLen int
}

const defaultHashLen = 10

var ErrGenerateStringFailed = errors.New("generate string has been failed")

func NewRandomStringGenerator() RandomStringGenerator {
	rsg := RandomStringGenerator{defaultHashLen}

	return rsg
}

func (rsg RandomStringGenerator) Generate() (string, error) {
	buf := make([]byte, rsg.hashLen)

	if _, err := rand.Read(buf); err != nil {
		return "", ErrGenerateStringFailed
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
