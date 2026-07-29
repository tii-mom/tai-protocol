package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tii-mom/tai-protocol/backend/internal/config"
	"github.com/tii-mom/tai-protocol/backend/internal/handler"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
}

func New(cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	s := &Server{cfg: cfg, router: r}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "tai-protocol"})
	})

	api := s.router.Group("/api/v1")

	// User routes
	user := api.Group("/user")
	{
		user.POST("/auth/tg", handler.TGAuth)           // Telegram 登录
		user.GET("/me", handler.GetMe)                  // 当前用户信息
		user.GET("/me/pets", handler.GetMyPets)         // 我的宠物列表
		user.GET("/me/earnings", handler.GetMyEarnings) // 我的收益
		user.POST("/wallet/bind", handler.BindWallet)   // 绑定 TON 钱包
	}

	// Pet routes
	pet := api.Group("/pet")
	{
		pet.POST("/claim", handler.ClaimPet)            // 领取初始宠物
		pet.GET("/:id", handler.GetPet)                 // 宠物详情
		pet.PUT("/:id/name", handler.RenamePet)         // 改名
		pet.POST("/:id/equip-skill", handler.EquipSkill) // 装备兽决
		pet.POST("/:id/onchain", handler.OnchainPet)    // 上链
		pet.POST("/execute", handler.PetExecute)        // AI代理执行(核心:TAI→3api)
		pet.GET("/:id/usage", handler.PetUsage)         // 宠物算力用量
	}

	// Market / Trade routes
	market := api.Group("/market")
	{
		market.GET("/listings", handler.GetListings)     // 在售列表
		market.GET("/kline", handler.GetKline)           // K线数据
		market.GET("/ranking", handler.GetRanking)       // 排行榜
		market.POST("/order", handler.CreateOrder)       // 挂单
		market.POST("/order/:id/cancel", handler.CancelOrder) // 撤单
		market.POST("/buy", handler.BuyNow)             // 一口价购买
	}

	// Skill (兽决) routes
	skill := api.Group("/skill")
	{
		skill.GET("/list", handler.GetSkillList)         // 兽决商城
		skill.POST("/buy", handler.BuySkill)            // 购买兽决
		skill.POST("/use", handler.UseSkillBook)        // 打书
	}

	// Bounty routes
	bounty := api.Group("/bounty")
	{
		bounty.GET("/available", handler.GetBounties)    // 可接任务
		bounty.POST("/:id/accept", handler.AcceptBounty) // 接单
		bounty.POST("/:id/submit", handler.SubmitBounty) // 提交结果
		bounty.POST("/:id/confirm", handler.ConfirmBounty) // 用户确认
	}

	// Breeding routes
	breed := api.Group("/breed")
	{
		breed.POST("/request", handler.RequestBreed)    // 发起繁殖
		breed.POST("/:id/confirm", handler.ConfirmBreed) // 确认繁殖
		breed.GET("/pool", handler.GetBreedPool)        // 公会繁殖池
	}

	// Guild routes
	guild := api.Group("/guild")
	{
		guild.POST("/create", handler.CreateGuild)
		guild.POST("/:id/join", handler.JoinGuild)
		guild.GET("/:id", handler.GetGuild)
		guild.GET("/ranking", handler.GetGuildRanking)
	}

	// Admin routes (protected)
	admin := api.Group("/admin")
	{
		admin.POST("/pet/mint", handler.AdminMintPet)       // 铸造宠物
		admin.POST("/skill/mint", handler.AdminMintSkill)   // 铸造兽决
		admin.PUT("/market/price", handler.AdminSetPrice)   // 调整价格
		admin.POST("/market/circuit-break", handler.AdminCircuitBreak) // 熔断
		admin.GET("/stats", handler.AdminStats)             // 后台统计
	}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	log.Printf("🚀 TAI Protocol API starting on %s", addr)
	return s.router.Run(addr)
}
