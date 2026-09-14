package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/recurring"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type RecurringRuleHandler struct {
	Handler
	recurringRuleService *service.RecurringRuleService
}

func NewRecurringRuleHandler(s *server.Server, recurringRuleService *service.RecurringRuleService) *RecurringRuleHandler {
	return &RecurringRuleHandler{
		Handler:              NewHandler(s),
		recurringRuleService: recurringRuleService,
	}
}

func (h *RecurringRuleHandler) List(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, query *recurring.ListRecurringRulesRequest) (*model.PaginatedResponse[recurring.RecurringRule], error) {
			userID := GetUserID(c)
			return h.recurringRuleService.GetRecurringRules(c.Request.Context(), userID, query)
		},
		http.StatusOK,
		&recurring.ListRecurringRulesRequest{},
	)(c)
}

func (h *RecurringRuleHandler) GetByID(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *recurring.GetRecurringRuleRequest) (*recurring.RecurringRule, error) {
			userID := GetUserID(c)
			return h.recurringRuleService.GetRecurringRuleByID(c.Request.Context(), userID, req.ID)
		},
		http.StatusOK,
		&recurring.GetRecurringRuleRequest{},
	)(c)
}

func (h *RecurringRuleHandler) Create(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *recurring.CreateRecurringRuleRequest) (*recurring.RecurringRule, error) {
			userID := GetUserID(c)
			return h.recurringRuleService.CreateRecurringRule(c.Request.Context(), userID, payload)
		},
		http.StatusCreated,
		&recurring.CreateRecurringRuleRequest{},
	)(c)
}

func (h *RecurringRuleHandler) Update(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *recurring.UpdateRecurringRuleRequest) (*recurring.RecurringRule, error) {
			userID := GetUserID(c)
			return h.recurringRuleService.UpdateRecurringRule(c.Request.Context(), userID, payload.ID, payload)
		},
		http.StatusOK,
		&recurring.UpdateRecurringRuleRequest{},
	)(c)
}

func (h *RecurringRuleHandler) Delete(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, payload *recurring.DeleteRecurringRuleRequest) error {
			userID := GetUserID(c)
			return h.recurringRuleService.DeleteRecurringRule(c.Request.Context(), userID, payload.ID)
		},
		http.StatusNoContent,
		&recurring.DeleteRecurringRuleRequest{},
	)(c)
}
