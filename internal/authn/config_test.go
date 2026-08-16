package authn

import "testing"

func TestLoadConfigDefaultsToFirebaseProduction(t *testing.T) {
	config, err := LoadConfig(func(string) string { return "" })
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.AppEnv != "production" || config.Mode != ModeFirebase {
		t.Fatalf("config = %#v, want production Firebase config", config)
	}
}

func TestLoadConfigAcceptsDevelopmentDevMode(t *testing.T) {
	values := map[string]string{
		"APP_ENV":     "development",
		"AUTH_MODE":   "dev",
		"DEV_USER_ID": "1",
	}
	config, err := LoadConfig(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.Mode != ModeDev || config.DevUserID != 1 {
		t.Fatalf("config = %#v, want dev user 1", config)
	}
}

func TestLoadConfigRejectsUnsafeDevMode(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
	}{
		{
			name: "production environment",
			values: map[string]string{
				"APP_ENV":     "production",
				"AUTH_MODE":   "dev",
				"DEV_USER_ID": "1",
			},
		},
		{
			name: "missing user ID",
			values: map[string]string{
				"APP_ENV":   "development",
				"AUTH_MODE": "dev",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := LoadConfig(func(key string) string { return test.values[key] }); err == nil {
				t.Fatal("LoadConfig() error = nil, want validation error")
			}
		})
	}
}
