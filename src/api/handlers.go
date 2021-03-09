package api

import (
	"github.com/labstack/echo"
)

func Router(e *echo.Echo) {
	e.POST("/api/users/", registration)
	e.POST("/api/login/", login)
	e.DELETE("/api/logout/", logout)
	e.GET("/api/users/", GetCurrentUser)
	e.GET("/api/users/:id", GetUserByID)
	e.PATCH("/api/users/:id", PatchUserByID)
	e.DELETE("/api/users/:id", DeleteUserByID)
}
