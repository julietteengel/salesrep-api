package user

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang-jwt/jwt/v5"
	"github.com/julietteengel/salesrep-api/pkg/auth"
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
			Method:  "POST",
			Path:    "/auth/user/get-or-create",
			Handler: c.GetOrCreateUser,
		},
		{
			Method:  "GET",
			Path:    "/user/me",
			Handler: c.GetCurrentUser,
		},
	}
}

// GetOrCreateUser add a new user in the database if it doesn't already exist, and get it
// @Summary Add a new user in the database if it doesn't already exist, and get it
// @Tags User
// @ID GetOrCreateUser
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Status unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/user/get-or-create [post]
// @Security BearerAuth
func (c *UserController) GetOrCreateUser(ctx echo.Context) (error, *utils.ControllerError) {
	// Get Auth0 data from context (set by Auth0Middleware)
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(*auth.CustomClaims)
	spew.Dump(claims)
	spew.Dump(user)

	if !claims.EmailVerified {
		return errors.New("user email not verified"), &utils.GenericAccessError
	}
	// Create the user
	userEntity, err := c.userService.GetOrCreateUser(ctx.Request().Context(), claims.Subject, claims.Email)

	spew.Dump(userEntity)

	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, userEntity), nil
}

func (c *UserController) GetCurrentUser(ctx echo.Context) (error, *utils.ControllerError) {
	// User is already synced by middleware, just return it
	currentUser := ctx.Get("user")
	if currentUser == nil {
		return nil, &utils.GenericNotFoundError
	}

	return ctx.JSON(http.StatusOK, currentUser), nil
}
