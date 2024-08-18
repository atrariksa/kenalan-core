package handler

import (
	"fmt"
	"net/http"

	"github.com/atrariksa/kenalan-core/app/repository"
	"github.com/atrariksa/kenalan-core/app/service"
	"github.com/atrariksa/kenalan-core/app/util"
	"github.com/atrariksa/kenalan-core/config"
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

	cfg := config.GetConfig()
	coreRepo := repository.NewCoreRepository()
	redisRepo := repository.NewRedisCoreRepository(util.GetRedisClient(cfg))
	svc := service.NewCoreService(coreRepo, redisRepo, cfg)
	RegisterCoreHandler(e, svc)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf("%v", cfg.ServerConfig.Host) + ":" + fmt.Sprintf("%v", cfg.ServerConfig.Port)))
}

func health(c echo.Context) error {
	return c.String(http.StatusOK, "Server Up")
}
