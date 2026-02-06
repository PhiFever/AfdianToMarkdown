package afdian

// Album 作品集信息
type Album struct {
	AlbumName string
	AlbumUrl  string
}

// Post 文章信息
type Post struct {
	Name     string
	Url      string
	Pictures []string
}

// Cookie 从 Chrome 中使用cookie master导出的 Cookies
type Cookie struct {
	Domain     string  `json:"domain"`
	Expiration float64 `json:"expirationDate"`
	HostOnly   bool    `json:"hostOnly"`
	HTTPOnly   bool    `json:"httpOnly"`
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	SameSite   string  `json:"sameSite"`
	Secure     bool    `json:"secure"`
	Session    bool    `json:"session"`
	StoreID    string  `json:"storeId"`
	Value      string  `json:"value"`
}
