package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()

	if cfg.AppPort != "8080" {
		t.Errorf("expected default AppPort '8080', got '%s'", cfg.AppPort)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected default AppEnv 'development', got '%s'", cfg.AppEnv)
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("expected default DBHost 'localhost', got '%s'", cfg.DBHost)
	}
	if cfg.DBPort != "5432" {
		t.Errorf("expected default DBPort '5432', got '%s'", cfg.DBPort)
	}
	if cfg.JWTExpiryHours != 24 {
		t.Errorf("expected default JWTExpiryHours 24, got %d", cfg.JWTExpiryHours)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_EXPIRY_HOURS", "48")
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("JWT_EXPIRY_HOURS")
	}()

	cfg := Load()

	if cfg.AppPort != "9090" {
		t.Errorf("expected AppPort '9090', got '%s'", cfg.AppPort)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("expected AppEnv 'production', got '%s'", cfg.AppEnv)
	}
	if cfg.JWTExpiryHours != 48 {
		t.Errorf("expected JWTExpiryHours 48, got %d", cfg.JWTExpiryHours)
	}
}

func TestLoad_InvalidInt(t *testing.T) {
	os.Setenv("JWT_EXPIRY_HOURS", "not_a_number")
	defer os.Unsetenv("JWT_EXPIRY_HOURS")

	cfg := Load()

	if cfg.JWTExpiryHours != 24 {
		t.Errorf("expected fallback JWTExpiryHours 24 for invalid int, got %d", cfg.JWTExpiryHours)
	}
}

func TestDatabaseURL(t *testing.T) {
	cfg := &Config{
		DBUser:     "user",
		DBPassword: "pass",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBName:     "testdb",
		DBSSLMode:  "disable",
	}

	expected := "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
	if cfg.DatabaseURL() != expected {
		t.Errorf("expected '%s', got '%s'", expected, cfg.DatabaseURL())
	}
}
