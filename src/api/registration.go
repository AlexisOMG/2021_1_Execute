package api

import (
	"net/http"

	"github.com/labstack/echo"
)

func registration(c echo.Context) error {
	db := c.(*Database)
	input := new(UserRegistrationRequest)
	if err := c.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	newUser, err, code := db.CreateUser(input)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}
	err = SetSession(c, newUser.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, CreateLoginResponse(newUser))
}
