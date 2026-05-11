package structs

type User struct {
	ID     int    `json:"id"`
	Email  string `json:"email"`
	Password string `json:"password"`
	ApiKey string `json:"api_key"`
}
