package memory

import (
	"context"
	"errors"

	"github.com/wyw14/cry013/internal/domain"
)

func (s *Store) CreateVault(ctx context.Context, ws domain.Vault) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.vaults[ws.ID]; exists {
		return domain.ErrConflict
	}
	s.vaults[ws.ID] = ws
	return nil
}

func (s *Store) UpdateVault(ctx context.Context, ws domain.Vault) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.vaults[ws.ID]; !exists {
		return domain.ErrNotFound
	}
	s.vaults[ws.ID] = ws
	return nil
}

func (s *Store) VaultByID(ctx context.Context, id string) (domain.Vault, error) {
	if err := checkContext(ctx); err != nil {
		return domain.Vault{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	ws, ok := s.vaults[id]
	if !ok {
		return domain.Vault{}, domain.ErrNotFound
	}
	return ws, nil
}

func (s *Store) ListVaults(ctx context.Context) ([]domain.Vault, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Vault, 0, len(s.vaults))
	for _, ws := range s.vaults {
		items = append(items, ws)
	}
	return items, nil
}

func (s *Store) UpsertVaultLease(ctx context.Context, member domain.VaultLease) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.members[member.VaultID] == nil {
		s.members[member.VaultID] = map[string]domain.VaultLease{}
	}
	s.members[member.VaultID][member.UserID] = member
	return nil
}

func (s *Store) VaultLease(ctx context.Context, vaultID, userID string) (domain.VaultLease, error) {
	if err := checkContext(ctx); err != nil {
		return domain.VaultLease{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, ok := s.members[vaultID][userID]
	if !ok {
		return domain.VaultLease{}, domain.ErrNotFound
	}
	return member, nil
}

func (s *Store) ListVaultLeases(ctx context.Context, vaultID string) ([]domain.VaultLease, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.VaultLease, 0, len(s.members[vaultID]))
	for _, member := range s.members[vaultID] {
		items = append(items, member)
	}
	return items, nil
}

func (s *Store) RestoreVault(ctx context.Context, vaultID, fromID, toID string) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.faultMu.RLock()
	fail := s.failOwnershipChange
	s.faultMu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	ws, ok := s.vaults[vaultID]
	if !ok {
		return domain.ErrNotFound
	}
	from, fromOK := s.members[vaultID][fromID]
	to, toOK := s.members[vaultID][toID]
	if !fromOK || !toOK || ws.OwnerID != fromID {
		return domain.ErrInvalidTransition
	}
	if fail {
		return errors.New("simulated transaction failure")
	}
	ws.OwnerID = toID
	from.Role = domain.RoleAdmin
	to.Role = domain.RoleOwner
	s.vaults[vaultID] = ws
	s.members[vaultID][fromID] = from
	s.members[vaultID][toID] = to
	return nil
}

func (s *Store) CreateBackup(ctx context.Context, createBackup domain.Backup) error {
	if err := s.wait(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.createBackups[createBackup.ID]; exists {
		return domain.ErrConflict
	}
	s.createBackups[createBackup.ID] = createBackup
	return nil
}

func (s *Store) BackupByID(ctx context.Context, id string) (domain.Backup, error) {
	if err := checkContext(ctx); err != nil {
		return domain.Backup{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	createBackup, ok := s.createBackups[id]
	if !ok {
		return domain.Backup{}, domain.ErrNotFound
	}
	return createBackup, nil
}

func (s *Store) UpdateBackup(ctx context.Context, createBackup domain.Backup) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.createBackups[createBackup.ID]; !exists {
		return domain.ErrNotFound
	}
	s.createBackups[createBackup.ID] = createBackup
	return nil
}

func (s *Store) ListBackups(ctx context.Context, vaultID string) ([]domain.Backup, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := []domain.Backup{}
	for _, createBackup := range s.createBackups {
		if createBackup.VaultID == vaultID {
			items = append(items, createBackup)
		}
	}
	return items, nil
}
