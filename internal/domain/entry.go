package domain

import "time"

type EntryVisibility string

const (
	VisibilityPublic    EntryVisibility = "public"
	VisibilityVault EntryVisibility = "vault"
	VisibilityPrivate   EntryVisibility = "private"
)

type EntryStatus string

const (
	EntryDraft     EntryStatus = "draft"
	EntryPublished EntryStatus = "published"
	EntryArchived  EntryStatus = "archived"
)

type Entry struct {
	ID          string         `json:"id"`
	VaultID string         `json:"vault_id"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	Type        string         `json:"type"`
	Visibility  EntryVisibility `json:"visibility"`
	Status      EntryStatus     `json:"status"`
	Priority    int            `json:"priority"`
	Tags        []string       `json:"tags"`
	Attachments []Attachment   `json:"attachments"`
	CreatorID   string         `json:"creator_id"`
	AssigneeID  string         `json:"assignee_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty"`
}

type Attachment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type Comment struct {
	ID        string    `json:"id"`
	EntryID    string    `json:"entry_id"`
	AuthorID  string    `json:"author_id"`
	Body      string    `json:"body"`
	Mentions  []string  `json:"mentions"`
	CreatedAt time.Time `json:"created_at"`
}

func (c Entry) CanTransition(to EntryStatus) bool {
	switch c.Status {
	case EntryDraft:
		return to == EntryPublished || to == EntryArchived
	case EntryPublished:
		return to == EntryArchived || to == EntryDraft
	case EntryArchived:
		return to == EntryDraft
	default:
		return false
	}
}

func CanViewEntry(actor User, entry Entry, member *VaultLease) bool {
	if actor.Status != UserActive || entry.DeletedAt != nil {
		return false
	}
	if actor.Role == PlatformAdmin {
		return true
	}
	if entry.CreatorID == actor.ID || entry.AssigneeID == actor.ID {
		return true
	}
	switch entry.Visibility {
	case VisibilityPublic:
		return entry.Status == EntryPublished
	case VisibilityVault:
		return member != nil && member.Status == VaultLeaseActive
	case VisibilityPrivate:
		return false
	default:
		return false
	}
}

func CanEditEntry(actor User, entry Entry, member *VaultLease) bool {
	if actor.Role == PlatformAdmin || actor.ID == entry.CreatorID {
		return true
	}
	return member != nil && member.Status == VaultLeaseActive && (member.Role == RoleOwner || member.Role == RoleAdmin)
}
