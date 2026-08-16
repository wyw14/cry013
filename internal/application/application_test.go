package application_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wyw14/cry013/internal/application"
	"github.com/wyw14/cry013/internal/domain"
	"github.com/wyw14/cry013/internal/repository/memory"
	"github.com/wyw14/cry013/internal/service"
)

var testNow = time.Date(2026, 8, 16, 3, 0, 0, 0, time.UTC)

func seedIdentity(t *testing.T, repo *memory.Store, id, email string, role domain.PlatformRole) domain.User {
	t.Helper()
	user := domain.User{ID: id, Email: email, PasswordHash: "hash", DisplayName: id, Role: role, Status: domain.UserActive, CreatedAt: testNow, UpdatedAt: testNow}
	if err := repo.CreateUser(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	return user
}

func seedVault(t *testing.T, repo *memory.Store, id string, owner domain.User, members ...domain.User) domain.Vault {
	t.Helper()
	ws := domain.Vault{ID: id, Name: "studio", OwnerID: owner.ID, CreatedAt: testNow, UpdatedAt: testNow}
	if err := repo.CreateVault(context.Background(), ws); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertVaultLease(context.Background(), domain.VaultLease{VaultID: id, UserID: owner.ID, Role: domain.RoleOwner, Status: domain.VaultLeaseActive, JoinedAt: testNow}); err != nil {
		t.Fatal(err)
	}
	for _, user := range members {
		if err := repo.UpsertVaultLease(context.Background(), domain.VaultLease{VaultID: id, UserID: user.ID, Role: domain.RoleMember, Status: domain.VaultLeaseActive, JoinedAt: testNow}); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func TestVaultViewsRespectEntryVisibility(t *testing.T) {
	repo := memory.New()
	owner := seedIdentity(t, repo, "owner", "owner@example.test", domain.PlatformUser)
	member := seedIdentity(t, repo, "member", "member@example.test", domain.PlatformUser)
	ws := seedVault(t, repo, "vault", owner, member)
	entries := []domain.Entry{
		{ID: "public", VaultID: ws.ID, Title: "published public", Visibility: domain.VisibilityPublic, Status: domain.EntryPublished, CreatorID: owner.ID, CreatedAt: testNow, UpdatedAt: testNow},
		{ID: "vault", VaultID: ws.ID, Title: "team design", Visibility: domain.VisibilityVault, Status: domain.EntryPublished, CreatorID: owner.ID, CreatedAt: testNow, UpdatedAt: testNow},
		{ID: "private", VaultID: ws.ID, Title: "secret boss", Visibility: domain.VisibilityPrivate, Status: domain.EntryDraft, CreatorID: owner.ID, CreatedAt: testNow, UpdatedAt: testNow},
	}
	for _, entry := range entries {
		if err := repo.CreateEntry(context.Background(), entry); err != nil {
			t.Fatal(err)
		}
		if err := repo.AppendActivity(context.Background(), domain.Activity{ID: "a-" + entry.ID, VaultID: ws.ID, ActorID: owner.ID, EntryID: entry.ID, Kind: "entry.updated", CreatedAt: testNow}); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.CreateComment(context.Background(), domain.Comment{ID: "c-private", EntryID: "private", AuthorID: owner.ID, Body: "hidden", CreatedAt: testNow}); err != nil {
		t.Fatal(err)
	}

	discovery := application.NewDiscoveryService(repo)
	items, err := discovery.Search(context.Background(), member.ID, ws.ID, application.SearchFilter{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("search leaked or hid entries: got %d, want 2", len(items))
	}
	for _, entry := range items {
		if entry.ID == "private" {
			t.Fatal("private entry leaked through search")
		}
	}
	activity, err := discovery.RecentActivity(context.Background(), member.ID, ws.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(activity) != 2 {
		t.Fatalf("activity leaked private entry: got %d, want 2", len(activity))
	}
	stats, err := discovery.VaultStats(context.Background(), member.ID, ws.ID, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if stats.VisibleEntries != 2 || stats.Comments != 0 {
		t.Fatalf("stats leaked private data: %+v", stats)
	}
}

func TestRefreshOnlyOneConcurrentRotationSucceeds(t *testing.T) {
	repo := memory.New()
	clock := service.FixedClock{Time: testNow}
	tokenManager := service.JWTManager{Secret: []byte("0123456789abcdef0123456789abcdef"), AccessTTL: 15 * time.Minute, Clock: clock.Now}
	hasher := service.BcryptHasher{Cost: 4}
	auth := application.NewAuthService(repo, clock, service.UUIDGenerator{}, hasher, tokenManager, 24*time.Hour)
	if _, err := auth.Register(context.Background(), "player@example.test", "a-secure-password", "Player"); err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.Login(context.Background(), "player@example.test", "a-secure-password")
	if err != nil {
		t.Fatal(err)
	}

	const workers = 24
	start := make(chan struct{})
	var successes atomic.Int32
	var replays atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := auth.Refresh(context.Background(), tokens.RefreshToken)
			switch {
			case err == nil:
				successes.Add(1)
			case errors.Is(err, domain.ErrTokenReplayed):
				replays.Add(1)
			case errors.Is(err, domain.ErrTokenRevoked), errors.Is(err, domain.ErrNotFound):
				replays.Add(1)
			default:
				t.Errorf("unexpected refresh error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if got := successes.Load(); got != 1 {
		t.Fatalf("refresh token rotated %d times, want exactly 1", got)
	}
	if got := replays.Load(); got != workers-1 {
		t.Fatalf("replay count %d, want %d", got, workers-1)
	}
}

func TestRepeatedCreateReturnsSingleEntry(t *testing.T) {
	repo := memory.New()
	owner := seedIdentity(t, repo, "owner", "owner@example.test", domain.PlatformUser)
	ws := seedVault(t, repo, "vault", owner)
	serviceUnderTest := application.NewEntryService(repo, service.FixedClock{Time: testNow}, service.UUIDGenerator{})

	const workers = 20
	start := make(chan struct{})
	ids := make(chan string, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			entry, err := serviceUnderTest.Create(context.Background(), owner.ID, "client-request-77", application.CreateEntryInput{VaultID: ws.ID, Title: "Shared mechanic", Type: "mechanic", Visibility: domain.VisibilityVault}, application.RequestMeta{RequestID: "request"})
			if err != nil {
				t.Errorf("create entry: %v", err)
				return
			}
			ids <- entry.ID
		}()
	}
	close(start)
	wg.Wait()
	close(ids)
	var first string
	for id := range ids {
		if first == "" {
			first = id
		}
		if id != first {
			t.Fatalf("idempotent requests returned different IDs: %s and %s", first, id)
		}
	}
	entries, err := repo.ListEntries(context.Background(), ws.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("created %d entries, want 1", len(entries))
	}
}

func TestCanceledCreateBackupLeavesNoSideEffects(t *testing.T) {
	repo := memory.New()
	owner := seedIdentity(t, repo, "owner", "owner@example.test", domain.PlatformUser)
	ws := seedVault(t, repo, "vault", owner)
	repo.SetDelay(200 * time.Millisecond)
	vaultService := application.NewVaultService(repo, service.FixedClock{Time: testNow}, service.UUIDGenerator{})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := vaultService.CreateBackup(ctx, owner.ID, ws.ID, "new@example.test", domain.RoleMember, application.RequestMeta{RequestID: "cancelled"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("createBackup error = %v, want deadline exceeded", err)
	}
	createBackups, err := repo.ListBackups(context.Background(), ws.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(createBackups) != 0 {
		t.Fatalf("cancelled request created %d backups", len(createBackups))
	}
	audits, err := repo.ListAudits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(audits) != 0 {
		t.Fatalf("cancelled request wrote %d audit events", len(audits))
	}
}

func TestOwnershipTransferFailureKeepsOriginalOwner(t *testing.T) {
	repo := memory.New()
	owner := seedIdentity(t, repo, "owner", "owner@example.test", domain.PlatformUser)
	target := seedIdentity(t, repo, "target", "target@example.test", domain.PlatformUser)
	ws := seedVault(t, repo, "vault", owner, target)
	repo.SetOwnershipFailure(true)
	vaultService := application.NewVaultService(repo, service.FixedClock{Time: testNow}, service.UUIDGenerator{})
	err := vaultService.RestoreVault(context.Background(), owner.ID, ws.ID, target.ID, application.RequestMeta{RequestID: "transfer"})
	if err == nil {
		t.Fatal("expected simulated transfer failure")
	}
	after, _ := repo.VaultByID(context.Background(), ws.ID)
	oldVaultLease, _ := repo.VaultLease(context.Background(), ws.ID, owner.ID)
	newVaultLease, _ := repo.VaultLease(context.Background(), ws.ID, target.ID)
	if after.OwnerID != owner.ID || oldVaultLease.Role != domain.RoleOwner || newVaultLease.Role != domain.RoleMember {
		t.Fatalf("failed transfer changed state: vault=%+v old=%+v target=%+v", after, oldVaultLease, newVaultLease)
	}
}
