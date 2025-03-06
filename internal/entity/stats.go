package entity

// Stats содержит статистику сервиса.
type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
