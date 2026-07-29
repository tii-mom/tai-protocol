package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// PetService handles pet creation, lookup, and state management.
type PetService struct {
	// TODO: inject *ent.Client
}

func NewPetService() *PetService {
	return &PetService{}
}

// Pet represents a TAI Protocol mecha-pet.
type Pet struct {
	ID          string
	OwnerID     int64
	Species     string
	Name        string
	Quality     string // common, rare, epic, legendary, mythic
	Generation  int
	GrowthRate  float64
	AptHP       int
	AptATK      int
	AptDEF      int
	AptSPD      int
	AptINT      int
	SkillSlots  int
	Personality string
	Level       int
	Exp         int
	Mood        int
	Energy      int
	Status      string // idle, working, breeding, trading, resting
	TAIBalance  float64 // pet's own TAI earnings
	ImageURL    string
	CreatedAt   time.Time
}

// Species pool for starter pets.
var starterSpecies = []string{
	"Rex-Frame",     // 暴龙骨架
	"Falcon-Unit",   // 猎鹰单元
	"Titan-Core",    // 泰坦核心
	"Viper-Drive",   // 毒蛇驱动
	"Phoenix-Shell", // 凤凰外壳
}

var personalities = []string{
	"aggressive", "defensive", "balanced", "speedster", "analytical",
}

// ClaimStarterPet creates a new starter pet for a user.
// Rules:
//   - One free starter per user (Generation 0, Common quality)
//   - Random species, random aptitudes (weighted by species)
//   - 2 skill slots (upgradeable via breeding)
func (s *PetService) ClaimStarterPet(ctx context.Context, ownerID int64, ownerName string) (*Pet, error) {
	// TODO: Check if user already claimed (query pets WHERE owner_id = ? AND generation = 0)
	// TODO: If exists, return error "already claimed"

	species := starterSpecies[rand.Intn(len(starterSpecies))]
	personality := personalities[rand.Intn(len(personalities))]

	// Generate aptitudes based on species archetype
	apt := generateAptitudes(species)

	pet := &Pet{
		ID:          fmt.Sprintf("gen0-%d-%d", ownerID, time.Now().UnixMilli()%10000),
		OwnerID:     ownerID,
		Species:     species,
		Name:        fmt.Sprintf("%s-α", species),
		Quality:     "common",
		Generation:  0,
		GrowthRate:  1.0 + rand.Float64()*0.2, // 1.0~1.2
		AptHP:       apt[0],
		AptATK:      apt[1],
		AptDEF:      apt[2],
		AptSPD:      apt[3],
		AptINT:      apt[4],
		SkillSlots:  2,
		Personality: personality,
		Level:       1,
		Exp:         0,
		Mood:        80,
		Energy:      100,
		Status:      "idle",
		TAIBalance:  10.0, // starter bonus: 10 TAI
		CreatedAt:   time.Now(),
	}

	// TODO: Insert into DB via Ent
	// TODO: Add 10 TAI to user's balance (starter bonus)

	return pet, nil
}

// GetByID returns a pet by its ID.
func (s *PetService) GetByID(ctx context.Context, petID string) (*Pet, error) {
	// TODO: Ent query
	return nil, fmt.Errorf("pet %s not found", petID)
}

// GetByOwner returns all pets owned by a user.
func (s *PetService) GetByOwner(ctx context.Context, ownerID int64) ([]*Pet, error) {
	// TODO: Ent query
	return []*Pet{}, nil
}

// DeductPetTAI deducts TAI from a pet's balance for compute costs.
func (s *PetService) DeductPetTAI(ctx context.Context, petID string, amount float64) error {
	// TODO: UPDATE pets SET tai_balance = tai_balance - ? WHERE id = ? AND tai_balance >= ?
	return nil
}

// AddPetTAI adds TAI earnings to a pet's balance.
func (s *PetService) AddPetTAI(ctx context.Context, petID string, amount float64) error {
	// TODO: UPDATE pets SET tai_balance = tai_balance + ? WHERE id = ?
	return nil
}

// generateAptitudes creates weighted aptitudes based on species archetype.
// Total aptitude points = 100, distributed by species tendency.
func generateAptitudes(species string) [5]int {
	// Base distributions [HP, ATK, DEF, SPD, INT]
	var base [5]int
	switch species {
	case "Rex-Frame": // tank: high HP/DEF
		base = [5]int{30, 20, 25, 10, 15}
	case "Falcon-Unit": // speed: high SPD/INT
		base = [5]int{15, 15, 10, 35, 25}
	case "Titan-Core": // power: high ATK/HP
		base = [5]int{25, 30, 20, 10, 15}
	case "Viper-Drive": // assassin: high SPD/ATK
		base = [5]int{15, 25, 10, 30, 20}
	case "Phoenix-Shell": // mage: high INT/HP
		base = [5]int{20, 10, 15, 15, 40}
	default:
		base = [5]int{20, 20, 20, 20, 20}
	}

	// Add randomness ±5 per stat
	var result [5]int
	for i := range base {
		jitter := rand.Intn(11) - 5 // -5 to +5
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
