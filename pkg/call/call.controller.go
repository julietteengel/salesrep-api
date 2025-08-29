package call

import (
	"net/http"
	"strconv"

	"github.com/julietteengel/salesrep-api/internal/common"
	"github.com/julietteengel/salesrep-api/internal/utils"
	"github.com/julietteengel/salesrep-api/pkg/insights"
	"github.com/julietteengel/salesrep-api/pkg/transcription"
	"github.com/julietteengel/salesrep-api/pkg/user"
	"github.com/labstack/echo/v4"
)

type CallController struct {
	service         ICallService
	transcriptSvc   transcription.ITranscriptionService
	insightsSvc     insights.IInsightsService
}

func NewCallController(service ICallService, transcriptSvc transcription.ITranscriptionService, insightsSvc insights.IInsightsService) *CallController {
	return &CallController{
		service:       service,
		transcriptSvc: transcriptSvc,
		insightsSvc:   insightsSvc,
	}
}

func (c *CallController) GetType() common.ControllerType {
	return common.Private
}

func (c *CallController) Routes() []common.Route {
	return []common.Route{
		{
			Method:  "POST",
			Path:    "/calls/upload",
			Handler: c.UploadCall,
		},
		{
			Method:  "GET",
			Path:    "/calls",
			Handler: c.GetUserCalls,
		},
		{
			Method:  "GET",
			Path:    "/calls/:id",
			Handler: c.GetCallByID,
		},
		{
			Method:  "GET",
			Path:    "/calls/:id/details",
			Handler: c.GetCallDetails,
		},
	}
}

// UploadCall handles file upload for audio/video calls
// @Summary Upload a call recording
// @Tags Calls
// @ID UploadCall
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Audio/Video file"
// @Param title formData string false "Call title"
// @Param client_name formData string false "Client name"
// @Param client_company formData string false "Client company"
// @Success 200 {object} Call
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/calls/upload [post]
// @Security BearerAuth
func (c *CallController) UploadCall(ctx echo.Context) (error, *utils.ControllerError) {
	// Get current user from context
	currentUser := ctx.Get("current_user")
	if currentUser == nil {
		return nil, &utils.GenericAccessError
	}

	userEntity, ok := currentUser.(*user.User)
	if !ok {
		return nil, &utils.GenericAccessError
	}

	// Parse multipart form
	err := ctx.Request().ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		return err, &utils.GenericParamsError
	}

	// Get file from form
	file, header, err := ctx.Request().FormFile("file")
	if err != nil {
		return err, &utils.GenericParamsError
	}
	defer file.Close()

	// Get metadata from form
	metadata := &CallUploadMetadata{
		Title:         ctx.FormValue("title"),
		ClientName:    ctx.FormValue("client_name"),
		ClientCompany: ctx.FormValue("client_company"),
	}

	// For now, create a temporary conversation ID
	// In production, you'd either get this from the form or create a conversation first
	metadata.ConversationID = 1

	// Upload the call
	call, err := c.service.UploadCall(ctx.Request().Context(), userEntity.ID, file, header, metadata)
	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, call), nil
}

// GetUserCalls returns all calls for the authenticated user
// @Summary Get user's calls
// @Tags Calls
// @ID GetUserCalls
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {array} Call
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/calls [get]
// @Security BearerAuth
func (c *CallController) GetUserCalls(ctx echo.Context) (error, *utils.ControllerError) {
	// Get current user from context
	currentUser := ctx.Get("current_user")
	if currentUser == nil {
		return nil, &utils.GenericAccessError
	}

	userEntity, ok := currentUser.(*user.User)
	if !ok {
		return nil, &utils.GenericAccessError
	}

	// Parse pagination parameters
	page := 1
	pageSize := 10

	if p := ctx.QueryParam("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if ps := ctx.QueryParam("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	// Get calls
	calls, err := c.service.GetUserCalls(ctx.Request().Context(), userEntity.ID, page, pageSize)
	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, calls), nil
}

// GetCallByID returns a single call by ID
// @Summary Get call by ID
// @Tags Calls
// @ID GetCallByID
// @Produce json
// @Param id path int true "Call ID"
// @Success 200 {object} Call
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/calls/{id} [get]
// @Security BearerAuth
func (c *CallController) GetCallByID(ctx echo.Context) (error, *utils.ControllerError) {
	// Get current user from context
	currentUser := ctx.Get("current_user")
	if currentUser == nil {
		return nil, &utils.GenericAccessError
	}

	userEntity, ok := currentUser.(*user.User)
	if !ok {
		return nil, &utils.GenericAccessError
	}

	// Parse call ID
	callID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return err, &utils.GenericParamsError
	}

	// Get call
	call, err := c.service.GetCallByID(ctx.Request().Context(), uint(callID), userEntity.ID)
	if err != nil {
		if err.Error() == "unauthorized access to call" {
			return err, &utils.GenericAccessError
		}
		return err, &utils.GenericNotFoundError
	}

	return ctx.JSON(http.StatusOK, call), nil
}

// GetCallDetails returns call with transcripts and insights
// @Summary Get call details with transcripts and insights
// @Tags Calls
// @ID GetCallDetails
// @Produce json
// @Param id path int true "Call ID"
// @Success 200 {object} CallWithTranscript
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/calls/{id}/details [get]
// @Security BearerAuth
func (c *CallController) GetCallDetails(ctx echo.Context) (error, *utils.ControllerError) {
	// Get current user from context
	currentUser := ctx.Get("current_user")
	if currentUser == nil {
		return nil, &utils.GenericAccessError
	}

	userEntity, ok := currentUser.(*user.User)
	if !ok {
		return nil, &utils.GenericAccessError
	}

	// Parse call ID
	callID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return err, &utils.GenericParamsError
	}

	// Get call
	call, err := c.service.GetCallByID(ctx.Request().Context(), uint(callID), userEntity.ID)
	if err != nil {
		if err.Error() == "unauthorized access to call" {
			return err, &utils.GenericAccessError
		}
		return err, &utils.GenericNotFoundError
	}

	// Get transcripts
	transcriptionTranscripts, err := c.transcriptSvc.GetTranscriptionByCallID(ctx.Request().Context(), uint(callID))
	if err != nil {
		// Log error but continue without transcripts
		transcriptionTranscripts = []*transcription.CallTranscript{}
	}
	
	// Convert transcription.CallTranscript to call.CallTranscript
	transcripts := make([]*CallTranscript, len(transcriptionTranscripts))
	for i, tt := range transcriptionTranscripts {
		transcripts[i] = &CallTranscript{
			CallID:      tt.CallID,
			SpeakerID:   tt.SpeakerID,
			SpeakerName: tt.SpeakerName,
			Text:        tt.Text,
			StartTime:   tt.StartTime,
			EndTime:     tt.EndTime,
			Confidence:  tt.Confidence,
			Sentiment:   tt.Sentiment,
			KeywordTags: tt.KeywordTags,
			IsQuestion:  tt.IsQuestion,
			IsObjection: tt.IsObjection,
		}
		// Set BaseModelV2 fields
		transcripts[i].ID = tt.ID
		transcripts[i].CreatedAt = tt.CreatedAt
		transcripts[i].UpdatedAt = tt.UpdatedAt
	}

	// Generate insights if we have transcripts
	var callInsights interface{}
	if len(transcripts) > 0 {
		// Convert call transcripts to insights transcripts to avoid circular import
		insightsTranscripts := make([]*insights.CallTranscript, len(transcripts))
		for i, t := range transcripts {
			insightsTranscripts[i] = &insights.CallTranscript{
				ID:         t.ID,
				CallID:     t.CallID,
				SpeakerID:  t.SpeakerID,
				Text:       t.Text,
				StartTime:  t.StartTime,
				EndTime:    t.EndTime,
				Confidence: t.Confidence,
			}
		}
		callInsights = c.insightsSvc.GenerateInsights(insightsTranscripts)
	}

	// Return enhanced call details
	callDetails := &CallWithTranscript{
		Call:        call,
		Transcripts: transcripts,
		Insights:    callInsights,
	}

	return ctx.JSON(http.StatusOK, callDetails), nil
}