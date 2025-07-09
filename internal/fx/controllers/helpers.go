package controllers

import (
	"github.com/julietteengel/salesrep-api/internal/common"
	"go.uber.org/fx"
)

// AsController is an FX constructor that annotates a struct as a Handler..
func AsController(controller any) any {
	return fx.Annotate(
		controller,
		fx.As(new(common.Controller)),
		fx.ResultTags(`group:"controllers"`),
	)
}
