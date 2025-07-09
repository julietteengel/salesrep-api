package handlers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RouteAPI GetSalesReps godoc
// @Summary      List salesreps
// @Description  Get a list of all sales representatives
// @Tags         salesreps
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v1/salesreps [get]
func RouteAPI(g *echo.Group, db *gorm.DB) {
	g.GET("/salesreps", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "list salesreps"})
	})
}
