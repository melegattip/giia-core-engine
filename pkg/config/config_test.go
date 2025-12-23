package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should create config successfully", func(t *testing.T) {
		cfg, err := New("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("should create config with prefix", func(t *testing.T) {
		cfg, err := New("APP")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("should handle missing config file gracefully", func(t *testing.T) {
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		tempDir := t.TempDir()
		os.Chdir(tempDir)

		cfg, err := New("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})
}

func TestViperConfig_GetString(t *testing.T) {
	t.Run("should get string from environment variable", func(t *testing.T) {
		os.Setenv("TEST_STRING_VAR", "test_value")
		defer os.Unsetenv("TEST_STRING_VAR")

		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetString("TEST_STRING_VAR")
		assert.Equal(t, "test_value", value)
	})

	t.Run("should return empty string for non-existent key", func(t *testing.T) {
		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetString("NON_EXISTENT_KEY")
		assert.Equal(t, "", value)
	})
}

func TestViperConfig_GetInt(t *testing.T) {
	t.Run("should get integer from environment variable", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR", "42")
		defer os.Unsetenv("TEST_INT_VAR")

		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetInt("TEST_INT_VAR")
		assert.Equal(t, 42, value)
	})

	t.Run("should return zero for non-existent key", func(t *testing.T) {
		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetInt("NON_EXISTENT_INT")
		assert.Equal(t, 0, value)
	})

	t.Run("should return zero for invalid integer", func(t *testing.T) {
		os.Setenv("TEST_INVALID_INT", "not_a_number")
		defer os.Unsetenv("TEST_INVALID_INT")

		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetInt("TEST_INVALID_INT")
		assert.Equal(t, 0, value)
	})
}

func TestViperConfig_GetBool(t *testing.T) {
	t.Run("should get boolean from environment variable", func(t *testing.T) {
		os.Setenv("TEST_BOOL_VAR", "true")
		defer os.Unsetenv("TEST_BOOL_VAR")

		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetBool("TEST_BOOL_VAR")
		assert.True(t, value)
	})

	t.Run("should return false for non-existent key", func(t *testing.T) {
		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetBool("NON_EXISTENT_BOOL")
		assert.False(t, value)
	})

	t.Run("should handle various boolean formats", func(t *testing.T) {
		testCases := []struct {
			name     string
			envValue string
			expected bool
		}{
			{"true lowercase", "true", true},
			{"True capitalized", "True", true},
			{"1 numeric", "1", true},
			{"false lowercase", "false", false},
			{"0 numeric", "0", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				os.Setenv("TEST_BOOL_FORMAT", tc.envValue)
				defer os.Unsetenv("TEST_BOOL_FORMAT")

				cfg, err := New("")
				require.NoError(t, err)

				value := cfg.GetBool("TEST_BOOL_FORMAT")
				assert.Equal(t, tc.expected, value)
			})
		}
	})
}

func TestViperConfig_GetFloat64(t *testing.T) {
	t.Run("should get float from environment variable", func(t *testing.T) {
		os.Setenv("TEST_FLOAT_VAR", "3.14")
		defer os.Unsetenv("TEST_FLOAT_VAR")

		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetFloat64("TEST_FLOAT_VAR")
		assert.Equal(t, 3.14, value)
	})

	t.Run("should return zero for non-existent key", func(t *testing.T) {
		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.GetFloat64("NON_EXISTENT_FLOAT")
		assert.Equal(t, 0.0, value)
	})
}

func TestViperConfig_Get(t *testing.T) {
	t.Run("should get interface value", func(t *testing.T) {
		os.Setenv("TEST_INTERFACE_VAR", "interface_value")
		defer os.Unsetenv("TEST_INTERFACE_VAR")

		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.Get("TEST_INTERFACE_VAR")
		assert.NotNil(t, value)
		assert.Equal(t, "interface_value", value)
	})

	t.Run("should return nil for non-existent key", func(t *testing.T) {
		cfg, err := New("")
		require.NoError(t, err)

		value := cfg.Get("NON_EXISTENT_INTERFACE")
		assert.Nil(t, value)
	})
}

func TestViperConfig_IsSet(t *testing.T) {
	t.Run("should return true for set variable", func(t *testing.T) {
		os.Setenv("TEST_SET_VAR", "value")
		defer os.Unsetenv("TEST_SET_VAR")

		cfg, err := New("")
		require.NoError(t, err)

		isSet := cfg.IsSet("TEST_SET_VAR")
		assert.True(t, isSet)
	})

	t.Run("should return false for unset variable", func(t *testing.T) {
		cfg, err := New("")
		require.NoError(t, err)

		isSet := cfg.IsSet("UNSET_VARIABLE")
		assert.False(t, isSet)
	})
}

func TestViperConfig_Validate(t *testing.T) {
	t.Run("should pass validation when all required keys are set", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://localhost/db")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		defer os.Unsetenv("DATABASE_URL")
		defer os.Unsetenv("REDIS_URL")

		cfg, err := New("")
		require.NoError(t, err)

		requiredKeys := []string{"DATABASE_URL", "REDIS_URL"}
		err = cfg.Validate(requiredKeys)
		assert.NoError(t, err)
	})

	t.Run("should fail validation when required key is missing", func(t *testing.T) {
		os.Unsetenv("REQUIRED_KEY_MISSING")

		cfg, err := New("")
		require.NoError(t, err)

		requiredKeys := []string{"REQUIRED_KEY_MISSING"}
		err = cfg.Validate(requiredKeys)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required configuration keys")
		assert.Contains(t, err.Error(), "REQUIRED_KEY_MISSING")
	})

	t.Run("should fail validation when multiple required keys are missing", func(t *testing.T) {
		os.Unsetenv("MISSING_KEY_1")
		os.Unsetenv("MISSING_KEY_2")
		os.Unsetenv("MISSING_KEY_3")

		cfg, err := New("")
		require.NoError(t, err)

		requiredKeys := []string{"MISSING_KEY_1", "MISSING_KEY_2", "MISSING_KEY_3"}
		err = cfg.Validate(requiredKeys)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MISSING_KEY_1")
		assert.Contains(t, err.Error(), "MISSING_KEY_2")
		assert.Contains(t, err.Error(), "MISSING_KEY_3")
	})

	t.Run("should pass validation with empty required keys list", func(t *testing.T) {
		cfg, err := New("")
		require.NoError(t, err)

		err = cfg.Validate([]string{})
		assert.NoError(t, err)
	})
}

func TestViperConfig_WithPrefix(t *testing.T) {
	t.Run("should read config with prefix", func(t *testing.T) {
		os.Setenv("APP_DATABASE_URL", "postgres://localhost/db")
		defer os.Unsetenv("APP_DATABASE_URL")

		cfg, err := New("APP")
		require.NoError(t, err)

		value := cfg.GetString("database.url")
		assert.Equal(t, "postgres://localhost/db", value)
	})
}

func TestViperConfig_DotNotation(t *testing.T) {
	t.Run("should handle dot notation in keys", func(t *testing.T) {
		os.Setenv("DATABASE_HOST", "localhost")
		os.Setenv("DATABASE_PORT", "5432")
		defer os.Unsetenv("DATABASE_HOST")
		defer os.Unsetenv("DATABASE_PORT")

		cfg, err := New("")
		require.NoError(t, err)

		host := cfg.GetString("DATABASE_HOST")
		port := cfg.GetInt("DATABASE_PORT")

		assert.Equal(t, "localhost", host)
		assert.Equal(t, 5432, port)
	})
}
