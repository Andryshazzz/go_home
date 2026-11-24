package config

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
)

const minJWTSecretLen = 32

var (
	allowedLogLevels = []string{"debug", "info", "warn", "error"}
	allowedSSLModes  = []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}
)

func (c *Config) Validate() error {
	var errs []error

	add := func(key, env, format string, args ...any) {
		errs = append(errs, fmt.Errorf("%s (env %s): %s", key, env, fmt.Sprintf(format, args...)))
	}

	// HTTP
	if err := validateAddr(c.HTTP.Addr); err != nil {
		add("http.addr", "HTTP_ADDR", "%v", err)
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		add("http.shutdown_timeout", "HTTP_SHUTDOWN_TIMEOUT", "must be greater than zero (got %s)", c.HTTP.ShutdownTimeout)
	}

	// PostgreSQL
	if strings.TrimSpace(c.Postgres.Host) == "" {
		add("postgres.host", "POSTGRES_HOST", "is required")
	}
	if err := validatePort(c.Postgres.Port); err != nil {
		add("postgres.port", "POSTGRES_PORT", "%v", err)
	}
	if strings.TrimSpace(c.Postgres.User) == "" {
		add("postgres.user", "POSTGRES_USER", "is required")
	}
	if c.Postgres.Password == "" {
		add("postgres.password", "POSTGRES_PASSWORD", "is required")
	}
	if strings.TrimSpace(c.Postgres.DB) == "" {
		add("postgres.db", "POSTGRES_DB", "is required")
	}
	if !slices.Contains(allowedSSLModes, c.Postgres.SSLMode) {
		add("postgres.sslmode", "POSTGRES_SSLMODE", "must be one of %s (got %q)",
			strings.Join(allowedSSLModes, ", "), c.Postgres.SSLMode)
	}
	if c.Postgres.Timeout <= 0 {
		add("postgres.timeout", "POSTGRES_TIMEOUT", "must be greater than zero (got %s)", c.Postgres.Timeout)
	}

	// Logger
	if !slices.Contains(allowedLogLevels, strings.ToLower(c.Logger.Level)) {
		add("logger.level", "LOGGER_LEVEL", "must be one of %s (got %q)",
			strings.Join(allowedLogLevels, ", "), c.Logger.Level)
	}
	if strings.TrimSpace(c.Logger.Folder) == "" {
		add("logger.folder", "LOGGER_FOLDER", "is required")
	}

	// JWT
	if len(c.JWT.Secret) < minJWTSecretLen {
		add("jwt.secret", "JWT_SECRET", "must be at least %d bytes (got %d)", minJWTSecretLen, len(c.JWT.Secret))
	}
	if c.JWT.AccessExpiration <= 0 {
		add("jwt.access_expiration", "JWT_ACCESS_EXPIRATION", "must be greater than zero (got %s)", c.JWT.AccessExpiration)
	}
	if c.JWT.RefreshExpiration <= 0 {
		add("jwt.refresh_expiration", "JWT_REFRESH_EXPIRATION", "must be greater than zero (got %s)", c.JWT.RefreshExpiration)
	}
	if c.JWT.AccessExpiration > 0 && c.JWT.RefreshExpiration > 0 &&
		c.JWT.AccessExpiration >= c.JWT.RefreshExpiration {
		errs = append(errs, fmt.Errorf(
			"jwt: access_expiration (%s) must be shorter than refresh_expiration (%s)",
			c.JWT.AccessExpiration, c.JWT.RefreshExpiration,
		))
	}

	return errors.Join(errs...)
}

func validateAddr(addr string) error {
	if strings.TrimSpace(addr) == "" {
		return errors.New("is required")
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf(`must be in host:port form, e.g. ":8080" (%v)`, err)
	}

	return validatePort(port)
}

func validatePort(port string) error {
	if port == "" {
		return errors.New("is required")
	}

	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("must be a number between 1 and 65535 (got %q)", port)
	}

	return nil
}
