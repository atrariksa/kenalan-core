package model

type ViewProfile struct {
	Email            string  `json:"email"`
	IsUnlimitedSwipe bool    `json:"is_unlimited_swipe"`
	ViewedProfileIDs []int64 `json:"viewed_profile_ids"`
	SwipeCount       int64   `json:"swipe_count"`
}
