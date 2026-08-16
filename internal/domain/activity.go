package domain

import "time"

type Activity struct {
	ID          string         `json:"id"`
	VaultID string         `json:"vault_id"`
	ActorID     string         `json:"actor_id"`
	EntryID      string         `json:"entry_id,omitempty"`
	Kind        string         `json:"kind"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type AuditEvent struct {
	ID          string         `json:"id"`
	ActorID     string         `json:"actor_id"`
	VaultID string         `json:"vault_id,omitempty"`
	Action      string         `json:"action"`
	TargetType  string         `json:"target_type"`
	TargetID    string         `json:"target_id"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	RequestID   string         `json:"request_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Stats struct {
	Users         int            `json:"users"`
	Vaults    int            `json:"active_vaults"`
	Published     int            `json:"published_entries"`
	Comments      int            `json:"comments"`
	Recent30Days  map[string]int `json:"recent_30_days"`
	VisibleEntries  int            `json:"visible_entries"`
	PrivateHidden int            `json:"-"`
}
