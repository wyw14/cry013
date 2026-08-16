package domain

import "strings"

func IdempotencyScope(vaultID, actorID, operation, key string) string {
	return strings.TrimSpace(key)
}
