package httptransport

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry013/internal/application"
	appmw "github.com/wyw14/cry013/internal/middleware"
	"go.uber.org/zap"
)

func Router(h *Handler, tokens application.TokenManager, logger *zap.Logger, timeout time.Duration, origins []string) *gin.Engine {
	r := gin.New()
	r.Use(appmw.RequestID(), appmw.Recover(logger), appmw.Logger(logger), appmw.SecurityHeaders(), appmw.CORS(origins), appmw.Timeout(timeout))
	r.GET("/healthz", h.Health)
	r.GET("/readyz", h.Readiness)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", h.Register)
		v1.POST("/auth/login", h.Login)
		v1.POST("/auth/refresh", h.Refresh)
		v1.POST("/auth/logout", h.Logout)
	}
	secured := v1.Group("")
	secured.Use(appmw.Auth(tokens))
	{
		secured.POST("/vaults", h.CreateVault)
		secured.POST("/vaults/:vaultID/backups", h.CreateBackup)
		secured.POST("/backups/:backupID/accept", h.AcceptBackup)
		secured.POST("/vaults/:vaultID/transfer", h.RestoreVault)
		secured.POST("/vaults/:vaultID/entries", h.CreateEntry)
		secured.GET("/vaults/:vaultID/entries", h.SearchEntries)
		secured.GET("/vaults/:vaultID/activity", h.Activity)
		secured.GET("/vaults/:vaultID/stats", h.VaultStats)
		secured.PATCH("/entries/:entryID/status", h.TransitionEntry)
		secured.POST("/entries/:entryID/comments", h.Comment)
		secured.GET("/admin/stats", h.AdminStats)
	}
	return r
}
