package config

import (
	"errors"
	"fmt"
	"os"
)

func Load() (Config, error) {
	cfg := Default()
	if v := os.Getenv("WIRECAST_TOWER_ID"); v != "" {
		cfg.TowerID = v
	}
	return cfg, cfg.Validate()
}

type configError string

func (e configError) Error() string { return fmt.Sprintf("wirecast config: %s", string(e)) }

func errConfig(msg string) error { return configError(msg) }

func IsConfigError(err error) bool {
	var ce configError
	return errors.As(err, &ce)
}
