package conf

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP string
	GRPC string
	DSN  string
}

// Load deliberately has no dependency on platform authentication or execution settings.
func Load() (Config, error) {
	v := viper.New()
	v.SetEnvPrefix("TESTMOON_SCHOOL_MANAGEMENT")
	v.AutomaticEnv()
	v.SetDefault("http", "127.0.0.1:8096")
	v.SetDefault("grpc", "127.0.0.1:9096")
	c := Config{HTTP: v.GetString("http"), GRPC: v.GetString("grpc"), DSN: v.GetString("dsn")}
	if strings.TrimSpace(c.DSN) == "" {
		return Config{}, fmt.Errorf("TESTMOON_SCHOOL_MANAGEMENT_DSN is required")
	}
	return c, nil
}
