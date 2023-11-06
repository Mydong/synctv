package conf

import (
	"github.com/synctv-org/synctv/utils"
)

type Config struct {
	// Log
	Log LogConfig `yaml:"log"`

	// Server
	Server ServerConfig `yaml:"server"`

	// Jwt
	Jwt JwtConfig `yaml:"jwt"`

	// Rtmp
	Rtmp RtmpConfig `yaml:"rtmp" hc:"you can use rtmp to publish live"`

	// Proxy
	Proxy ProxyConfig `yaml:"proxy" hc:"you can use proxy to proxy movie and live when custom headers or network is slow to connect to origin server"`

	// Database
	Database DatabaseConfig `yaml:"database"`

	// OAuth2
	OAuth2 OAuth2Config `yaml:"oauth2"`

	// RateLimit
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

func (c *Config) Save(file string) error {
	return utils.WriteYaml(file, c)
}

func DefaultConfig() *Config {
	return &Config{
		// Log
		Log: DefaultLogConfig(),

		// Server
		Server: DefaultServerConfig(),

		// Jwt
		Jwt: DefaultJwtConfig(),

		// Rtmp
		Rtmp: DefaultRtmpConfig(),

		// Proxy
		Proxy: DefaultProxyConfig(),

		// Database
		Database: DefaultDatabaseConfig(),

		// OAuth2
		OAuth2: DefaultOAuth2Config(),

		// RateLimit
		RateLimit: DefaultRateLimitConfig(),
	}
}
