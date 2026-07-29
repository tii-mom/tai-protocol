package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port        string `mapstructure:"PORT"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	RedisURL    string `mapstructure:"REDIS_URL"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
	TGBotToken  string `mapstructure:"TG_BOT_TOKEN"`

	// TON
	TonCenterAPI string `mapstructure:"TON_CENTER_API"`
	TonNetwork   string `mapstructure:"TON_NETWORK"` // mainnet / testnet

	// Contract addresses (filled after deployment)
	PetNFTCollection string `mapstructure:"PET_NFT_COLLECTION"`
	TaiTokenMaster   string `mapstructure:"TAI_TOKEN_MASTER"`
	BreedingContract string `mapstructure:"BREEDING_CONTRACT"`
	MarketContract   string `mapstructure:"MARKET_CONTRACT"`
	BountyVault      string `mapstructure:"BOUNTY_VAULT"`
	AdVault          string `mapstructure:"AD_VAULT"`
	TreasuryWallet   string `mapstructure:"TREASURY_WALLET"`

	// 3api.shop integration (Phase 0: platform pool model)
	ThreeAPIBaseURL        string `mapstructure:"THREEAPI_BASE_URL"`
	ThreeAPIPlatformKey    string `mapstructure:"THREEAPI_PLATFORM_KEY"`    // single key for all pets
	ThreeAPIInternalSecret string `mapstructure:"THREEAPI_INTERNAL_SECRET"` // for reconciliation endpoints

	// Market maker
	MMEnabled     bool `mapstructure:"MM_ENABLED"`
	MMBotCount    int  `mapstructure:"MM_BOT_COUNT"`
	MMIntervalMin int  `mapstructure:"MM_INTERVAL_MIN_SEC"`
	MMIntervalMax int  `mapstructure:"MM_INTERVAL_MAX_SEC"`
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("TON_NETWORK", "testnet")
	viper.SetDefault("TON_CENTER_API", "https://testnet.toncenter.com/api/v2")
	viper.SetDefault("MM_ENABLED", false)
	viper.SetDefault("MM_BOT_COUNT", 5)
	viper.SetDefault("MM_INTERVAL_MIN_SEC", 30)
	viper.SetDefault("MM_INTERVAL_MAX_SEC", 300)

	// Bind env vars explicitly (required for Unmarshal to see them)
	for _, key := range []string{
		"PORT", "DATABASE_URL", "REDIS_URL", "JWT_SECRET", "TG_BOT_TOKEN",
		"TON_CENTER_API", "TON_NETWORK",
		"PET_NFT_COLLECTION", "TAI_TOKEN_MASTER", "BREEDING_CONTRACT",
		"MARKET_CONTRACT", "BOUNTY_VAULT", "AD_VAULT", "TREASURY_WALLET",
		"THREEAPI_BASE_URL", "THREEAPI_PLATFORM_KEY", "THREEAPI_INTERNAL_SECRET",
		"MM_ENABLED", "MM_BOT_COUNT", "MM_INTERVAL_MIN_SEC", "MM_INTERVAL_MAX_SEC",
	} {
		_ = viper.BindEnv(key)
	}

	// .env is optional — in production, env vars via AutomaticEnv() take over
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
