package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tii-mom/tai-protocol/backend/ent"
	entPet "github.com/tii-mom/tai-protocol/backend/ent/pet"
)

// PetService handles pet creation, lookup, and state management.
type PetService struct {
	db *ent.Client
}

func NewPetService(db *ent.Client) *PetService {
	return &PetService{db: db}
}

// Pet is the API-facing pet representation.
type Pet struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	Species     string    `json:"species"`
	Quality     string    `json:"quality"`
	Generation  int       `json:"generation"`
	GrowthRate  float64   `json:"growth_rate"`
	AptHP       int       `json:"apt_hp"`
	AptATK      int       `json:"apt_atk"`
	AptDEF      int       `json:"apt_def"`
	AptSPD      int       `json:"apt_spd"`
	AptINT      int       `json:"apt_int"`
	SkillSlots  int       `json:"skill_slots"`
	Personality string    `json:"personality"`
	Level       int       `json:"level"`
	Exp         int64     `json:"exp"`
	Mood        int       `json:"mood"`
	Energy      int       `json:"energy"`
	Status      string    `json:"status"`
	TAIBalance  float64   `json:"tai_balance"`
	ImageURL    string    `json:"image_url,omitempty"`
	IsOnChain   bool      `json:"is_on_chain"`
	TasksDone   int       `json:"total_tasks_done"`
	TokensUsed  int64     `json:"total_tokens_used"`
	CreatedAt   time.Time `json:"created_at"`
}

var starterSpecies = []string{
	"Rex-Frame", "Falcon-Unit", "Titan-Core", "Viper-Drive", "Phoenix-Shell",
}

var personalities = []string{
	"aggressive", "defensive", "balanced", "speedster", "analytical",
}

// ClaimStarterPet creates a new starter pet (one free Gen-0 per user).
func (s *PetService) ClaimStarterPet(ctx context.Context, ownerID string, ownerName string) (*Pet, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, fmt.Errorf("invalid owner id")
	}

	existing, err := s.db.Pet.Query().
		Where(entPet.OwnerID(uid), entPet.Generation(0)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("check existing: %w", err)
	}
	if existing > 0 {
		return nil, fmt.Errorf("you already claimed your starter pet")
	}

	species := starterSpecies[rand.Intn(len(starterSpecies))]
	personality := personalities[rand.Intn(len(personalities))]
	apt := generateAptitudes(species)

	newPet, err := s.db.Pet.Create().
		SetOwnerID(uid).
		SetName(fmt.Sprintf("%s-α", species)).
		SetSpecies(species).
		SetQuality("common").
		SetGeneration(0).
		SetGrowthRate(1.0 + rand.Float64()*0.2).
		SetAptHp(apt[0]).
		SetAptAtk(apt[1]).
		SetAptDef(apt[2]).
		SetAptSpd(apt[3]).
		SetAptInt(apt[4]).
		SetSkillSlots(2).
		SetPersonality(personality).
		SetLevel(1).
		SetMood(80).
		SetEnergy(100).
		SetStatus("idle").
		SetTaiBalance(10).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create pet: %w", err)
	}

	return s.toAPIPet(newPet), nil
}

// GetByID returns a pet by UUID string.
func (s *PetService) GetByID(ctx context.Context, petID string) (*Pet, error) {
	pid, err := uuid.Parse(petID)
	if err != nil {
		return nil, fmt.Errorf("invalid pet id")
	}
	p, err := s.db.Pet.Get(ctx, pid)
	if err != nil {
		return nil, fmt.Errorf("pet not found")
	}
	return s.toAPIPet(p), nil
}

// GetByOwner returns all pets owned by a user.
func (s *PetService) GetByOwner(ctx context.Context, ownerID string) ([]*Pet, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, fmt.Errorf("invalid owner id")
	}
	pets, err := s.db.Pet.Query().
		Where(entPet.OwnerID(uid)).
		Order(ent.Desc(entPet.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*Pet, len(pets))
	for i, p := range pets {
		result[i] = s.toAPIPet(p)
	}
	return result, nil
}

// DeductPetTAI deducts TAI from a pet's balance.
func (s *PetService) DeductPetTAI(ctx context.Context, petID string, amount float64) error {
	pid, err := uuid.Parse(petID)
	if err != nil {
		return fmt.Errorf("invalid pet id")
	}
	_, err = s.db.Pet.UpdateOneID(pid).
		AddTaiBalance(-amount).
		AddTotalSpentTai(amount).
		Save(ctx)
	return err
}

// AddPetTAI adds TAI earnings to a pet.
func (s *PetService) AddPetTAI(ctx context.Context, petID string, amount float64) error {
	pid, err := uuid.Parse(petID)
	if err != nil {
		return fmt.Errorf("invalid pet id")
	}
	_, err = s.db.Pet.UpdateOneID(pid).
		AddTaiBalance(amount).
		AddTotalEarnedTai(amount).
		Save(ctx)
	return err
}

// RecordTaskCompletion updates pet stats after a task.
func (s *PetService) RecordTaskCompletion(ctx context.Context, petID string, tokensUsed int64, taiCost float64) error {
	pid, err := uuid.Parse(petID)
	if err != nil {
		return err
	}
	_, err = s.db.Pet.UpdateOneID(pid).
		AddTotalTokensUsed(tokensUsed).
		AddTotalTasksDone(1).
		AddTotalSpentTai(taiCost).
		AddTaiBalance(-taiCost).
		AddExp(tokensUsed / 10).
		Save(ctx)
	return err
}

func (s *PetService) toAPIPet(p *ent.Pet) *Pet {
	return &Pet{
		ID:          p.ID.String(),
		OwnerID:     p.OwnerID.String(),
		Name:        p.Name,
		Species:     p.Species,
		Quality:     p.Quality,
		Generation:  p.Generation,
		GrowthRate:  p.GrowthRate,
		AptHP:       p.AptHp,
		AptATK:      p.AptAtk,
		AptDEF:      p.AptDef,
		AptSPD:      p.AptSpd,
		AptINT:      p.AptInt,
		SkillSlots:  p.SkillSlots,
		Personality: p.Personality,
		Level:       p.Level,
		Exp:         p.Exp,
		Mood:        p.Mood,
		Energy:      p.Energy,
		Status:      p.Status,
		TAIBalance:  p.TaiBalance,
		ImageURL:    p.ImageURL,
		IsOnChain:   p.IsOnChain,
		TasksDone:   p.TotalTasksDone,
		TokensUsed:  p.TotalTokensUsed,
		CreatedAt:   p.CreatedAt,
	}
}

func generateAptitudes(species string) [5]int {
	var base [5]int
	switch species {
	case "Rex-Frame":
		base = [5]int{30, 20, 25, 10, 15}
	case "Falcon-Unit":
		base = [5]int{15, 15, 10, 35, 25}
	case "Titan-Core":
		base = [5]int{25, 30, 20, 10, 15}
	case "Viper-Drive":
		base = [5]int{15, 25, 10, 30, 20}
	case "Phoenix-Shell":
		base = [5]int{20, 10, 15, 15, 40}
	default:
		base = [5]int{20, 20, 20, 20, 20}
	}
	var result [5]int
	for i := range base {
		jitter := rand.Intn(11) - 5
		result[i] = base[i] + jitter
		if result[i] < 5 {
			result[i] = 5
		}
		if result[i] > 50 {
			result[i] = 50
		}
	}
	return result
}
