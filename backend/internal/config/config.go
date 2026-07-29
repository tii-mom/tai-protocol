package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port        string `mapstructure:"PORT"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	RedisURL    string `mapstructure:"REDIS_URL"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`

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
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("TON_NETWORK", "testnet")
	viper.SetDefault("TON_CENTER_API", "https://testnet.toncenter.com/api/v2")
	viper.SetDefault("MM_ENABLED", false)
	viper.SetDefault("MM_BOT_COUNT", 5)
	viper.SetDefault("MM_INTERVAL_MIN_SEC", 30)
	viper.SetDefault("MM_INTERVAL_MAX_SEC", 300)

	if err := viper.ReadInConfig(); err != nil {
		// .env file is optional in production (use env vars)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
