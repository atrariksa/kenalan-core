package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atrariksa/kenalan-core/app/util"
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
	if strings.ToUpper(sur.Gender) != util.GenderMale && strings.ToUpper(sur.Gender) != util.GenderFemale {
		errMessage += fmt.Sprintf(errTemplate, "gender")
	}
	if sur.DOB == "" {
		errMessage += fmt.Sprintf(errTemplate, "dob")
	}
	if _, err := util.ToDateTimeYYYYMMDD(sur.DOB); err != nil {
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

type PurchaseRequest struct {
	Token       string
	UserID      int64  `json:"user_id"`
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`
	ExpiredAt   string `json:"expired_at"`
}

func (pr *PurchaseRequest) Validate() error {
	var errMessage string
	errTemplate := "%s is not valid;"
	if pr.UserID < 1 {
		errMessage += fmt.Sprintf(errTemplate, "user_id")
	}
	if pr.ProductCode == "" {
		errMessage += fmt.Sprintf(errTemplate, "product_code")
	}
	if pr.ProductCode != util.UnlimitedSwipeProductCode &&
		pr.ProductCode != util.AccountVerified {
		errMessage += fmt.Sprintf(errTemplate, "product_code")
	}
	if pr.ProductName == "" {
		errMessage += fmt.Sprintf(errTemplate, "product_name")
	}
	if pr.ExpiredAt == "" {
		errMessage += fmt.Sprintf(errTemplate, "expired_at")
	}
	if _, err := util.ToDateTimeYYYYMMDDTHHmmss(pr.ExpiredAt); err != nil {
		fmt.Println(err)
		errMessage += fmt.Sprintf(errTemplate, "expired_at")
	}
	if errMessage != "" {
		return errors.New(errMessage)
	}

	return nil
}
