package services

import (
	"bkc_microservice/services/sync-cbs-service/internal/domain/entities"
	"bkc_microservice/services/sync-cbs-service/internal/domain/repositories"
	"context"
	"fmt"
	"log"
	"time"
)

type SyncService struct {
	repo repositories.SycroneCoreRepository
}

func NewSyncService(repo repositories.SycroneCoreRepository) *SyncService {
	return &SyncService{repo: repo}
}

// InputCBSData - Admin input CBS data for pending user
func (s *SyncService) InputCBSData(ctx context.Context, userID string, req InputCBSRequest) (*entities.SycroneCore, error) {
	// Get existing mapping (must exist, created at user signup)
	sc, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user mapping not found: %w", err)
	}

	// Can't update if already completed
	if sc.SyncStatus == "completed" {
		return nil, fmt.Errorf("already synced")
	}

	// Validate input
	if req.UserCore == "" || req.KodeGroup1 == "" {
		return nil, fmt.Errorf("userCore and kodeGroup1 are required")
	}

	// Update with CBS data
	sc.UserCore = req.UserCore
	sc.KodeGroup1 = req.KodeGroup1
	sc.KodePerkiraan = req.KodePerkiraan
	if req.KodeCabang != nil {
		sc.KodeCabang = req.KodeCabang
	}

	// Mark as completed
	sc.SyncStatus = "completed"
	now := time.Now()
	sc.LastSyncAt = &now

	// Save
	if err := s.repo.Update(ctx, sc); err != nil {
		return nil, fmt.Errorf("failed to save: %w", err)
	}

	log.Printf("[SyncService] CBS data input for user %s → %s", userID, req.UserCore)
	return sc, nil
}

// GetMapping - Get CBS mapping for user
func (s *SyncService) GetMapping(ctx context.Context, userID string) (*entities.SycroneCore, error) {
	log.Printf("[SyncService] Get mapping for user %s", userID)
	return s.repo.GetByUserID(ctx, userID)
}

// ListPending - Get all pending mappings (admin dashboard)
func (s *SyncService) ListPending(ctx context.Context, page, size int) ([]*entities.SycroneCore, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	return s.repo.ListByStatus(ctx, "pending", page, size)
}

type InputCBSRequest struct {
	UserCore      string  `json:"userCore" validate:"required"`
	KodeGroup1    string  `json:"kodeGroup1" validate:"required"`
	KodePerkiraan string  `json:"kodePerkiraan" validate:"required"`
	KodeCabang    *string `json:"kodeCabang,omitempty"`
}
