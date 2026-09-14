package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/category"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type CategoryService struct {
	server       *server.Server
	categoryRepo *repository.CategoryRepo
	userRepo     *repository.UserRepo
}

func NewCategoryService(server *server.Server, categoryRepo *repository.CategoryRepo, userRepo *repository.UserRepo) *CategoryService {
	return &CategoryService{
		server:       server,
		categoryRepo: categoryRepo,
		userRepo:     userRepo,
	}
}

func (s *CategoryService) GetCategories(ctx context.Context, userID string, req *category.ListCategoriesRequest) (*model.PaginatedResponse[category.Category], error) {
	req.Normalize()
	categories, err := s.categoryRepo.List(ctx, userID, req)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch categories")
		return nil, err
	}
	return categories, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error) {
	cat, err := s.categoryRepo.GetByID(ctx, categoryID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch category by ID")
		return nil, err
	}

	return cat, nil
}

func (s *CategoryService) CreateCategory(ctx context.Context, userID string, payload *category.CreateCategoryRequest) (*category.Category, error) {
	cat := &category.Category{
		ID:     uuid.New(),
		UserID: userID,
		Name:   payload.Name,
		Type:   category.CategoryType(payload.Type),
	}

	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to create category")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "category_created").
		Str("category_id", cat.ID.String()).
		Str("name", cat.Name).
		Msg("Category created successfully")

	return cat, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryRequest) (*category.Category, error) {
	cat, err := s.categoryRepo.GetByID(ctx, categoryID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch category for update")
		return nil, errs.NewNotFoundError("resource not found", false, nil)
	}

	cat.Name = payload.Name
	cat.Type = category.CategoryType(payload.Type)

	if err := s.categoryRepo.Update(ctx, cat); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to update category")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "category_updated").
		Str("category_id", cat.ID.String()).
		Str("name", cat.Name).
		Msg("Category updated successfully")

	return cat, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, userID string, categoryID uuid.UUID) error {
	if err := s.categoryRepo.Delete(ctx, categoryID, userID); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to delete category")
		return err
	}

	s.server.Logger.Info().
		Str("event", "category_deleted").
		Str("category_id", categoryID.String()).
		Msg("Category deleted successfully")

	return nil
}
