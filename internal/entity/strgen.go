package entity

import (
	"math/rand"
	"time"
)

type RandomStringGenerator struct {
	r *rand.Rand
}

const hashLen = 10

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func NewRandomStringGenerator() RandomStringGenerator {
	s := rand.NewSource(time.Now().UnixNano())
	rsg := RandomStringGenerator{rand.New(s)}

	return rsg
}

func (rsg RandomStringGenerator) Generate() string {
	b := make([]byte, hashLen)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
