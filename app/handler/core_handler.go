package handler

import (
	"context"
	"net/http"

	"github.com/atrariksa/kenalan-core/app/model"
	"github.com/atrariksa/kenalan-core/app/service"
	"github.com/labstack/echo/v4"
)

// CoreHandler  represent the httphandler for core
type CoreHandler struct {
	CoreService service.ICoreService
}

// RegisterCoreHandler will initialize the cores/ resources endpoint
func RegisterCoreHandler(e *echo.Echo, svc service.ICoreService) {
	handler := &CoreHandler{
		CoreService: svc,
	}
	e.POST("v1/kenalan/sign_up", handler.SignUp)
	e.POST("v1/kenalan/login", handler.Login)
	e.POST("v1/kenalan/view_profile", handler.ViewProfile)
}

func (ch *CoreHandler) SignUp(c echo.Context) (err error) {
	var signUpRequest model.SignUpRequest
	err = c.Bind(&signUpRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = signUpRequest.Validate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = ch.CoreService.SignUp(context.Background(), signUpRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, model.SignUpResponse{
		Code:    "0000",
		Message: "Success",
	})
}

func (ch *CoreHandler) Login(c echo.Context) (err error) {
	var loginRequest model.LoginRequest
	err = c.Bind(&loginRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = loginRequest.Validate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	token, err := ch.CoreService.Login(context.Background(), loginRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, model.LoginResponse{
		Code:  "0000",
		Token: token,
	})
}

func (ch *CoreHandler) ViewProfile(c echo.Context) (err error) {
	var viewProfileRequest model.ViewProfileRequest
	err = c.Bind(&viewProfileRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = viewProfileRequest.Validate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	token := c.Request().Header.Get("Authorization")
	viewProfileRequest.Token = token

	profile, err := ch.CoreService.ViewProfile(context.Background(), viewProfileRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, model.ViewProfileResponse{
		Code:       "0000",
		ID:         profile.ID,
		Fullname:   profile.Fullname,
		IsVerified: profile.IsVerified,
		PhotoURL:   profile.PhotoURL,
	})
}
