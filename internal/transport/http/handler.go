package httptransport

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry013/internal/application"
	"github.com/wyw14/cry013/internal/domain"
	appmw "github.com/wyw14/cry013/internal/middleware"
)

type Handler struct {
	Auth          *application.AuthService
	Vaults    *application.VaultService
	Entries         *application.EntryService
	Discovery     *application.DiscoveryService
	Collaboration *application.CollaborationService
	Admin         *application.AdminService
	Ready         func() error
}

func (h *Handler) Health(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }
func (h *Handler) Readiness(c *gin.Context) {
	if h.Ready != nil {
		if err := h.Ready(); err != nil {
			c.JSON(http.StatusServiceUnavailable, appmw.ErrorBody(c, "not_ready", "database is unavailable", nil))
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) Register(c *gin.Context) {
	var in struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=10"`
		DisplayName string `json:"display_name" binding:"required,max=80"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	user, err := h.Auth.Register(c.Request.Context(), in.Email, in.Password, in.DisplayName)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) Login(c *gin.Context) {
	var in struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	tokens, err := h.Auth.Login(c.Request.Context(), in.Email, in.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) Refresh(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	tokens, err := h.Auth.Refresh(c.Request.Context(), in.RefreshToken)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) Logout(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	if err := h.Auth.Logout(c.Request.Context(), in.RefreshToken); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateVault(c *gin.Context) {
	var in struct {
		Name string `json:"name" binding:"required,max=120"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	ws, err := h.Vaults.Create(c.Request.Context(), appmw.ActorID(c), in.Name, appmw.Meta(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ws)
}

func (h *Handler) CreateBackup(c *gin.Context) {
	var in struct {
		Email string               `json:"email" binding:"required,email"`
		Role  domain.VaultRole `json:"role" binding:"required,oneof=admin member visitor"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	createBackup, err := h.Vaults.CreateBackup(c.Request.Context(), appmw.ActorID(c), c.Param("vaultID"), in.Email, in.Role, appmw.Meta(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, createBackup)
}

func (h *Handler) AcceptBackup(c *gin.Context) {
	if err := h.Vaults.AcceptBackup(c.Request.Context(), appmw.ActorID(c), c.Param("backupID")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) RestoreVault(c *gin.Context) {
	var in struct {
		TargetUserID string `json:"target_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	if err := h.Vaults.RestoreVault(c.Request.Context(), appmw.ActorID(c), c.Param("vaultID"), in.TargetUserID, appmw.Meta(c)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateEntry(c *gin.Context) {
	var in struct {
		Title      string                `json:"title" binding:"required,max=200"`
		Body       string                `json:"body" binding:"max=20000"`
		Type       string                `json:"type" binding:"required"`
		Visibility domain.EntryVisibility `json:"visibility" binding:"required,oneof=public vault private"`
		Priority   int                   `json:"priority" binding:"min=0,max=5"`
		Tags       []string              `json:"tags"`
		AssigneeID string                `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	entry, err := h.Entries.Create(c.Request.Context(), appmw.ActorID(c), c.GetHeader("Idempotency-Key"), application.CreateEntryInput{VaultID: c.Param("vaultID"), Title: in.Title, Body: in.Body, Type: in.Type, Visibility: in.Visibility, Priority: in.Priority, Tags: in.Tags, AssigneeID: in.AssigneeID}, appmw.Meta(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *Handler) TransitionEntry(c *gin.Context) {
	var in struct {
		Status domain.EntryStatus `json:"status" binding:"required,oneof=draft published archived"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	entry, err := h.Entries.Transition(c.Request.Context(), appmw.ActorID(c), c.Param("entryID"), in.Status, appmw.Meta(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *Handler) SearchEntries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, err := h.Discovery.Search(c.Request.Context(), appmw.ActorID(c), c.Param("vaultID"), application.SearchFilter{Query: c.Query("q"), AuthorID: c.Query("author_id"), Status: domain.EntryStatus(c.Query("status")), Page: page, PageSize: size, Sort: c.Query("sort")})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": size})
}

func (h *Handler) Activity(c *gin.Context) {
	items, err := h.Discovery.RecentActivity(c.Request.Context(), appmw.ActorID(c), c.Param("vaultID"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) Comment(c *gin.Context) {
	var in struct {
		Body     string   `json:"body" binding:"required,max=5000"`
		Mentions []string `json:"mentions"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, err)
		return
	}
	comment, err := h.Collaboration.Comment(c.Request.Context(), appmw.ActorID(c), c.Param("entryID"), in.Body, in.Mentions, appmw.Meta(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *Handler) VaultStats(c *gin.Context) {
	stats, err := h.Discovery.VaultStats(c.Request.Context(), appmw.ActorID(c), c.Param("vaultID"), time.Now().UTC())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) AdminStats(c *gin.Context) {
	stats, err := h.Admin.PlatformStats(c.Request.Context(), appmw.ActorID(c), time.Now().UTC())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func writeError(c *gin.Context, err error) {
	status, code, message := http.StatusBadRequest, "invalid_request", err.Error()
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "not_found", "resource not found"
	case errors.Is(err, domain.ErrForbidden):
		status, code, message = http.StatusForbidden, "forbidden", "operation is not allowed"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrTokenReplayed):
		status, code, message = http.StatusConflict, "conflict", "resource state conflicts with the request"
	}
	c.JSON(status, appmw.ErrorBody(c, code, message, nil))
}
