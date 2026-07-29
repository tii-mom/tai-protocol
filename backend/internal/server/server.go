package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tii-mom/tai-protocol/backend/ent"
	"github.com/tii-mom/tai-protocol/backend/internal/auth"
	"github.com/tii-mom/tai-protocol/backend/internal/config"
	"github.com/tii-mom/tai-protocol/backend/internal/handler"
	"github.com/tii-mom/tai-protocol/backend/internal/service"
	"github.com/tii-mom/tai-protocol/backend/internal/threeapi"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
	jwt    *auth.JWTManager
}

func New(cfg *config.Config, db *ent.Client) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Initialize services
	jwtMgr := auth.NewJWTManager(cfg.JWTSecret)
	userSvc := service.NewUserService(db)
	petSvc := service.NewPetService(db)
	apiClient := threeapi.NewClient(threeapi.Config{
		BaseURL:     cfg.ThreeAPIBaseURL,
		PlatformKey: cfg.ThreeAPIPlatformKey,
	})

	// Wire handlers
	handler.InitHandlers(userSvc, petSvc, jwtMgr, apiClient, cfg.TGBotToken)

	s := &Server{cfg: cfg, router: r, jwt: jwtMgr}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "tai-protocol"})
	})

	api := s.router.Group("/api/v1")

	// ─── Public routes (no auth) ─────────────────────────
	api.POST("/user/auth/tg", handler.TGAuth) // Telegram 登录
	api.GET("/market/listings", handler.GetListings)
	api.GET("/market/kline", handler.GetKline)
	api.GET("/market/ranking", handler.GetRanking)
	api.GET("/skill/list", handler.GetSkillList)
	api.GET("/bounty/available", handler.GetBounties)

	// ─── Protected routes (JWT required) ─────────────────
	protected := api.Group("")
	protected.Use(s.jwt.Middleware())
	{
		// User
		protected.GET("/user/me", handler.GetMe)
		protected.GET("/user/me/pets", handler.GetMyPets)
		protected.GET("/user/me/earnings", handler.GetMyEarnings)
		protected.POST("/user/wallet/bind", handler.BindWallet)

		// Pet
		protected.POST("/pet/claim", handler.ClaimPet)
		protected.GET("/pet/:id", handler.GetPet)
		protected.PUT("/pet/:id/name", handler.RenamePet)
		protected.POST("/pet/:id/equip-skill", handler.EquipSkill)
		protected.POST("/pet/:id/onchain", handler.OnchainPet)
		protected.POST("/pet/execute", handler.PetExecute)
		protected.GET("/pet/:id/usage", handler.PetUsage)

		// Market (write ops need auth)
		protected.POST("/market/order", handler.CreateOrder)
		protected.POST("/market/order/:id/cancel", handler.CancelOrder)
		protected.POST("/market/buy", handler.BuyNow)

		// Skill
		protected.POST("/skill/buy", handler.BuySkill)
		protected.POST("/skill/use", handler.UseSkillBook)

		// Bounty
		protected.POST("/bounty/:id/accept", handler.AcceptBounty)
		protected.POST("/bounty/:id/submit", handler.SubmitBounty)
		protected.POST("/bounty/:id/confirm", handler.ConfirmBounty)

		// Breeding
		protected.POST("/breed/request", handler.RequestBreed)
		protected.POST("/breed/:id/confirm", handler.ConfirmBreed)
		protected.GET("/breed/pool", handler.GetBreedPool)

		// Guild
		protected.POST("/guild/create", handler.CreateGuild)
		protected.POST("/guild/:id/join", handler.JoinGuild)
		protected.GET("/guild/:id", handler.GetGuild)
		protected.GET("/guild/ranking", handler.GetGuildRanking)
	}

	// ─── Admin routes (separate auth, TODO: admin middleware) ───
	admin := api.Group("/admin")
	{
		admin.POST("/pet/mint", handler.AdminMintPet)
		admin.POST("/skill/mint", handler.AdminMintSkill)
		admin.PUT("/market/price", handler.AdminSetPrice)
		admin.POST("/market/circuit-break", handler.AdminCircuitBreak)
		admin.GET("/stats", handler.AdminStats)
	}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	log.Printf("TAI Protocol API starting on %s", addr)
	return s.router.Run(addr)
}
