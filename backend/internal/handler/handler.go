package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tii-mom/tai-protocol/backend/ent"
	"github.com/tii-mom/tai-protocol/backend/internal/auth"
	"github.com/tii-mom/tai-protocol/backend/internal/service"
	"github.com/tii-mom/tai-protocol/backend/internal/threeapi"
)

// Handlers holds all dependencies for HTTP handlers.
var (
	UserService   *service.UserService
	PetService    *service.PetService
	BountyService *service.BountyService
	JWT           *auth.JWTManager
	ThreeAPI      *threeapi.Client
	DB            *ent.Client
	BotToken      string
)

// InitHandlers sets up shared dependencies. Called from server.New().
func InitHandlers(userSvc *service.UserService, petSvc *service.PetService, bountySvc *service.BountyService, jwtMgr *auth.JWTManager, apiClient *threeapi.Client, db *ent.Client, botToken string) {
	UserService = userSvc
	PetService = petSvc
	BountyService = bountySvc
	JWT = jwtMgr
	ThreeAPI = apiClient
	DB = db
	BotToken = botToken
}

// === User ===

// TGAuth verifies Telegram initData and returns a JWT.
func TGAuth(c *gin.Context) {
	var req struct {
		InitData string `json:"init_data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "init_data required"})
		return
	}

	data, err := auth.VerifyInitData(req.InitData, BotToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid telegram auth: " + err.Error()})
		return
	}

	user, isNew, err := UserService.FindOrCreateByTG(c.Request.Context(), data.User.ID, data.User.Username, data.User.FirstName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user creation failed"})
		return
	}

	token, err := JWT.IssueToken(user.ID, user.TGUserID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"user":    user,
		"is_new":  isNew,
		"message": "welcome to TAI Protocol",
	})
}

func GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := UserService.GetByID(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetMyPets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	pets, err := PetService.GetByOwner(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pets": pets})
}

func GetMyEarnings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"earnings": []any{}, "total_tai": 0, "total_usdt": 0})
}

func BindWallet(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "wallet": req.Address})
}

// === Pet ===

// ClaimPet issues a free starter pet (one per user).
func ClaimPet(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	pet, err := PetService.ClaimStarterPet(c.Request.Context(), userID.(string), username.(string))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"pet":     pet,
		"bonus":   "10 TAI starter bonus credited",
		"message": "Your mecha-pet is ready! Send it on bounty missions to earn TAI.",
	})
}

func GetPet(c *gin.Context) {
	petID := c.Param("id")
	pet, err := PetService.GetByID(c.Request.Context(), petID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pet": pet})
}

func RenamePet(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required,max=32"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required (max 32 chars)"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "name": req.Name})
}

func EquipSkill(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "result": "TODO"})
}

func OnchainPet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"tx_hash": "TODO"})
}

// PetExecute is the core AI proxy endpoint (Phase 0).
// Bot → this → 3api (platform key). TAI deducted from pet balance.
func PetExecute(c *gin.Context) {
	var req struct {
		PetID    string `json:"pet_id" binding:"required"`
		TaskID   string `json:"task_id"`
		Model    string `json:"model" binding:"required"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages" binding:"required,min=1"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Load pet
	pet, err := PetService.GetByID(c.Request.Context(), req.PetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pet not found"})
		return
	}
	if pet.Status == "resting" || pet.Status == "breeding" {
		c.JSON(http.StatusConflict, gin.H{"error": "pet is " + pet.Status})
		return
	}

	// 2. Check balance
	estimatedTokens := int64(req.MaxTokens)
	if estimatedTokens == 0 {
		estimatedTokens = 2000
	}
	estimatedCost := threeapi.CalculateTAICost(req.Model, estimatedTokens)
	if pet.TAIBalance < estimatedCost {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":          "insufficient TAI",
			"pet_balance":    pet.TAIBalance,
			"estimated_cost": estimatedCost,
		})
		return
	}

	// 3. Proxy to 3api
	messages := make([]threeapi.ChatMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = threeapi.ChatMessage{Role: m.Role, Content: m.Content}
	}

	result, err := ThreeAPI.ChatCompletion(c.Request.Context(), threeapi.ChatRequest{
		PetID:       req.PetID,
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "3api: " + err.Error()})
		return
	}

	// 4. Deduct TAI + record usage
	_ = PetService.DeductPetTAI(c.Request.Context(), req.PetID, result.TAICost)
	_, _ = DB.UsageLog.Create().
		SetPetID(req.PetID).
		SetModel(req.Model).
		SetPromptTokens(int(result.TokensUsed / 2)).
		SetCompletionTokens(int(result.TokensUsed / 2)).
		SetTaiCost(result.TAICost).
		SetBountyID(req.TaskID).
		SetCreatedAt(time.Now()).
		Save(c.Request.Context())

	// 5. Respond
	c.JSON(http.StatusOK, gin.H{
		"content":     result.Content,
		"model":       result.Model,
		"tokens_used": result.TokensUsed,
		"tai_cost":    result.TAICost,
		"duration_ms": result.Duration,
		"pet_balance": pet.TAIBalance - result.TAICost,
	})
}

// PetUsage returns a pet's compute usage stats.
func PetUsage(c *gin.Context) {
	petID := c.Param("id")
	usage := ThreeAPI.GetPetUsage(petID)
	if usage == nil {
		c.JSON(http.StatusOK, gin.H{"pet_id": petID, "total_calls": 0, "total_tokens": 0, "total_tai": 0.0})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"pet_id":       petID,
		"total_calls":  usage.Calls,
		"total_tokens": usage.TokensUsed,
		"total_tai":    usage.TAISpent,
	})
}

// === Market ===

func GetListings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"listings": []any{}, "total": 0})
}

func GetKline(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"candles": []any{}})
}

func GetRanking(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ranking": []any{}})
}

func CreateOrder(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"order_id": "TODO"})
}

func CancelOrder(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func BuyNow(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"trade_id": "TODO"})
}

// === Skill ===

func GetSkillList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"skills": []any{}})
}

func BuySkill(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UseSkillBook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"result": "TODO", "overwritten": false})
}

// === Bounty ===

// CreateBounty publishes a new bounty task.
func CreateBounty(c *gin.Context) {
	var req struct {
		Title         string  `json:"title" binding:"required"`
		Description   string  `json:"description"`
		Difficulty    string  `json:"difficulty" binding:"required,oneof=D C B A S"`
		RewardTAI     float64 `json:"reward_tai"`
		DeadlineHours int     `json:"deadline_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto-calculate reward if not specified
	if req.RewardTAI == 0 {
		req.RewardTAI = service.EstimateReward(req.Difficulty)
	}

	userID, _ := c.Get("user_id")
	ttl := time.Duration(req.DeadlineHours) * time.Hour

	bounty, err := BountyService.CreateBounty(c.Request.Context(), userID.(string), req.Title, req.Description, req.Difficulty, req.RewardTAI, ttl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"bounty":  bounty,
		"message": "bounty published, pets will auto-accept matching tasks",
	})
}

// GetMyBounties returns bounties published by the current user.
func GetMyBounties(c *gin.Context) {
	userID, _ := c.Get("user_id")
	bounties, err := BountyService.GetByPublisher(c.Request.Context(), userID.(string), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bounties": bounties})
}

// GetBounties returns available (open) bounties.
func GetBounties(c *gin.Context) {
	bounties, err := BountyService.GetAvailable(c.Request.Context(), 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bounties": bounties})
}

// AcceptBounty assigns a bounty to a pet.
func AcceptBounty(c *gin.Context) {
	bountyID := c.Param("id")
	var req struct {
		PetID string `json:"pet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pet_id required"})
		return
	}

	if err := BountyService.Accept(c.Request.Context(), bountyID, req.PetID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "bounty accepted"})
}

// SubmitBounty stores task result for review.
func SubmitBounty(c *gin.Context) {
	bountyID := c.Param("id")
	var req struct {
		PetID      string `json:"pet_id" binding:"required"`
		Submission string `json:"submission" binding:"required"`
		TokensUsed int64  `json:"tokens_used"`
		TAICost    float64 `json:"tai_cost"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := BountyService.Submit(c.Request.Context(), bountyID, req.PetID, req.Submission, req.TokensUsed, req.TAICost); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "submission received, awaiting publisher confirmation"})
}

// ConfirmBounty releases payment after publisher approves.
func ConfirmBounty(c *gin.Context) {
	bountyID := c.Param("id")
	userID, _ := c.Get("user_id")

	bounty, err := BountyService.Confirm(c.Request.Context(), bountyID, userID.(string))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"bounty":  bounty,
		"message": "reward released to pet",
	})
}

// === Breeding ===

func RequestBreed(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"request_id": "TODO"})
}

func ConfirmBreed(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"offspring": "TODO"})
}

func GetBreedPool(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"pool": []any{}})
}

// === Guild ===

func CreateGuild(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"guild_id": "TODO"})
}

func JoinGuild(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetGuild(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"guild": "TODO"})
}

func GetGuildRanking(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ranking": []any{}})
}

// === Admin ===

func AdminMintPet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"pet_id": "TODO"})
}

func AdminMintSkill(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminSetPrice(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminCircuitBreak(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "trading_paused": true})
}

func AdminStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"stats": "TODO"})
}
