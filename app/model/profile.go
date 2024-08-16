package model

type Profile struct {
	ID         int64  `json:"id"`
	Fullname   string `json:"full_name"`
	IsVerified bool   `json:"is_verified"`
	PhotoURL   string `json:"photo_url"`
}
