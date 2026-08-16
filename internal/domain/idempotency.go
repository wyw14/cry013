package domain

import "strings"

func IdempotencyScope(vaultID, actorID, operation, key string) string {
	parts := []string{vaultID, actorID, operation, key}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, ":")
}
