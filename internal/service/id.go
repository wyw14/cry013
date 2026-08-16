package service

import "github.com/google/uuid"

type UUIDGenerator struct{}

func (UUIDGenerator) New() string { return uuid.NewString() }

type SequenceGenerator struct {
	Prefix string
	Next   int
}

func (g *SequenceGenerator) New() string {
	g.Next++
	return g.Prefix + string(rune('a'+g.Next-1))
}
