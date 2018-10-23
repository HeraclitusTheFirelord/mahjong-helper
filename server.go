package main

import (
	"github.com/labstack/echo"
	"net/http"
	"time"
	"github.com/labstack/echo/middleware"
)

type echoHandler struct {
	analysing bool
}

func (h *echoHandler) index(c echo.Context) error {
	return c.String(http.StatusOK, time.Now().Format("2006-01-02 15:04:05"))
}

func (h *echoHandler) analysis(c echo.Context) error {
	if h.analysing {
		return c.NoContent(http.StatusForbidden)
	}

	h.analysing = true
	defer func() { h.analysing = false }()

	d := struct {
		Reset      bool   `json:"reset"`
		Tiles      string `json:"tiles"`
		ShowDetail bool   `json:"show_detail"`
	}{}
	if err := c.Bind(&d); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if d.Reset {
		reset()
	}

	if d.ShowDetail {
		detailFlag = true
		defer func() { detailFlag = false }()
	}
	if _, _, err := analysis(d.Tiles); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func runServer() {
	h := &echoHandler{}

	e := echo.New()
	e.Use(middleware.Recover())

	e.GET("/", h.index)
	e.POST("/analysis", h.analysis)

	if err := e.Start(":12121"); err != nil {
		_errorExit(err.Error())
	}
}