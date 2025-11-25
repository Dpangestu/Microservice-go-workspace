package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bkc_microservice/services/sync-cbs-service/internal/application/services"
	"bkc_microservice/services/sync-cbs-service/internal/shared"

	"github.com/gorilla/mux"
)

type SyncCBSHandlers struct {
	syncService *services.SyncService
	logger      *shared.Logger
}

func NewSyncCBSHandlers(syncService *services.SyncService, logger *shared.Logger) *SyncCBSHandlers {
	return &SyncCBSHandlers{
		syncService: syncService,
		logger:      logger,
	}
}

// InputCBSData - POST /sync/users/{userID}/input-cbs-data
func (h *SyncCBSHandlers) InputCBSData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["userID"]

	h.logger.Info("InputCBSData request received", map[string]interface{}{
		"userID": userID,
	})

	var req services.InputCBSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate request
	if err := shared.ValidateStruct(req); err != nil {
		h.logger.Warn("validation error", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Input CBS data
	sc, err := h.syncService.InputCBSData(ctx, userID, req)
	if err != nil {
		h.logger.Error("failed to input CBS data", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("CBS data input success", map[string]interface{}{
		"userID":   userID,
		"userCore": sc.UserCore,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "CBS data saved successfully",
		"data":    sc,
	})
}

// GetMapping - GET /sync/users/{userID}/mapping
func (h *SyncCBSHandlers) GetMapping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["userID"]

	h.logger.Info("GetMapping request received", map[string]interface{}{
		"userID": userID,
	})

	if userID == "" {
		h.logger.Error("userID is empty", map[string]interface{}{
			"path": r.URL.Path,
		})
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	sc, err := h.syncService.GetMapping(ctx, userID)
	if err != nil {
		h.logger.Warn("mapping not found", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		http.Error(w, "mapping not found", http.StatusNotFound)
		return
	}

	h.logger.Info("mapping found", map[string]interface{}{
		"userID":   userID,
		"userCore": sc.UserCore,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   sc,
	})
}

// ListPending - GET /sync/mappings/pending?page=1&size=10
func (h *SyncCBSHandlers) ListPending(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	h.logger.Info("ListPending request received", map[string]interface{}{
		"page": page,
		"size": size,
	})

	items, total, err := h.syncService.ListPending(ctx, page, size)
	if err != nil {
		h.logger.Error("failed to list pending", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "failed to list pending mappings", http.StatusInternalServerError)
		return
	}

	h.logger.Info("list pending success", map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"count": len(items),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   items,
		"total":  total,
		"page":   page,
		"size":   size,
	})
}

// HealthCheck - GET /healthz
func (h *SyncCBSHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "sync-cbs-service",
	})
}
