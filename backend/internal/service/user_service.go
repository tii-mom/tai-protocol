package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// UserService handles user creation and lookup.
type UserService struct {
	// TODO: inject *ent.Client when DB is connected
}

func NewUserService() *UserService {
	return &UserService{}
}

// User represents a TAI Protocol user.
type User struct {
	ID         int64
	TGUserID   int64
	Username   string
	FirstName  string
	Wallet     string
	TAIBalance float64
	USDTBalance float64
	ReferralCode string
	ReferredBy int64
	CreatedAt  time.Time
}

// FindOrCreateByTG finds an existing user by Telegram ID, or creates a new one.
func (s *UserService) FindOrCreateByTG(ctx context.Context, tgUserID int64, username, firstName string) (*User, bool, error) {
	// TODO: Replace with actual Ent query:
	//   user, err := s.db.User.Query().Where(user.TgUserID(tgUserID)).Only(ctx)
	//   if ent.IsNotFound(err) → create new user
	//   return user, isNew, nil

	// Placeholder: simulate find-or-create
	_ = ctx
	user := &User{
		ID:           tgUserID % 100000, // placeholder
		TGUserID:     tgUserID,
		Username:     username,
		FirstName:    firstName,
		TAIBalance:   0,
		USDTBalance:  0,
		ReferralCode: generateReferralCode(),
		CreatedAt:    time.Now(),
	}
	return user, true, nil // true = isNew
}

// GetByID returns a user by internal ID.
func (s *UserService) GetByID(ctx context.Context, id int64) (*User, error) {
	// TODO: Ent query
	return nil, fmt.Errorf("not implemented")
}

// AddTAI adds TAI to a user's balance (off-chain ledger).
func (s *UserService) AddTAI(ctx context.Context, userID int64, amount float64, reason string) error {
	// TODO: Transaction: UPDATE users SET tai_balance = tai_balance + amount WHERE id = ?
	// TODO: Insert ledger entry (user_id, amount, reason, timestamp)
	return nil
}

// DeductTAI deducts TAI from a user's balance. Returns error if insufficient.
func (s *UserService) DeductTAI(ctx context.Context, userID int64, amount float64, reason string) error {
	// TODO: Transaction: UPDATE users SET tai_balance = tai_balance - amount WHERE id = ? AND tai_balance >= amount
	// TODO: Check rows affected == 1, else return "insufficient balance"
	// TODO: Insert ledger entry
	return nil
}

func generateReferralCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
