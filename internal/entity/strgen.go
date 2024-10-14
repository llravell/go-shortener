package entity

import (
	"math/rand"
	"time"
)

type randomStringGenerator struct {
	r *rand.Rand
}

const hashLen = 10

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func NewRandomStringGenerator() randomStringGenerator {
	s := rand.NewSource(time.Now().UnixNano())
	rsg := randomStringGenerator{rand.New(s)}

	return rsg
}

func (rsg randomStringGenerator) Generate(url string) string {
	b := make([]byte, hashLen)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
