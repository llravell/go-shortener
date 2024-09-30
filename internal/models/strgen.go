package models

import (
	"math/rand"
	"time"
)

type randomStringGenerator struct {
	r *rand.Rand
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func NewRandomStringGenerator() randomStringGenerator {
	s := rand.NewSource(time.Now().UnixNano())
	rsg := randomStringGenerator{rand.New(s)}

	return rsg
}

func (rsg randomStringGenerator) Generate(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
