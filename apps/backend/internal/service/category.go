package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/domain"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/domain/ports"
)

type CategoryService struct {
	repo      ports.CategoryRepository
	userRepo  ports.UserRepository
	planGuard ports.PlanGuard
}

func NewCategoryService(repo ports.CategoryRepository, userRepo ports.UserRepository, planGuard ports.PlanGuard) *CategoryService {
	return &CategoryService{
		repo:      repo,
		userRepo:  userRepo,
		planGuard: planGuard,
	}
}

func (s *CategoryService) List(ctx context.Context, userID string) ([]model.Category, error) {
	return s.repo.List(ctx, userID)
}

func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Category, error) {
	cat, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return cat, nil
}

type CreateCategoryInput struct {
	Name string
	Type model.CategoryType
}

func (s *CategoryService) Create(ctx context.Context, userID string, input CreateCategoryInput) (*model.Category, error) {
	if input.Type != model.CategoryTypeIncome && input.Type != model.CategoryTypeExpense {
		return nil, domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "type must be 'income' or 'expense'")
	}

	if s.planGuard != nil {
		if err := s.planGuard.CheckCategoryLimit(ctx, userID); err != nil {
			return nil, err
		}
	}

	cat := &model.Category{
		ID:     uuid.New(),
		UserID: userID,
		Name:   input.Name,
		Type:   input.Type,
	}

	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

type UpdateCategoryInput struct {
	Name string
	Type model.CategoryType
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, userID string, input UpdateCategoryInput) (*model.Category, error) {
	cat, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if input.Type != model.CategoryTypeIncome && input.Type != model.CategoryTypeExpense {
		return nil, domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "type must be 'income' or 'expense'")
	}

	cat.Name = input.Name
	cat.Type = input.Type

	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}
