package controllers

import (
	"github.com/julietteengel/salesrep-api/pkg/salesrep"
	"go.uber.org/fx"
)

var PrivateControllers = fx.Options(
	fx.Provide(
		AsController(salesrep.NewSalesrepController),
	),
)
