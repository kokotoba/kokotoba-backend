package authn

import (
	"fmt"
	"strconv"
	"strings"
)

type Mode string

const (
	ModeFirebase Mode = "firebase"
	ModeDev      Mode = "dev"
)

type Config struct {
	AppEnv            string
	Mode              Mode
	DevUserID         int64
	FirebaseProjectID string
}

func LoadConfig(getenv func(string) string) (Config, error) {
	config := Config{
		AppEnv:            strings.TrimSpace(getenv("APP_ENV")),
		Mode:              Mode(strings.TrimSpace(getenv("AUTH_MODE"))),
		FirebaseProjectID: strings.TrimSpace(getenv("FIREBASE_PROJECT_ID")),
	}
	if config.AppEnv == "" {
		config.AppEnv = "production"
	}
	if config.Mode == "" {
		config.Mode = ModeFirebase
	}

	switch config.Mode {
	case ModeFirebase:
		return config, nil
	case ModeDev:
		if config.AppEnv != "development" {
			return Config{}, fmt.Errorf(
				"AUTH_MODE=dev is only allowed when APP_ENV=development",
			)
		}
		devUserID, err := strconv.ParseInt(strings.TrimSpace(getenv("DEV_USER_ID")), 10, 64)
		if err != nil || devUserID <= 0 {
			return Config{}, fmt.Errorf("DEV_USER_ID must be a positive integer")
		}
		config.DevUserID = devUserID
		return config, nil
	default:
		return Config{}, fmt.Errorf("AUTH_MODE must be firebase or dev")
	}
}
