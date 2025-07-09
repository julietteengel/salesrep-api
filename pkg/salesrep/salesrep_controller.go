package salesrep

import (
	"github.com/julietteengel/salesrep-api/internal/common"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
)

type SalesrepController struct {
	db *gorm.DB
}

func NewSalesrepController(db *gorm.DB) *SalesrepController {
	return &SalesrepController{db: db}
}

func (c *SalesrepController) GetType() common.ControllerType {
	return common.Private
}

func (c *SalesrepController) Routes() []common.Route {
	return []common.Route{
		{
			Method:  echo.GET,
			Path:    "/salesreps",
			Handler: c.GetSalesreps,
		},
	}
}

// GetSalesreps godoc
// @Summary      List salesreps
// @Description  Get a list of all sales representatives
// @Tags         salesreps
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /salesreps [get]
func (c *SalesrepController) GetSalesreps(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"message": "list salesreps"})
}
