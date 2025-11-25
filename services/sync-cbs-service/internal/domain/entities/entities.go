package entities

import "time"

type SycroneCore struct {
	ID            string     `json:"id"`
	UserID        string     `json:"userId"`
	UserCore      string     `json:"userCore"`      // CBS user ID
	KodeGroup1    string     `json:"kodeGroup1"`    // CBS group code
	KodePerkiraan string     `json:"kodePerkiraan"` // Account code
	KodeCabang    *string    `json:"kodeCabang,omitempty"`
	Status        string     `json:"status"`     // active, suspended, inactive
	SyncStatus    string     `json:"syncStatus"` // synced, pending, failed
	LastSyncAt    *time.Time `json:"lastSyncAt,omitempty"`
	SyncError     *string    `json:"syncError,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}
