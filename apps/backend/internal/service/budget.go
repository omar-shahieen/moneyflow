package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/domain"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/domain/ports"
)

type BudgetService struct {
	repo         ports.BudgetRepository
	memberRepo   ports.BudgetMemberRepository
	categoryRepo ports.CategoryRepository
	planGuard    ports.PlanGuard
}

func NewBudgetService(repo ports.BudgetRepository, memberRepo ports.BudgetMemberRepository, categoryRepo ports.CategoryRepository, planGuard ports.PlanGuard) *BudgetService {
	return &BudgetService{
		repo:         repo,
		memberRepo:   memberRepo,
		categoryRepo: categoryRepo,
		planGuard:    planGuard,
	}
}

func (s *BudgetService) List(ctx context.Context, userID string) ([]model.BudgetWithMembers, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *BudgetService) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.BudgetWithMembers, error) {
	budget, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if _, err := s.memberRepo.Get(ctx, id, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}

	members, err := s.memberRepo.ListByBudget(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.BudgetWithMembers{
		Budget:  *budget,
		Members: members,
	}, nil
}

type CreateBudgetInput struct {
	CategoryID        uuid.UUID
	MonthlyLimitMinor int64
	Currency          string
}

func (s *BudgetService) Create(ctx context.Context, userID string, input CreateBudgetInput) (*model.BudgetWithMembers, error) {
	if input.MonthlyLimitMinor <= 0 {
		return nil, domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "monthly_limit_minor must be positive")
	}

	if s.planGuard != nil {
		if err := s.planGuard.CheckBudgetLimit(ctx, userID); err != nil {
			return nil, err
		}
	}

	if _, err := s.categoryRepo.GetByID(ctx, input.CategoryID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	budget := &model.Budget{
		ID:                uuid.New(),
		CategoryID:        input.CategoryID,
		MonthlyLimitMinor: input.MonthlyLimitMinor,
		Currency:          input.Currency,
	}

	if err := s.repo.Create(ctx, budget); err != nil {
		return nil, err
	}

	owner := &model.BudgetMember{
		BudgetID: budget.ID,
		UserID:   userID,
		Role:     model.BudgetMemberRoleOwner,
	}
	if err := s.memberRepo.Add(ctx, owner); err != nil {
		return nil, err
	}

	return &model.BudgetWithMembers{
		Budget:  *budget,
		Members: []model.BudgetMember{*owner},
	}, nil
}

type UpdateBudgetInput struct {
	CategoryID        uuid.UUID
	MonthlyLimitMinor int64
	Currency          string
}

func (s *BudgetService) Update(ctx context.Context, id uuid.UUID, userID string, input UpdateBudgetInput) (*model.Budget, error) {
	budget, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	member, err := s.memberRepo.Get(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}

	if member.Role != model.BudgetMemberRoleOwner {
		return nil, domain.ErrForbidden
	}

	budget.CategoryID = input.CategoryID
	budget.MonthlyLimitMinor = input.MonthlyLimitMinor
	budget.Currency = input.Currency

	if err := s.repo.Update(ctx, budget); err != nil {
		return nil, err
	}
	return budget, nil
}

func (s *BudgetService) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	member, err := s.memberRepo.Get(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrForbidden
		}
		return err
	}

	if member.Role != model.BudgetMemberRoleOwner {
		return domain.ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

type AddMemberInput struct {
	UserID string
}

func (s *BudgetService) AddMember(ctx context.Context, budgetID uuid.UUID, ownerID string, input AddMemberInput) error {
	owner, err := s.memberRepo.Get(ctx, budgetID, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrForbidden
		}
		return err
	}

	if owner.Role != model.BudgetMemberRoleOwner {
		return domain.ErrForbidden
	}

	member := &model.BudgetMember{
		BudgetID: budgetID,
		UserID:   input.UserID,
		Role:     model.BudgetMemberRoleMember,
	}

	return s.memberRepo.Add(ctx, member)
}

func (s *BudgetService) RemoveMember(ctx context.Context, budgetID uuid.UUID, ownerID string, targetUserID string) error {
	owner, err := s.memberRepo.Get(ctx, budgetID, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrForbidden
		}
		return err
	}

	if owner.Role != model.BudgetMemberRoleOwner {
		return domain.ErrForbidden
	}

	if ownerID == targetUserID {
		return domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "owner cannot remove themselves")
	}

	return s.memberRepo.Remove(ctx, budgetID, targetUserID)
}

func (s *BudgetService) GetUsage(ctx context.Context, budgetID uuid.UUID, month time.Time) (int64, error) {
	return s.repo.GetUsage(ctx, budgetID, month)
}
