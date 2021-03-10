package api

// CreateResponse -
type CreateResponse struct {
	ID int `json:"id"`
}

// Response -
type Response struct {
	Message string `json:"message"`
}

// RandomResponse -
type RandomResponse struct {
	Result  string `json:"result"`
	English string `json:"english"`
	Level   string `json:"level"`
}
