package models

type MetaInfo struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Total  int    `json:"total"`
	Pages  int    `json:"pages"`
	SortBy string `json:"sortBy"`
	Order  string `json:"order"`
	Search string `json:"search"`
}

type UserResponse struct {
	Data []User   `json:"data"` // ganti dari []User ke []UserProfileResponse
	Meta MetaInfo `json:"meta"`
}
