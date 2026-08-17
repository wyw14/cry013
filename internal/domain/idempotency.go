package domain

import "strings"

func IdempotencyScope(vaultID, actorID, operation, key string) string {
	return operation + ":" + vaultID + ":" + actorID + ":" + strings.TrimSpace(key)
}
