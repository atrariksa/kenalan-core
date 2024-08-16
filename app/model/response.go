package model

type SignUpResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type LoginResponse struct {
	Code  string `json:"code"`
	Token string `json:"token"`
}

type ViewProfileResponse struct {
	Code       string `json:"code"`
	ID         int64  `json:"id"`
	IsVerified bool   `json:"is_verified"`
	Fullname   string `json:"full_name"`
	PhotoURL   string `json:"photo_url"`
}
