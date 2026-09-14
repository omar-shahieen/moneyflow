package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/budget"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type BudgetHandler struct {
	Handler
	budgetService *service.BudgetService
}

func NewBudgetHandler(s *server.Server, budgetService *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{
		Handler:       NewHandler(s),
		budgetService: budgetService,
	}
}

func (h *BudgetHandler) List(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, query *budget.ListBudgetsRequest) (*model.PaginatedResponse[budget.BudgetResponse], error) {
			userID := GetUserID(c)
			return h.budgetService.GetBudgets(c.Request.Context(), userID, query)
		},
		http.StatusOK,
		&budget.ListBudgetsRequest{},
	)(c)
}

func (h *BudgetHandler) GetByID(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *budget.GetBudgetRequest) (*budget.BudgetResponse, error) {
			userID := GetUserID(c)
			return h.budgetService.GetBudgetByID(c.Request.Context(), userID, req.ID)
		},
		http.StatusOK,
		&budget.GetBudgetRequest{},
	)(c)
}

func (h *BudgetHandler) Create(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *budget.CreateBudgetRequest) (*budget.BudgetWithMembers, error) {
			userID := GetUserID(c)
			return h.budgetService.CreateBudget(c.Request.Context(), userID, payload)
		},
		http.StatusCreated,
		&budget.CreateBudgetRequest{},
	)(c)
}

func (h *BudgetHandler) Update(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *budget.UpdateBudgetRequest) (*budget.Budget, error) {
			userID := GetUserID(c)
			return h.budgetService.UpdateBudget(c.Request.Context(), userID, payload.ID, payload)
		},
		http.StatusOK,
		&budget.UpdateBudgetRequest{},
	)(c)
}

func (h *BudgetHandler) AddMember(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, req *budget.AddMemberRequest) error {
			userID := GetUserID(c)
			return h.budgetService.AddMember(c.Request.Context(), userID, req.ID, req.UserID)
		},
		http.StatusNoContent,
		&budget.AddMemberRequest{},
	)(c)
}

func (h *BudgetHandler) RemoveMember(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, req *budget.RemoveMemberRequest) error {
			userID := GetUserID(c)
			return h.budgetService.RemoveMember(c.Request.Context(), userID, req.ID, req.UserID)
		},
		http.StatusNoContent,
		&budget.RemoveMemberRequest{},
	)(c)
}
