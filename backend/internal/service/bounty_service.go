package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tii-mom/tai-protocol/backend/ent"
)

// BountyService handles bounty task lifecycle.
type BountyService struct {
	db *ent.Client
}

func NewBountyService(db *ent.Client) *BountyService {
	return &BountyService{db: db}
}

// Bounty represents a task that pets can execute for rewards.
type Bounty struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Difficulty  string    `json:"difficulty"` // D/C/B/A/S
	RewardTAI   float64   `json:"reward_tai"`
	RewardUSDT  float64   `json:"reward_usdt"`
	RequiredSkills []string `json:"required_skills"`
	MaxCalls    int       `json:"max_calls"` // estimated API calls
	PublisherID string    `json:"publisher_id"`
	AcceptorID  string    `json:"acceptor_id,omitempty"`
	PetID       string    `json:"pet_id,omitempty"`
	Status      string    `json:"status"` // open/accepted/submitted/completed/expired
	Result      string    `json:"result,omitempty"`
	Deadline    time.Time `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
}

// BountyStatus constants
const (
	BountyOpen      = "open"
	BountyAccepted  = "accepted"
	BountySubmitted = "submitted"
	BountyCompleted = "completed"
	BountyExpired   = "expired"
)

// CreateBounty publishes a new bounty task.
func (s *BountyService) CreateBounty(ctx context.Context, b *Bounty) (*Bounty, error) {
	if b.Title == "" || b.Description == "" {
		return nil, fmt.Errorf("title and description required")
	}
	if b.RewardTAI <= 0 && b.RewardUSDT <= 0 {
		return nil, fmt.Errorf("at least one reward must be positive")
	}
	if b.Deadline.IsZero() {
		b.Deadline = time.Now().Add(72 * time.Hour) // default 3 days
	}
	if b.Difficulty == "" {
		b.Difficulty = "C"
	}
	if b.MaxCalls == 0 {
		b.MaxCalls = 3
	}

	// TODO: Insert into DB via Ent (bounty schema)
	// For now, generate ID and return
	b.ID = uuid.New().String()
	b.Status = BountyOpen
	b.CreatedAt = time.Now()

	return b, nil
}

// GetAvailable returns open bounties matching a pet's capabilities.
func (s *BountyService) GetAvailable(ctx context.Context, petID string) ([]*Bounty, error) {
	// TODO: Query bounties WHERE status='open' AND deadline > now()
	// TODO: Filter by pet's skills and intelligence
	// TODO: Order by reward_tai DESC

	// Placeholder: return empty (real query after Ent bounty schema is added)
	return []*Bounty{}, nil
}

// Accept assigns a bounty to a pet.
func (s *BountyService) Accept(ctx context.Context, bountyID, petID string) error {
	// TODO: Transaction:
	//   1. Lock bounty row (FOR UPDATE)
	//   2. Verify status == 'open'
	//   3. Verify pet exists and is 'idle'
	//   4. Set status='accepted', pet_id=petID, acceptor_id=pet.owner_id
	//   5. Set pet status='working'
	return nil
}

// Submit stores the task result for review.
func (s *BountyService) Submit(ctx context.Context, bountyID, petID string, result string, success bool, tokensUsed int64, taiCost float64) error {
	// TODO: Transaction:
	//   1. Verify bounty status == 'accepted' AND pet_id matches
	//   2. Store result
	//   3. Set status='submitted'
	//   4. Set pet status='idle'
	//   5. Record task completion stats on pet
	return nil
}

// Confirm releases payment after user approves the result.
func (s *BountyService) Confirm(ctx context.Context, bountyID, userID string) (*Bounty, error) {
	// TODO: Transaction:
	//   1. Verify bounty status == 'submitted'
	//   2. Verify user is the publisher
	//   3. Transfer reward_tai to pet's tai_balance
	//   4. Transfer reward_usdt to user's balance_usdt (from escrow)
	//   5. Set status='completed'
	//   6. Add exp to pet
	return nil, fmt.Errorf("not implemented")
}

// AutoExpire marks overdue bounties as expired.
// Called by a background cron job.
func (s *BountyService) AutoExpire(ctx context.Context) (int, error) {
	// TODO: UPDATE bounties SET status='expired' WHERE status IN ('open','accepted') AND deadline < now()
	return 0, nil
}

// EstimateReward suggests reward amounts based on difficulty.
func EstimateReward(difficulty string) (taiReward, usdtReward float64) {
	switch difficulty {
	case "D":
		return 5, 0.01
	case "C":
		return 15, 0.03
	case "B":
		return 50, 0.1
	case "A":
		return 150, 0.3
	case "S":
		return 500, 1.0
	default:
		return 15, 0.03
	}
}
