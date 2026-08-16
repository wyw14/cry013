package service

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct{ Cost int }

func (h BcryptHasher) Hash(password string) (string, error) {
	cost := h.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	value, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(value), err
}

func (BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
