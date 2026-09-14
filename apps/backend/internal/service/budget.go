package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/budget"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type BudgetService struct {
	server       *server.Server
	budgetRepo   *repository.BudgetRepo
	memberRepo   *repository.BudgetMemberRepo
	categoryRepo *repository.CategoryRepo
}

func NewBudgetService(server *server.Server, budgetRepo *repository.BudgetRepo, memberRepo *repository.BudgetMemberRepo, categoryRepo *repository.CategoryRepo) *BudgetService {
	return &BudgetService{
		server:       server,
		budgetRepo:   budgetRepo,
		memberRepo:   memberRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *BudgetService) GetBudgets(ctx context.Context, userID string, query *budget.ListBudgetsRequest) (*model.PaginatedResponse[budget.BudgetResponse], error) {
	result, err := s.budgetRepo.ListByUser(ctx, userID, query.ToListQuery())
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch budgets")
		return nil, err
	}

	responses := make([]budget.BudgetResponse, 0, len(result.Data))
	now := time.Now()
	for _, b := range result.Data {
		usage, _ := s.budgetRepo.GetUsage(ctx, b.ID, now)
		var usagePercent float64
		if b.MonthlyLimitMinor > 0 {
			usagePercent = float64(usage) / float64(b.MonthlyLimitMinor) * 100
		}
		responses = append(responses, budget.BudgetResponse{
			BudgetWithMembers: b,
			Usage:             usage,
			UsagePercent:      usagePercent,
			Exceeded:          usage > b.MonthlyLimitMinor,
		})
	}

	return &model.PaginatedResponse[budget.BudgetResponse]{
		Data:       responses,
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	}, nil
}

func (s *BudgetService) GetBudgetByID(ctx context.Context, userID string, budgetID uuid.UUID) (*budget.BudgetResponse, error) {
	b, err := s.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch budget by ID")
		return nil, errs.NewNotFoundError("resource not found", false, nil)
	}

	if _, err := s.memberRepo.Get(ctx, budgetID, userID); err != nil {
		s.server.Logger.Error().Err(err).Msg("user is not a member of this budget")
		return nil, errs.NewForbiddenError("forbidden", false)
	}

	members, err := s.memberRepo.ListByBudget(ctx, budgetID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch budget members")
		return nil, err
	}

	bw := budget.BudgetWithMembers{
		Budget:  *b,
		Members: members,
	}

	usage, _ := s.budgetRepo.GetUsage(ctx, budgetID, time.Now())
	var usagePercent float64
	if bw.MonthlyLimitMinor > 0 {
		usagePercent = float64(usage) / float64(bw.MonthlyLimitMinor) * 100
	}

	return &budget.BudgetResponse{
		BudgetWithMembers: bw,
		Usage:             usage,
		UsagePercent:      usagePercent,
		Exceeded:          usage > bw.MonthlyLimitMinor,
	}, nil
}

func (s *BudgetService) CreateBudget(ctx context.Context, userID string, payload *budget.CreateBudgetRequest) (*budget.BudgetWithMembers, error) {
	categoryID := payload.CategoryID

	if _, err := s.categoryRepo.GetByID(ctx, categoryID, userID); err != nil {
		s.server.Logger.Error().Err(err).Msg("category not found")
		return nil, errs.NewNotFoundError("resource not found", false, nil)
	}

	b := &budget.Budget{
		ID:                uuid.New(),
		CategoryID:        categoryID,
		MonthlyLimitMinor: payload.MonthlyLimitMinor,
		Currency:          payload.Currency,
	}

	if err := s.budgetRepo.Create(ctx, b); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to create budget")
		return nil, err
	}

	owner := &budget.BudgetMember{
		BudgetID: b.ID,
		UserID:   userID,
		Role:     budget.BudgetMemberRoleOwner,
	}
	if err := s.memberRepo.Add(ctx, owner); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to add budget owner")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "budget_created").
		Str("budget_id", b.ID.String()).
		Msg("Budget created successfully")

	return &budget.BudgetWithMembers{
		Budget:  *b,
		Members: []budget.BudgetMember{*owner},
	}, nil
}

func (s *BudgetService) UpdateBudget(ctx context.Context, userID string, budgetID uuid.UUID, payload *budget.UpdateBudgetRequest) (*budget.Budget, error) {
	b, err := s.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("budget not found")
		return nil, errs.NewNotFoundError("resource not found", false, nil)
	}

	member, err := s.memberRepo.Get(ctx, budgetID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("user is not a member of this budget")
		return nil, errs.NewForbiddenError("forbidden", false)
	}

	if member.Role != budget.BudgetMemberRoleOwner {
		return nil, errs.NewForbiddenError("forbidden", false)
	}

	b.CategoryID = payload.CategoryID
	b.MonthlyLimitMinor = payload.MonthlyLimitMinor
	b.Currency = payload.Currency

	if err := s.budgetRepo.Update(ctx, b); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to update budget")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "budget_updated").
		Str("budget_id", b.ID.String()).
		Msg("Budget updated successfully")

	return b, nil
}

func (s *BudgetService) DeleteBudget(ctx context.Context, userID string, budgetID uuid.UUID) error {
	member, err := s.memberRepo.Get(ctx, budgetID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("user is not a member of this budget")
		return errs.NewForbiddenError("forbidden", false)
	}

	if member.Role != budget.BudgetMemberRoleOwner {
		return errs.NewForbiddenError("forbidden", false)
	}

	if err := s.budgetRepo.Delete(ctx, budgetID); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to delete budget")
		return err
	}

	s.server.Logger.Info().
		Str("event", "budget_deleted").
		Str("budget_id", budgetID.String()).
		Msg("Budget deleted successfully")

	return nil
}

func (s *BudgetService) AddMember(ctx context.Context, userID string, budgetID uuid.UUID, targetUserID string) error {
	owner, err := s.memberRepo.Get(ctx, budgetID, userID)
	if err != nil {
		return errs.NewForbiddenError("forbidden", false)
	}

	if owner.Role != budget.BudgetMemberRoleOwner {
		return errs.NewForbiddenError("forbidden", false)
	}

	member := &budget.BudgetMember{
		BudgetID: budgetID,
		UserID:   targetUserID,
		Role:     budget.BudgetMemberRoleMember,
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to add budget member")
		return err
	}

	return nil
}

func (s *BudgetService) RemoveMember(ctx context.Context, userID string, budgetID uuid.UUID, targetUserID string) error {
	owner, err := s.memberRepo.Get(ctx, budgetID, userID)
	if err != nil {
		return errs.NewForbiddenError("forbidden", false)
	}

	if owner.Role != budget.BudgetMemberRoleOwner {
		return errs.NewForbiddenError("forbidden", false)
	}

	if userID == targetUserID {
		return errs.NewBadRequestError("owner cannot remove themselves", false, nil, nil, nil)
	}

	if err := s.memberRepo.Remove(ctx, budgetID, targetUserID); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to remove budget member")
		return err
	}

	return nil
}

func (s *BudgetService) GetUsage(ctx context.Context, budgetID uuid.UUID, month time.Time) (int64, error) {
	return s.budgetRepo.GetUsage(ctx, budgetID, month)
}
