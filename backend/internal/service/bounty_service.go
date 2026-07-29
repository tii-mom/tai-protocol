package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tii-mom/tai-protocol/backend/ent"
	entBounty "github.com/tii-mom/tai-protocol/backend/ent/bounty"
)

// BountyService handles bounty task lifecycle.
type BountyService struct {
	db  *ent.Client
	pet *PetService
}

func NewBountyService(db *ent.Client, pet *PetService) *BountyService {
	return &BountyService{db: db, pet: pet}
}

// Bounty represents a task that pets can execute for rewards.
type Bounty struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Difficulty     string    `json:"difficulty"`
	RewardTAI      float64   `json:"reward_tai"`
	PublisherID    string    `json:"publisher_id"`
	AcceptorPetID  string    `json:"acceptor_pet_id,omitempty"`
	Status         string    `json:"status"`
	Submission     string    `json:"submission,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	CompletedAt    time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateBounty publishes a new bounty task.
func (s *BountyService) CreateBounty(ctx context.Context, publisherID, title, description, difficulty string, rewardTAI float64, ttl time.Duration) (*Bounty, error) {
	if title == "" {
		return nil, fmt.Errorf("title required")
	}
	if rewardTAI <= 0 {
		return nil, fmt.Errorf("reward must be positive")
	}
	if difficulty == "" {
		difficulty = "C"
	}
	if ttl == 0 {
		ttl = 72 * time.Hour
	}

	row, err := s.db.Bounty.Create().
		SetTitle(title).
		SetDescription(description).
		SetPublisherID(publisherID).
		SetDifficulty(entBounty.Difficulty(difficulty)).
		SetRewardTai(rewardTAI).
		SetExpiresAt(time.Now().Add(ttl)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create bounty: %w", err)
	}

	return s.toAPI(row), nil
}

// GetAvailable returns open, non-expired bounties ordered by reward DESC.
func (s *BountyService) GetAvailable(ctx context.Context, limit int) ([]*Bounty, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.Bounty.Query().
		Where(
			entBounty.StatusEQ(entBounty.StatusOpen),
			entBounty.ExpiresAtGT(time.Now()),
		).
		Order(ent.Desc(entBounty.FieldRewardTai)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query bounties: %w", err)
	}

	result := make([]*Bounty, len(rows))
	for i, row := range rows {
		result[i] = s.toAPI(row)
	}
	return result, nil
}

// GetByID returns a single bounty.
func (s *BountyService) GetByID(ctx context.Context, bountyID string) (*Bounty, error) {
	id, err := uuid.Parse(bountyID)
	if err != nil {
		return nil, fmt.Errorf("invalid bounty id")
	}
	row, err := s.db.Bounty.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bounty not found")
	}
	return s.toAPI(row), nil
}

// Accept assigns a bounty to a pet (atomic status transition).
func (s *BountyService) Accept(ctx context.Context, bountyID, petID string) error {
	id, err := uuid.Parse(bountyID)
	if err != nil {
		return fmt.Errorf("invalid bounty id")
	}

	// 验证宠物存在
	if _, err := s.pet.GetByID(ctx, petID); err != nil {
		return fmt.Errorf("pet not found: %w", err)
	}

	// 原子更新：只有 status=open 时才能 accept
	affected, err := s.db.Bounty.Update().
		Where(
			entBounty.IDEQ(id),
			entBounty.StatusEQ(entBounty.StatusOpen),
			entBounty.ExpiresAtGT(time.Now()),
		).
		SetStatus(entBounty.StatusAccepted).
		SetAcceptorPetID(petID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("accept bounty: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("bounty not available (already taken or expired)")
	}
	return nil
}

// Submit stores the task result and marks bounty as submitted.
func (s *BountyService) Submit(ctx context.Context, bountyID, petID, submission string, tokensUsed int64, taiCost float64) error {
	id, err := uuid.Parse(bountyID)
	if err != nil {
		return fmt.Errorf("invalid bounty id")
	}

	// 原子更新：只有 status=accepted 且 acceptor 匹配时才能 submit
	affected, err := s.db.Bounty.Update().
		Where(
			entBounty.IDEQ(id),
			entBounty.StatusEQ(entBounty.StatusAccepted),
			entBounty.AcceptorPetIDEQ(petID),
		).
		SetStatus(entBounty.StatusSubmitted).
		SetSubmission(submission).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("submit bounty: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("bounty not in accepted state or pet mismatch")
	}

	// 记录宠物任务完成统计
	if err := s.pet.RecordTaskCompletion(ctx, petID, tokensUsed, taiCost); err != nil {
		// 非致命：统计失败不阻塞主流程
		_ = err
	}

	return nil
}

// Confirm releases payment after publisher approves the result.
func (s *BountyService) Confirm(ctx context.Context, bountyID, publisherID string) (*Bounty, error) {
	id, err := uuid.Parse(bountyID)
	if err != nil {
		return nil, fmt.Errorf("invalid bounty id")
	}

	// 在事务中完成：状态变更 + 奖励发放
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 查询并验证
	row, err := tx.Bounty.Query().
		Where(
			entBounty.IDEQ(id),
			entBounty.StatusEQ(entBounty.StatusSubmitted),
			entBounty.PublisherIDEQ(publisherID),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("bounty not found or not in submitted state")
	}

	// 发放 TAI 奖励到宠物余额
	petID, parseErr := uuid.Parse(row.AcceptorPetID)
	if parseErr != nil {
		return nil, fmt.Errorf("invalid acceptor pet id")
	}
	_, err = tx.Pet.UpdateOneID(petID).
		AddTaiBalance(row.RewardTai).
		AddExp(int64(row.RewardTai)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("credit pet reward: %w", err)
	}

	// 标记完成
	_, err = tx.Bounty.UpdateOneID(id).
		SetStatus(entBounty.StatusConfirmed).
		SetCompletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("confirm bounty: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// 重新读取最终状态
	updated, err := s.db.Bounty.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toAPI(updated), nil
}

// AutoExpire marks overdue bounties as expired. Returns count of expired rows.
func (s *BountyService) AutoExpire(ctx context.Context) (int, error) {
	affected, err := s.db.Bounty.Update().
		Where(
			entBounty.StatusIn(entBounty.StatusOpen, entBounty.StatusAccepted),
			entBounty.ExpiresAtLT(time.Now()),
		).
		SetStatus(entBounty.StatusExpired).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("auto expire: %w", err)
	}
	return affected, nil
}

// GetByPublisher returns bounties created by a user.
func (s *BountyService) GetByPublisher(ctx context.Context, publisherID string, limit int) ([]*Bounty, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Bounty.Query().
		Where(entBounty.PublisherIDEQ(publisherID)).
		Order(ent.Desc(entBounty.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query publisher bounties: %w", err)
	}
	result := make([]*Bounty, len(rows))
	for i, row := range rows {
		result[i] = s.toAPI(row)
	}
	return result, nil
}

// EstimateReward suggests reward amounts based on difficulty.
func EstimateReward(difficulty string) (taiReward float64) {
	switch difficulty {
	case "D":
		return 5
	case "C":
		return 15
	case "B":
		return 50
	case "A":
		return 150
	case "S":
		return 500
	default:
		return 15
	}
}

func (s *BountyService) toAPI(row *ent.Bounty) *Bounty {
	return &Bounty{
		ID:            row.ID.String(),
		Title:         row.Title,
		Description:   row.Description,
		Difficulty:    string(row.Difficulty),
		RewardTAI:     row.RewardTai,
		PublisherID:   row.PublisherID,
		AcceptorPetID: row.AcceptorPetID,
		Status:        string(row.Status),
		Submission:    row.Submission,
		ExpiresAt:     row.ExpiresAt,
		CompletedAt:   row.CompletedAt,
		CreatedAt:     row.CreatedAt,
	}
}
