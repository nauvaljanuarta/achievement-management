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


type AlumniResponse struct {
	Data []User   `bson:"data" json:"data"`
	Meta MetaInfo `bson:"meta" json:"meta"`
}
