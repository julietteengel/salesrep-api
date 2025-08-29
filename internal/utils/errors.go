package utils

import (
	common "github.com/julietteengel/salesrep-api/pkg/common"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
	"net/http"
)

type ControllerError struct {
	Name          string
	HttpErrorCode int
	Translation   common.BaseTranslationMandatory
}

func WrapErrorHTTP(c echo.Context, err error, newError ControllerError) error {
	return echo.NewHTTPError(newError.HttpErrorCode, GetSafeMandatoryTranslatedString(newError.Translation, c.Get("locale").(string)))
}

func GetSafeMandatoryTranslatedString(base common.BaseTranslationMandatory, locale string) string {
	switch locale {
	case language.French.String():
		return base.Fr
	case language.English.String():
		return base.En
	default:
		return base.Fr
	}
}

var (
	GenericAccessError = ControllerError{
		Name:          "GenericAccessError",
		HttpErrorCode: http.StatusForbidden,
		Translation: common.BaseTranslationMandatory{
			Fr: "Acces interdit",
			En: "Access denied",
		},
	}
	GenericPaginationError = ControllerError{
		Name:          "GenericPaginationError",
		HttpErrorCode: http.StatusBadRequest,
		Translation: common.BaseTranslationMandatory{
			Fr: "Impossible de récupérer les paramètres de pagination.",
			En: "Cannot get pagination params.",
		},
	}
	GenericParamsError = ControllerError{
		Name:          "GenericParamsError",
		HttpErrorCode: http.StatusBadRequest,
		Translation: common.BaseTranslationMandatory{
			Fr: "Impossible de récupérer les paramètres.",
			En: "Cannot get params.",
		},
	}
	GenericServerError = ControllerError{
		Name:          "GenericServerError",
		HttpErrorCode: http.StatusInternalServerError,
		Translation: common.BaseTranslationMandatory{
			Fr: "Erreur interne du serveur.",
			En: "Internal server error.",
		},
	}
	GenericNotFoundError = ControllerError{
		Name:          "GenericNotFoundError",
		HttpErrorCode: http.StatusNotFound,
		Translation: common.BaseTranslationMandatory{
			Fr: "Ressource non trouvée.",
			En: "Resource not found.",
		},
	}
)
