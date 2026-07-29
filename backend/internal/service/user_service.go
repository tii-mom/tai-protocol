package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tii-mom/tai-protocol/backend/ent"
	entUser "github.com/tii-mom/tai-protocol/backend/ent/user"
)

// UserService handles user creation and lookup.
type UserService struct {
	db *ent.Client
}

func NewUserService(db *ent.Client) *UserService {
	return &UserService{db: db}
}

// User is the API-facing user representation.
type User struct {
	ID           string    `json:"id"`
	TGUserID     int64     `json:"tg_user_id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	Wallet       string    `json:"wallet,omitempty"`
	TAIBalance   float64   `json:"tai_balance"`
	USDTBalance  float64   `json:"usdt_balance"`
	ReferralCode string    `json:"referral_code"`
	InviteCount  int       `json:"invite_count"`
	TotalEarned  float64   `json:"total_earned"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// FindOrCreateByTG finds an existing user by Telegram ID, or creates a new one.
func (s *UserService) FindOrCreateByTG(ctx context.Context, tgUserID int64, username, firstName string) (*User, bool, error) {
	existing, err := s.db.User.Query().
		Where(entUser.TgUserID(tgUserID)).
		Only(ctx)

	if err == nil && existing != nil {
		if username != "" && existing.TgUsername != username {
			existing, _ = s.db.User.UpdateOneID(existing.ID).
				SetTgUsername(username).
				Save(ctx)
		}
		return s.toAPIUser(existing), false, nil
	}

	if !ent.IsNotFound(err) {
		return nil, false, fmt.Errorf("query user: %w", err)
	}

	newUser, err := s.db.User.Create().
		SetTgUserID(tgUserID).
		SetTgUsername(username).
		SetFirstName(firstName).
		SetBalanceTai(0).
		SetBalanceUsdt(0).
		SetRole("user").
		SetReferralCode(generateReferralCode()).
		SetStatus("active").
		Save(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("create user: %w", err)
	}

	return s.toAPIUser(newUser), true, nil
}

// GetByID returns a user by internal UUID.
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	u, err := s.db.User.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return s.toAPIUser(u), nil
}

// AddTAI adds TAI to a user's balance.
func (s *UserService) AddTAI(ctx context.Context, userID string, amount float64) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}
	_, err = s.db.User.UpdateOneID(uid).
		AddBalanceTai(amount).
		AddTotalEarned(amount).
		Save(ctx)
	return err
}

// DeductTAI deducts TAI from a user's balance. Fails if insufficient.
func (s *UserService) DeductTAI(ctx context.Context, userID string, amount float64) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	u, err := tx.User.Get(ctx, uid)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if u.BalanceTai < amount {
		return fmt.Errorf("insufficient TAI: have %.2f, need %.2f", u.BalanceTai, amount)
	}

	_, err = tx.User.UpdateOneID(uid).AddBalanceTai(-amount).Save(ctx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *UserService) toAPIUser(u *ent.User) *User {
	return &User{
		ID:           u.ID.String(),
		TGUserID:     u.TgUserID,
		Username:     u.TgUsername,
		FirstName:    u.FirstName,
		Wallet:       u.WalletAddress,
		TAIBalance:   u.BalanceTai,
		USDTBalance:  u.BalanceUsdt,
		ReferralCode: u.ReferralCode,
		InviteCount:  u.InviteCount,
		TotalEarned:  u.TotalEarned,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
	}
}

func generateReferralCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
