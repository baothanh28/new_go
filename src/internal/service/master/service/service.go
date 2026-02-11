package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"myapp/internal/service/master/model"
	"myapp/internal/service/master/repository"
)

var (
	// ErrMasterNotFound is returned when master record is not found
	ErrMasterNotFound = errors.New("master record not found")
	// ErrCodeExists is returned when code already exists
	ErrCodeExists = errors.New("master record with this code already exists")
)

// Service handles master business logic
type Service struct {
	repo *repository.Repository
}

// NewService creates a new master service
func NewService(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateMaster creates a new master record
func (s *Service) CreateMaster(ctx context.Context, req *model.CreateMasterRequest) (*model.Master, error) {
	// Check if code already exists
	exists, err := s.repo.CodeExists(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check code existence: %w", err)
	}
	if exists {
		return nil, ErrCodeExists
	}

	master := &model.Master{
		Name:        req.Name,
		Description: req.Description,
		Code:        req.Code,
		Type:        req.Type,
		IsActive:    true,
	}

	if err := s.repo.Insert(ctx, master); err != nil {
		return nil, fmt.Errorf("create master: %w", err)
	}

	return master, nil
}

// GetMasterByID retrieves a master record by ID
func (s *Service) GetMasterByID(ctx context.Context, id uint) (*model.Master, error) {
	master, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMasterNotFound
		}
		return nil, fmt.Errorf("get master by ID: %w", err)
	}
	return master, nil
}

// GetMasterByCode retrieves a master record by code
func (s *Service) GetMasterByCode(ctx context.Context, code string) (*model.Master, error) {
	master, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMasterNotFound
		}
		return nil, fmt.Errorf("get master by code: %w", err)
	}
	return master, nil
}

// GetAllMasters retrieves all master records with pagination
func (s *Service) GetAllMasters(ctx context.Context, limit, offset int) ([]*model.Master, error) {
	masters, err := s.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get all masters: %w", err)
	}
	return masters, nil
}

// GetActiveMasters retrieves all active master records with pagination
func (s *Service) GetActiveMasters(ctx context.Context, limit, offset int) ([]*model.Master, error) {
	masters, err := s.repo.GetActiveMasters(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get active masters: %w", err)
	}
	return masters, nil
}

// GetMastersByType retrieves master records by type
func (s *Service) GetMastersByType(ctx context.Context, masterType string, limit, offset int) ([]*model.Master, error) {
	masters, err := s.repo.GetByType(ctx, masterType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get masters by type: %w", err)
	}
	return masters, nil
}

// SearchMasters searches master records by name or description
func (s *Service) SearchMasters(ctx context.Context, query string, limit, offset int) ([]*model.Master, error) {
	masters, err := s.repo.SearchMasters(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search masters: %w", err)
	}
	return masters, nil
}

// UpdateMaster updates a master record
func (s *Service) UpdateMaster(ctx context.Context, id uint, req *model.UpdateMasterRequest) (*model.Master, error) {
	master, err := s.GetMasterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		master.Name = *req.Name
	}
	if req.Description != nil {
		master.Description = *req.Description
	}
	if req.Code != nil {
		// Check if new code already exists (if changed)
		if *req.Code != master.Code {
			exists, err := s.repo.CodeExists(ctx, *req.Code)
			if err != nil {
				return nil, fmt.Errorf("check code existence: %w", err)
			}
			if exists {
				return nil, ErrCodeExists
			}
		}
		master.Code = *req.Code
	}
	if req.Type != nil {
		master.Type = *req.Type
	}
	if req.IsActive != nil {
		master.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateByID(ctx, id, master); err != nil {
		return nil, fmt.Errorf("update master: %w", err)
	}

	return master, nil
}

// DeleteMaster deletes a master record
func (s *Service) DeleteMaster(ctx context.Context, id uint) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMasterNotFound
		}
		return fmt.Errorf("delete master: %w", err)
	}
	return nil
}
