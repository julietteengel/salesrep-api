package conversation

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julietteengel/salesrep-api/internal/common"
	"github.com/julietteengel/salesrep-api/internal/utils"
	"github.com/julietteengel/salesrep-api/pkg/auth"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ConversationController struct {
	conversationService IConversationService
}

func NewConversationController(conversationService IConversationService) *ConversationController {
	return &ConversationController{
		conversationService: conversationService,
	}
}

func (c *ConversationController) GetType() common.ControllerType {
	return common.Private
}

func (c *ConversationController) Routes() []common.Route {
	return []common.Route{
		{
			Method:  echo.GET,
			Path:    "/conversations",
			Handler: c.GetConversations,
		},
		{
			Method:  echo.POST,
			Path:    "/conversations",
			Handler: c.CreateConversation,
		},
		{
			Method:  echo.GET,
			Path:    "/conversations/stats",
			Handler: c.GetConversationStats,
		},
		{
			Method:  echo.GET,
			Path:    "/conversations/dashboard-stats",
			Handler: c.GetConversationDashboardStats,
		},
		{
			Method:  echo.GET,
			Path:    "/conversations/recent",
			Handler: c.GetRecentConversations,
		},
		{
			Method:  echo.GET,
			Path:    "/conversations/:id",
			Handler: c.GetConversation,
		},
		{
			Method:  echo.PUT,
			Path:    "/conversations/:id",
			Handler: c.UpdateConversation,
		},
		{
			Method:  echo.DELETE,
			Path:    "/conversations/:id",
			Handler: c.DeleteConversation,
		},
	}
}

// Request/Response types
type CreateConversationRequest struct {
	Title            string   `json:"title" validate:"required"`
	Description      *string  `json:"description"`
	ClientName       *string  `json:"client_name"`
	ClientCompany    *string  `json:"client_company"`
	ClientEmail      *string  `json:"client_email"`
	SalesStage       *string  `json:"sales_stage"`
	DealValue        *float64 `json:"deal_value"`
	MeetingType      *string  `json:"meeting_type"`
	MeetingObjective *string  `json:"meeting_objective"`
}

type UpdateConversationRequest struct {
	Title            *string  `json:"title"`
	Description      *string  `json:"description"`
	Status           *string  `json:"status"`
	ClientName       *string  `json:"client_name"`
	ClientCompany    *string  `json:"client_company"`
	ClientEmail      *string  `json:"client_email"`
	SalesStage       *string  `json:"sales_stage"`
	DealValue        *float64 `json:"deal_value"`
	MeetingType      *string  `json:"meeting_type"`
	MeetingObjective *string  `json:"meeting_objective"`
}

type ConversationStatsResponse struct {
	TotalConversations int     `json:"total_conversations"`
	CompletedCount     int     `json:"completed_count"`
	ScheduledCount     int     `json:"scheduled_count"`
	CancelledCount     int     `json:"cancelled_count"`
	AverageDealValue   float64 `json:"average_deal_value"`
	TotalDealValue     float64 `json:"total_deal_value"`
}

type ConversationDashboardStatsResponse struct {
	TotalRevenue     float64 `json:"total_revenue"`
	MonthlyRevenue   float64 `json:"monthly_revenue"`
	RevenueTrend     string  `json:"revenue_trend"`   // percentage like "12%"
	TrendDirection   string  `json:"trend_direction"` // "up" or "down"
	AverageDealValue float64 `json:"average_deal_value"`
	ConversionsRate  float64 `json:"conversions_rate"` // percentage of completed deals
}

// GetConversations lists conversations with filtering and search
// @Summary      List conversations
// @Description  Get a list of conversations with optional filtering
// @Tags         Conversations
// @ID           GetConversations
// @Produce      json
// @Param        search query string false "Search in title, client name, or company"
// @Param        status query string false "Filter by status (scheduled, in_progress, completed, cancelled)"
// @Param        sales_stage query string false "Filter by sales stage"
// @Param        limit query int false "Number of conversations to return (default 20)"
// @Param        offset query int false "Number of conversations to skip (default 0)"
// @Success      200 {array} Conversation
// @Failure      400 {string} string "Bad request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal server error"
// @Router       /conversations [get]
// @Security     BearerAuth
func (c *ConversationController) GetConversations(ctx echo.Context) (error, *utils.ControllerError) {
	// Check permissions
	jwtUser := ctx.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*auth.CustomClaims)
	if err := auth.CheckClaims(claims.Scope, "read:conversation"); err != nil {
		return errors.New("user does not have permission to read conversations"), &utils.GenericAccessError
	}

	// Parse query parameters
	search := ctx.QueryParam("search")
	status := ctx.QueryParam("status")
	salesStage := ctx.QueryParam("sales_stage")

	limit := 20
	if l := ctx.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := ctx.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Build filters
	filters := ConversationFilters{
		Search:     search,
		Status:     status,
		SalesStage: salesStage,
		Limit:      limit,
		Offset:     offset,
	}

	// Get conversations from service
	conversations, err := c.conversationService.GetConversations(ctx.Request().Context(), filters)
	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, conversations), nil
}

// CreateConversation creates a new conversation
// @Summary      Create conversation
// @Description  Create a new conversation
// @Tags         Conversations
// @ID           CreateConversation
// @Accept       json
// @Produce      json
// @Param        conversation body CreateConversationRequest true "Conversation data"
// @Success      201 {object} Conversation
// @Failure      400 {string} string "Bad request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal server error"
// @Router       /conversations [post]
// @Security     BearerAuth
func (c *ConversationController) CreateConversation(ctx echo.Context) (error, *utils.ControllerError) {
	// Check permissions
	jwtUser := ctx.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*auth.CustomClaims)
	if err := auth.CheckClaims(claims.Scope, "write:conversation"); err != nil {
		return errors.New("user does not have permission to create conversations"), &utils.GenericAccessError
	}

	// Parse and validate request
	var req CreateConversationRequest
	if err := ctx.Bind(&req); err != nil {
		return err, &utils.GenericParamsError
	}

	if err := ctx.Validate(&req); err != nil {
		return err, &utils.GenericParamsError
	}

	// Get user ID from JWT
	//userID := uint(claims.Subject.(float64))

	// Create conversation via service
	//conversation, err := c.conversationService.CreateConversation(ctx.Request().Context(), req, userID)
	//if err != nil {
	//	return err, &utils.GenericServerError
	//}

	return ctx.JSON(http.StatusCreated, nil), nil
}

// GetConversation gets a specific conversation by ID
// @Summary      Get conversation
// @Description  Get a conversation by ID
// @Tags         Conversations
// @ID           GetConversation
// @Produce      json
// @Param        id path int true "Conversation ID"
// @Success      200 {object} Conversation
// @Failure      400 {string} string "Bad request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      404 {string} string "Conversation not found"
// @Failure      500 {string} string "Internal server error"
// @Router       /conversations/{id} [get]
// @Security     BearerAuth
func (c *ConversationController) GetConversation(ctx echo.Context) (error, *utils.ControllerError) {
	// Check permissions
	jwtUser := ctx.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*auth.CustomClaims)
	if err := auth.CheckClaims(claims.Scope, "read:conversation"); err != nil {
		return errors.New("user does not have permission to read conversations"), &utils.GenericAccessError
	}

	// Parse ID
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return errors.New("invalid conversation ID"), &utils.GenericParamsError
	}

	// Get conversation via service
	conversation, err := c.conversationService.GetConversationByID(ctx.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("conversation not found"), &utils.GenericNotFoundError
		}
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, conversation), nil
}

// UpdateConversation updates a conversation
// @Summary      Update conversation
// @Description  Update a conversation by ID
// @Tags         Conversations
// @ID           UpdateConversation
// @Accept       json
// @Produce      json
// @Param        id path int true "Conversation ID"
// @Param        conversation body UpdateConversationRequest true "Updated conversation data"
// @Success      200 {object} Conversation
// @Failure      400 {string} string "Bad request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      404 {string} string "Conversation not found"
// @Failure      500 {string} string "Internal server error"
// @Router       /conversations/{id} [put]
// @Security     BearerAuth
func (c *ConversationController) UpdateConversation(ctx echo.Context) (error, *utils.ControllerError) {
	// Check permissions
	jwtUser := ctx.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*auth.CustomClaims)
	if err := auth.CheckClaims(claims.Scope, "write:conversation"); err != nil {
		return errors.New("user does not have permission to update conversations"), &utils.GenericAccessError
	}

	// Parse ID
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return errors.New("invalid conversation ID"), &utils.GenericParamsError
	}

	// Parse and validate request
	var req UpdateConversationRequest
	if err := ctx.Bind(&req); err != nil {
		return err, &utils.GenericParamsError
	}

	// Update conversation via service
	conversation, err := c.conversationService.UpdateConversation(ctx.Request().Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("conversation not found"), &utils.GenericNotFoundError
		}
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, conversation), nil
}

// DeleteConversation deletes a conversation
// @Summary      Delete conversation
// @Description  Delete a conversation by ID
// @Tags         Conversations
// @ID           DeleteConversation
// @Param        id path int true "Conversation ID"
// @Success      204 "No Content"
// @Failure      400 {string} string "Bad request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      404 {string} string "Conversation not found"
// @Failure      500 {string} string "Internal server error"
// @Router       /conversations/{id} [delete]
// @Security     BearerAuth
func (c *ConversationController) DeleteConversation(ctx echo.Context) (error, *utils.ControllerError) {
	// Check permissions
	jwtUser := ctx.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*auth.CustomClaims)
	if err := auth.CheckClaims(claims.Scope, "delete:conversation"); err != nil {
		return errors.New("user does not have permission to delete conversations"), &utils.GenericAccessError
	}

	// Parse ID
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return errors.New("invalid conversation ID"), &utils.GenericParamsError
	}

	// Delete conversation via service
	if err := c.conversationService.DeleteConversation(ctx.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("conversation not found"), &utils.GenericNotFoundError
		}
		return err, &utils.GenericServerError
	}

	return ctx.NoContent(http.StatusNoContent), nil
}

// GetConversationDashboardStats gets conversation statistics specifically for dashboard metrics
// @Summary      Get conversation dashboard statistics
// @Description  Get conversation revenue and performance statistics for dashboard
// @Tags         Conversations
// @ID           GetConversationDashboardStats
// @Produce      json
// @Success      200 {object} ConversationDashboardStatsResponse
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal server error"
// @Router       /conversations/dashboard-stats [get]
// @Security     BearerAuth
func (c *ConversationController) GetConversationDashboardStats(ctx echo.Context) (error, *utils.ControllerError) {
	// Check permissions
	jwtUser := ctx.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*auth.CustomClaims)
	if err := auth.CheckClaims(claims.Scope, "read:conversation"); err != nil {
		return errors.New("user does not have permission to read conversations"), &utils.GenericAccessError
	}

	// Get dashboard stats via service
	stats, err := c.conversationService.GetConversationDashboardStats(ctx.Request().Context())
	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, stats), nil
}

// GetRecentConversations gets recent conversations for dashboard
// @Summary      Get recent conversations
// @Description  Get recent conversations for dashboard display
// @Tags         Conversations
// @ID           GetRecentConversations
// @Produce      json
// @Param        limit query int false "Number of conversations to return (default 5)"
// @Success      200 {array} Conversation
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal server error"
// @Router       /conversations/recent [get]
// @Security     BearerAuth
func (c *ConversationController) GetRecentConversations(ctx echo.Context) (error, *utils.ControllerError) {
	// Check permissions
	jwtUser := ctx.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*auth.CustomClaims)
	if err := auth.CheckClaims(claims.Scope, "read:conversation"); err != nil {
		return errors.New("user does not have permission to read conversations"), &utils.GenericAccessError
	}

	// Parse limit parameter
	limit := 5
	if l := ctx.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 20 {
			limit = parsed
		}
	}

	// Get recent conversations via service
	conversations, err := c.conversationService.GetRecentConversations(ctx.Request().Context(), limit)
	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, conversations), nil
}

// GetConversationStats gets conversation statistics for dashboard
// @Summary      Get conversation statistics
// @Description  Get conversation statistics for dashboard
// @Tags         Conversations
// @ID           GetConversationStats
// @Produce      json
// @Success      200 {object} ConversationStatsResponse
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal server error"
// @Router       /conversations/stats [get]
// @Security     BearerAuth
func (c *ConversationController) GetConversationStats(ctx echo.Context) (error, *utils.ControllerError) {
	// Check permissions
	jwtUser := ctx.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*auth.CustomClaims)
	if err := auth.CheckClaims(claims.Scope, "read:conversation"); err != nil {
		return errors.New("user does not have permission to read conversations"), &utils.GenericAccessError
	}

	// Get stats via service
	stats, err := c.conversationService.GetConversationStats(ctx.Request().Context())
	if err != nil {
		return err, &utils.GenericServerError
	}

	return ctx.JSON(http.StatusOK, stats), nil
}
