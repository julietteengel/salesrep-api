package salesrep

import (
	"github.com/julietteengel/salesrep-api/pkg/common"
)

type SalesRep struct {
	common.BaseModelV2
	FirstName string
	LastName  string
	Email     string `gorm:"uniqueIndex" validate:"required"`
}
