package user

import (
	"net/http"

	"github.com/julietteengel/salesrep-api/internal/common"
	"github.com/julietteengel/salesrep-api/internal/utils"
	"github.com/labstack/echo/v4"
)

type UserController struct {
	userService IUserService
}

func NewUserController(userService IUserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (c *UserController) GetType() common.ControllerType {
	return common.Private
}

func (c *UserController) Routes() []common.Route {
	return []common.Route{
		{
			Method:  "GET",
			Path:    "/user/me",
			Handler: c.GetCurrentUser,
		},
	}
}

func (c *UserController) GetCurrentUser(ctx echo.Context) (error, *utils.ControllerError) {
	// User is already synced by middleware, just return it
	currentUser := ctx.Get("user")
	if currentUser == nil {
		return nil, &utils.GenericNotFoundError
	}

	return ctx.JSON(http.StatusOK, currentUser), nil
}
