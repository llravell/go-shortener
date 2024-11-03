package entity

import "github.com/google/uuid"

type URL struct {
	UUID     string `json:"uuid"`
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

func NewURL(original, short string) *URL {
	id := uuid.New()

	return &URL{
		UUID:     id.String(),
		Short:    short,
		Original: original,
	}
}
