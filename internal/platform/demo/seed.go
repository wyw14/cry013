package demo

import (
	"context"
	"errors"
	"time"

	"github.com/wyw14/cry013/internal/application"
	"github.com/wyw14/cry013/internal/domain"
)

const DemoPassword = "Gameforge-Demo-2026!"

func Seed(ctx context.Context, repo application.Repository, passwords application.PasswordHasher, now time.Time) error {
	if _, err := repo.UserByEmail(ctx, "admin@cry013.local"); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	hash, err := passwords.Hash(DemoPassword)
	if err != nil {
		return err
	}
	users := []domain.User{
		{ID: "00000000-0000-4000-8000-000000000001", Email: "admin@cry013.local", PasswordHash: hash, DisplayName: "平台管理员", Role: domain.PlatformAdmin, Status: domain.UserActive, CreatedAt: now, UpdatedAt: now},
		{ID: "00000000-0000-4000-8000-000000000002", Email: "alice@cry013.local", PasswordHash: hash, DisplayName: "Alice", Role: domain.PlatformUser, Status: domain.UserActive, CreatedAt: now, UpdatedAt: now},
		{ID: "00000000-0000-4000-8000-000000000003", Email: "bob@cry013.local", PasswordHash: hash, DisplayName: "Bob", Role: domain.PlatformUser, Status: domain.UserActive, CreatedAt: now, UpdatedAt: now},
		{ID: "00000000-0000-4000-8000-000000000004", Email: "carol@cry013.local", PasswordHash: hash, DisplayName: "Carol", Role: domain.PlatformUser, Status: domain.UserActive, CreatedAt: now, UpdatedAt: now},
		{ID: "00000000-0000-4000-8000-000000000005", Email: "dave@cry013.local", PasswordHash: hash, DisplayName: "Dave", Role: domain.PlatformUser, Status: domain.UserActive, CreatedAt: now, UpdatedAt: now},
	}
	for _, user := range users {
		if err := repo.CreateUser(ctx, user); err != nil {
			return err
		}
	}
	vaults := []domain.Vault{
		{ID: "10000000-0000-4000-8000-000000000001", Name: "星港工作室", OwnerID: users[1].ID, CreatedAt: now, UpdatedAt: now},
		{ID: "10000000-0000-4000-8000-000000000002", Name: "像素远征队", OwnerID: users[3].ID, CreatedAt: now, UpdatedAt: now},
	}
	for _, ws := range vaults {
		if err := repo.CreateVault(ctx, ws); err != nil {
			return err
		}
	}
	members := []domain.VaultLease{
		{VaultID: vaults[0].ID, UserID: users[1].ID, Role: domain.RoleOwner, Status: domain.VaultLeaseActive, JoinedAt: now},
		{VaultID: vaults[0].ID, UserID: users[2].ID, Role: domain.RoleAdmin, Status: domain.VaultLeaseActive, JoinedAt: now},
		{VaultID: vaults[0].ID, UserID: users[4].ID, Role: domain.RoleMember, Status: domain.VaultLeaseActive, JoinedAt: now},
		{VaultID: vaults[1].ID, UserID: users[3].ID, Role: domain.RoleOwner, Status: domain.VaultLeaseActive, JoinedAt: now},
		{VaultID: vaults[1].ID, UserID: users[2].ID, Role: domain.RoleVisitor, Status: domain.VaultLeaseActive, JoinedAt: now},
	}
	for _, member := range members {
		if err := repo.UpsertVaultLease(ctx, member); err != nil {
			return err
		}
	}
	entries := []domain.Entry{
		{ID: "20000000-0000-4000-8000-000000000001", VaultID: vaults[0].ID, Title: "昼夜循环关卡", Body: "让潮汐改变潜行路线", Type: "mechanic", Visibility: domain.VisibilityPublic, Status: domain.EntryPublished, Priority: 4, Tags: []string{"关卡", "潜行"}, CreatorID: users[1].ID, AssigneeID: users[2].ID, CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "20000000-0000-4000-8000-000000000002", VaultID: vaults[0].ID, Title: "Boss 语音草稿", Body: "未公开配音方向", Type: "narrative", Visibility: domain.VisibilityPrivate, Status: domain.EntryDraft, Priority: 5, Tags: []string{"叙事", "保密"}, CreatorID: users[1].ID, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now.Add(-time.Hour)},
		{ID: "20000000-0000-4000-8000-000000000003", VaultID: vaults[0].ID, Title: "协作式制作台", Body: "多人拆解装备制作步骤", Type: "system", Visibility: domain.VisibilityVault, Status: domain.EntryPublished, Priority: 3, Tags: []string{"系统", "协作"}, CreatorID: users[2].ID, CreatedAt: now.Add(-12 * time.Hour), UpdatedAt: now.Add(-30 * time.Minute)},
		{ID: "20000000-0000-4000-8000-000000000004", VaultID: vaults[1].ID, Title: "低分辨率天气", Body: "像素雨雪过渡", Type: "art", Visibility: domain.VisibilityPublic, Status: domain.EntryPublished, Priority: 2, Tags: []string{"美术", "天气"}, CreatorID: users[3].ID, CreatedAt: now.Add(-7 * 24 * time.Hour), UpdatedAt: now.Add(-3 * time.Hour)},
	}
	for _, entry := range entries {
		if err := repo.CreateEntry(ctx, entry); err != nil {
			return err
		}
		_ = repo.AppendActivity(ctx, domain.Activity{ID: entry.ID + "-activity", VaultID: entry.VaultID, ActorID: entry.CreatorID, EntryID: entry.ID, Kind: "entry.created", CreatedAt: entry.CreatedAt})
	}
	comments := []domain.Comment{
		{ID: "30000000-0000-4000-8000-000000000001", EntryID: entries[0].ID, AuthorID: users[2].ID, Body: "可以让月相参与难度曲线", Mentions: []string{users[1].ID}, CreatedAt: now.Add(-time.Hour)},
		{ID: "30000000-0000-4000-8000-000000000002", EntryID: entries[2].ID, AuthorID: users[4].ID, Body: "我来补一版交互草图", CreatedAt: now.Add(-20 * time.Minute)},
	}
	for _, comment := range comments {
		if err := repo.CreateComment(ctx, comment); err != nil {
			return err
		}
	}
	return nil
}
