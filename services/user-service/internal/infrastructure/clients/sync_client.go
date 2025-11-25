package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SyncCBSClient struct {
	baseURL    string
	httpClient *http.Client
}

// Response struct dari sync-cbs-service
type SycroneCoreData struct {
	ID            string     `json:"id"`
	UserID        string     `json:"userId"`
	UserCore      string     `json:"userCore"`
	KodeGroup1    string     `json:"kodeGroup1"`
	KodePerkiraan string     `json:"kodePerkiraan"`
	KodeCabang    *string    `json:"kodeCabang,omitempty"`
	Status        string     `json:"status"`
	SyncStatus    string     `json:"syncStatus"`
	LastSyncAt    *time.Time `json:"lastSyncAt,omitempty"`
}

type GetMappingResponse struct {
	Status string          `json:"status"`
	Data   SycroneCoreData `json:"data"`
}

func NewSyncCBSClient(baseURL string) *SyncCBSClient {
	return &SyncCBSClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *SyncCBSClient) GetMapping(ctx context.Context, userID string) (*SycroneCoreData, error) {
	url := fmt.Sprintf("%s/sync/users/%s/mapping", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Mapping tidak ada (pending)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result GetMappingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Data, nil
}
