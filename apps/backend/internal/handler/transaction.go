package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type TransactionHandler struct {
	Handler
	transactionService *service.TransactionService
}

func NewTransactionHandler(s *server.Server, transactionService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		Handler:            NewHandler(s),
		transactionService: transactionService,
	}
}

func (h *TransactionHandler) List(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *transaction.ListTransactionsRequest) (*model.PaginatedResponse[transaction.Transaction], error) {
			userID := GetUserID(c)
			return h.transactionService.GetTransactions(c.Request.Context(), userID, req)
		},
		http.StatusOK,
		&transaction.ListTransactionsRequest{},
	)(c)
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *transaction.GetTransactionRequest) (*transaction.Transaction, error) {
			userID := GetUserID(c)
			return h.transactionService.GetTransactionByID(c.Request.Context(), userID, req.ID)
		},
		http.StatusOK,
		&transaction.GetTransactionRequest{},
	)(c)
}

func (h *TransactionHandler) Create(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *transaction.CreateTransactionRequest) (*transaction.Transaction, error) {
			userID := GetUserID(c)
			return h.transactionService.CreateTransaction(c.Request.Context(), userID, payload)
		},
		http.StatusCreated,
		&transaction.CreateTransactionRequest{},
	)(c)
}

func (h *TransactionHandler) Update(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *transaction.UpdateTransactionRequest) (*transaction.Transaction, error) {
			userID := GetUserID(c)
			return h.transactionService.UpdateTransaction(c.Request.Context(), userID, payload.ID, payload)
		},
		http.StatusOK,
		&transaction.UpdateTransactionRequest{},
	)(c)
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, payload *transaction.DeleteTransactionRequest) error {
			userID := GetUserID(c)
			return h.transactionService.DeleteTransaction(c.Request.Context(), userID, payload.ID)
		},
		http.StatusNoContent,
		&transaction.DeleteTransactionRequest{},
	)(c)
}

func (h *TransactionHandler) Summary(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *transaction.SummaryRequest) (*transaction.TransactionSummary, error) {
			userID := GetUserID(c)
			return h.transactionService.Summary(c.Request.Context(), userID, req.Month)
		},
		http.StatusOK,
		&transaction.SummaryRequest{},
	)(c)
}
