package entity

// URL содержит данные о сокращенном урле.
type URL struct {
	UUID     string `json:"uuid"`
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
	UserUUID string `json:"user_uuid"`
	Deleted  bool   `json:"is_deleted"`
}

// URLDeleteItem dto удаление урлов.
type URLDeleteItem struct {
	UserUUID string
	Hashes   []string
}
