package limen

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: &Config{
				Database: newTestMemoryAdapter(t),
				Secret:   testSecret,
			},
			wantErr: false,
		},
		{
			name:    "nil configuration",
			config:  nil,
			wantErr: true,
			errMsg:  "missing configuration",
		},
		{
			name: "configuration with nil adapter",
			config: &Config{
				Database: nil,
			},
			wantErr: true,
			errMsg:  "invalid configuration: database adapter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limen, err := New(tt.config)

			if tt.wantErr {
				assert.Error(t, err, "New() expected error but got none")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
				assert.Nil(t, limen, "Expected nil Limen instance when error occurs")
			} else {
				assert.NoError(t, err, "New() unexpected error")
				assert.NotNil(t, limen, "Expected Limen instance but got nil")
			}
		})
	}
}

func TestNewUserBuildsUserFromColumnMap(t *testing.T) {
	t.Parallel()

	l := newTestLimen(t)
	emailField := l.core.Schema.User.GetEmailField()

	user := l.NewUser(map[string]any{
		emailField: "newuser@example.com",
		"custom":   "value",
	})
	require.NotNil(t, user)
	assert.Equal(t, "newuser@example.com", user.Email)
	require.NotNil(t, user.Raw())
	assert.Equal(t, "newuser@example.com", user.Raw()[emailField])
	assert.Equal(t, "value", user.Raw()["custom"])
}

// CreateSession is the programmatic hook AlgaePool uses to issue sessions
// inside its own Gin handlers (AUTH-05). The delegation must return a cookie
// that follows the configured cookie policy without exposing LimenCore.
func TestLimen_CreateSessionDelegation(t *testing.T) {
	t.Parallel()

	l := newTestLimen(t)
	user := SeedTestUser(t, l, "session-delegation@example.com")

	req := httptest.NewRequest("POST", "/api/v1/auth/signin/credential", nil)
	result, err := l.CreateSession(context.Background(), req, &AuthenticationResult{User: user})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Token)
	require.NotNil(t, result.Cookie)
	assert.Equal(t, "limen_session", result.Cookie.Name)
	assert.True(t, result.Cookie.HttpOnly)
}
