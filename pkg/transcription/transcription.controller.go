package transcription

import (
	"net/http"
	"strconv"

	"github.com/julietteengel/salesrep-api/internal/common"
	"github.com/julietteengel/salesrep-api/internal/utils"
	"github.com/labstack/echo/v4"
)

type TranscriptionController struct {
	service ITranscriptionService
}

func NewTranscriptionController(service ITranscriptionService) *TranscriptionController {
	return &TranscriptionController{
		service: service,
	}
}

func (c *TranscriptionController) GetType() common.ControllerType {
	return common.Private
}

func (c *TranscriptionController) Routes() []common.Route {
	return []common.Route{
		{
			Method:  "POST",
			Path:    "/calls/:id/transcribe",
			Handler: c.TranscribeCall,
		},
		{
			Method:  "GET",
			Path:    "/calls/:id/transcript",
			Handler: c.GetTranscription,
		},
	}
}

// TranscribeCall manually triggers transcription for a call
// @Summary Transcribe a call using OpenAI Whisper
// @Tags Transcription
// @ID TranscribeCall
// @Produce json
// @Param id path int true "Call ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/calls/{id}/transcribe [post]
// @Security BearerAuth
func (c *TranscriptionController) TranscribeCall(ctx echo.Context) (error, *utils.ControllerError) {
	// Parse call ID
	callID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return err, &utils.GenericParamsError
	}

	// For now, we'll need to get the audio file URL from somewhere
	// In a real implementation, you'd get this from the call record
	audioFileURL := ctx.QueryParam("audio_url")
	if audioFileURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "audio_url query parameter required"), &utils.GenericParamsError
	}

	// Start transcription
	err = c.service.TranscribeCall(ctx.Request().Context(), uint(callID), audioFileURL)
	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Transcription started successfully",
		"call_id": ctx.Param("id"),
	}), nil
}

// GetTranscription returns the transcription for a call
// @Summary Get call transcription
// @Tags Transcription
// @ID GetTranscription
// @Produce json
// @Param id path int true "Call ID"
// @Success 200 {array} call.CallTranscript
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/calls/{id}/transcript [get]
// @Security BearerAuth
func (c *TranscriptionController) GetTranscription(ctx echo.Context) (error, *utils.ControllerError) {
	// Parse call ID
	callID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return err, &utils.GenericParamsError
	}

	// Get transcription
	transcripts, err := c.service.GetTranscriptionByCallID(ctx.Request().Context(), uint(callID))
	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, transcripts), nil
}