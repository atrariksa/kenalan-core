package handler

import (
	"net/http"

	"github.com/atrariksa/kenalan-core/app/repository"
	"github.com/atrariksa/kenalan-core/app/service"
	"github.com/atrariksa/kenalan-core/app/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupServer() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/health", health)

	coreRepo := repository.NewCoreRepository()
	redisRepo := repository.NewRedisCoreRepository(util.GetRedisClient())
	svc := service.NewCoreService(coreRepo, redisRepo)
	RegisterCoreHandler(e, svc)

	// Start server
	e.Logger.Fatal(e.Start(":6020"))
}

func health(c echo.Context) error {
	return c.String(http.StatusOK, "Server Up")
}
