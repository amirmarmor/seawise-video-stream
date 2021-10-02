package web

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Web struct {
	Client *echo.Echo
}

func Create() *Web {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowMethods:     []string{"POST", "GET"},
		AllowCredentials: true,
	}))
	return &Web{
		Client: e,
	}
}

func (w *Web) Start() {
	w.Client.Logger.Fatal(w.Client.Start(":1323"))
}
