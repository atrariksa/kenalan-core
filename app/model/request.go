package model

import (
	"errors"
	"fmt"
)

type SignUpRequest struct {
	Fullname string `json:"full_name"`
	Gender   string `json:"gender"`
	DOB      string `json:"dob"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (sur *SignUpRequest) Validate() error {
	var errMessage string
	errTemplate := "%s is not valid;"
	if sur.Fullname == "" {
		errMessage += fmt.Sprintf(errTemplate, "full_name")
	}
	if sur.Gender == "" {
		errMessage += fmt.Sprintf(errTemplate, "gender")
	}
	if sur.DOB == "" {
		errMessage += fmt.Sprintf(errTemplate, "dob")
	}
	if sur.Email == "" {
		errMessage += fmt.Sprintf(errTemplate, "email")
	}
	if sur.Password == "" {
		errMessage += fmt.Sprintf(errTemplate, "password")
	}
	if errMessage != "" {
		return errors.New(errMessage)
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (lr *LoginRequest) Validate() error {
	var errMessage string
	errTemplate := "%s is not valid;"
	if lr.Email == "" {
		errMessage += fmt.Sprintf(errTemplate, "email")
	}
	if lr.Password == "" {
		errMessage += fmt.Sprintf(errTemplate, "password")
	}
	if errMessage != "" {
		return errors.New(errMessage)
	}
	return nil
}

type ViewProfileRequest struct {
	Token                  string
	SwipeLeft              bool  `json:"swipe_left"`
	SwipeRight             bool  `json:"swipe_right"`
	CurrentViewedProfileID int64 `json:"current_viewed_profile_id"`
}

func (vpr *ViewProfileRequest) Validate() error {
	var errMessage string
	errTemplate := "%s is not valid;"
	if vpr.SwipeLeft && vpr.SwipeRight {
		errMessage += "swipe_left && swipe_right cannot have both true;"
	}

	if !vpr.SwipeLeft && !vpr.SwipeRight {
		errMessage += "swipe_left && swipe_right cannot have both false;"
	}

	if vpr.SwipeRight && vpr.CurrentViewedProfileID == 0 {
		errMessage += fmt.Sprintf(errTemplate, "current_viewed_profile_id;")
	}

	if errMessage != "" {
		return errors.New(errMessage)
	}

	return nil
}
