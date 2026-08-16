package memory

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry013/internal/domain"
)

type idemResult struct {
	done  chan struct{}
	value string
	err   error
}

type Store struct {
	mu          sync.RWMutex
	users       map[string]domain.User
	userByEmail map[string]string
	vaults  map[string]domain.Vault
	members     map[string]map[string]domain.VaultLease
	createBackups     map[string]domain.Backup
	entries       map[string]domain.Entry
	comments    map[string][]domain.Comment
	activities  map[string][]domain.Activity
	audits      []domain.AuditEvent
	tokens      map[string]domain.RefreshToken

	idemMu sync.Mutex
	idem   map[string]*idemResult

	faultMu             sync.RWMutex
	delay               time.Duration
	failOwnershipChange bool
}

func New() *Store {
	return &Store{
		users:       map[string]domain.User{},
		userByEmail: map[string]string{},
		vaults:  map[string]domain.Vault{},
		members:     map[string]map[string]domain.VaultLease{},
		createBackups:     map[string]domain.Backup{},
		entries:       map[string]domain.Entry{},
		comments:    map[string][]domain.Comment{},
		activities:  map[string][]domain.Activity{},
		tokens:      map[string]domain.RefreshToken{},
		idem:        map[string]*idemResult{},
	}
}

func (s *Store) SetDelay(d time.Duration) {
	s.faultMu.Lock()
	s.delay = d
	s.faultMu.Unlock()
}

func (s *Store) SetOwnershipFailure(enabled bool) {
	s.faultMu.Lock()
	s.failOwnershipChange = enabled
	s.faultMu.Unlock()
}

func (s *Store) wait(ctx context.Context) error {
	s.faultMu.RLock()
	delay := s.delay
	s.faultMu.RUnlock()
	if delay <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func cloneEntry(entry domain.Entry) domain.Entry {
	entry.Tags = append([]string(nil), entry.Tags...)
	entry.Attachments = append([]domain.Attachment(nil), entry.Attachments...)
	return entry
}
