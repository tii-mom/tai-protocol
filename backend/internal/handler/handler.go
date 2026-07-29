package handler

import "github.com/gin-gonic/gin"

// === User ===

func TGAuth(c *gin.Context) {
	// TODO: Verify Telegram initData signature
	// TODO: Find or create user by tg_user_id
	// TODO: Issue JWT token
	c.JSON(200, gin.H{"token": "TODO", "user": gin.H{"id": "TODO"}})
}

func GetMe(c *gin.Context) {
	// TODO: Return current user profile + balances
	c.JSON(200, gin.H{"user": "TODO"})
}

func GetMyPets(c *gin.Context) {
	// TODO: List user's pets with status
	c.JSON(200, gin.H{"pets": []any{}})
}

func GetMyEarnings(c *gin.Context) {
	// TODO: Return earnings history + summary
	c.JSON(200, gin.H{"earnings": "TODO"})
}

func BindWallet(c *gin.Context) {
	// TODO: Bind TON wallet address to user
	c.JSON(200, gin.H{"ok": true})
}

// === Pet ===

func ClaimPet(c *gin.Context) {
	// TODO: Issue starter pet (N quality, 1 slot, random species)
	c.JSON(200, gin.H{"pet": "TODO"})
}

func GetPet(c *gin.Context) {
	// TODO: Return full pet details + equipped skills
	c.JSON(200, gin.H{"pet": "TODO"})
}

func RenamePet(c *gin.Context) {
	// TODO: Update pet name
	c.JSON(200, gin.H{"ok": true})
}

func EquipSkill(c *gin.Context) {
	// TODO: Equip skill book to pet (打书 logic)
	// Random overwrite if slots full
	c.JSON(200, gin.H{"ok": true, "result": "TODO"})
}

func OnchainPet(c *gin.Context) {
	// TODO: Mint NFT on TON, update pet.is_on_chain
	c.JSON(200, gin.H{"tx_hash": "TODO"})
}

// === Market ===

func GetListings(c *gin.Context) {
	// TODO: Paginated listings with filters (quality, species, price range)
	c.JSON(200, gin.H{"listings": []any{}, "total": 0})
}

func GetKline(c *gin.Context) {
	// TODO: Return OHLCV data by period (1m/5m/1h/1d)
	c.JSON(200, gin.H{"candles": []any{}})
}

func GetRanking(c *gin.Context) {
	// TODO: Rankings by power/earnings/appreciation
	c.JSON(200, gin.H{"ranking": []any{}})
}

func CreateOrder(c *gin.Context) {
	// TODO: Create buy/sell order
	c.JSON(200, gin.H{"order_id": "TODO"})
}

func CancelOrder(c *gin.Context) {
	// TODO: Cancel open order
	c.JSON(200, gin.H{"ok": true})
}

func BuyNow(c *gin.Context) {
	// TODO: Execute immediate purchase
	c.JSON(200, gin.H{"trade_id": "TODO"})
}

// === Skill ===

func GetSkillList(c *gin.Context) {
	// TODO: Skill shop listing
	c.JSON(200, gin.H{"skills": []any{}})
}

func BuySkill(c *gin.Context) {
	// TODO: Purchase skill book with TAI
	c.JSON(200, gin.H{"ok": true})
}

func UseSkillBook(c *gin.Context) {
	// TODO: Apply skill book to pet (gambling: random overwrite)
	c.JSON(200, gin.H{"result": "TODO", "overwritten": false})
}

// === Bounty ===

func GetBounties(c *gin.Context) {
	// TODO: Available bounties matching user's pets
	c.JSON(200, gin.H{"bounties": []any{}})
}

func AcceptBounty(c *gin.Context) {
	// TODO: Assign bounty to pet
	c.JSON(200, gin.H{"ok": true})
}

func SubmitBounty(c *gin.Context) {
	// TODO: Submit task result for review
	c.JSON(200, gin.H{"ok": true})
}

func ConfirmBounty(c *gin.Context) {
	// TODO: User confirms → release payment
	c.JSON(200, gin.H{"earned": "TODO"})
}

// === Breeding ===

func RequestBreed(c *gin.Context) {
	// TODO: Initiate breeding request (check cooldown, generation)
	c.JSON(200, gin.H{"request_id": "TODO"})
}

func ConfirmBreed(c *gin.Context) {
	// TODO: Both parties confirmed → execute breeding
	c.JSON(200, gin.H{"offspring": "TODO"})
}

func GetBreedPool(c *gin.Context) {
	// TODO: Guild breeding pool listing
	c.JSON(200, gin.H{"pool": []any{}})
}

// === Guild ===

func CreateGuild(c *gin.Context) {
	// TODO: Create guild (costs TAI)
	c.JSON(200, gin.H{"guild_id": "TODO"})
}

func JoinGuild(c *gin.Context) {
	// TODO: Join guild
	c.JSON(200, gin.H{"ok": true})
}

func GetGuild(c *gin.Context) {
	// TODO: Guild details + members
	c.JSON(200, gin.H{"guild": "TODO"})
}

func GetGuildRanking(c *gin.Context) {
	// TODO: Season guild ranking
	c.JSON(200, gin.H{"ranking": []any{}})
}

// === Admin ===

func AdminMintPet(c *gin.Context) {
	// TODO: Admin mint pet (limited editions, genesis)
	c.JSON(200, gin.H{"pet_id": "TODO"})
}

func AdminMintSkill(c *gin.Context) {
	// TODO: Admin mint skill books
	c.JSON(200, gin.H{"ok": true})
}

func AdminSetPrice(c *gin.Context) {
	// TODO: Adjust market maker target price
	c.JSON(200, gin.H{"ok": true})
}

func AdminCircuitBreak(c *gin.Context) {
	// TODO: Emergency pause trading
	c.JSON(200, gin.H{"ok": true, "trading_paused": true})
}

func AdminStats(c *gin.Context) {
	// TODO: Dashboard stats (users, volume, revenue)
	c.JSON(200, gin.H{"stats": "TODO"})
}
