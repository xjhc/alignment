package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

// DiscordUser represents a Discord user from the OAuth response
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
}

// Claims represents JWT claims for authenticated users
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Provider string `json:"provider"`
	jwt.RegisteredClaims
}

// AuthService handles OAuth2 and JWT authentication
type AuthService struct {
	oauthConfig *oauth2.Config
	jwtSecret   []byte
}

// NewAuthService creates a new authentication service
func NewAuthService() (*AuthService, error) {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	clientSecret := os.Getenv("DISCORD_CLIENT_SECRET")
	redirectURL := os.Getenv("DISCORD_REDIRECT_URL")
	
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("Discord OAuth credentials not configured")
	}
	
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/api/auth/callback"
	}

	// Generate or get JWT secret
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		// Generate a random secret if not provided
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		jwtSecret = secret
	}

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"identify", "email"},
		Endpoint:     endpoints.Discord,
	}

	return &AuthService{
		oauthConfig: oauthConfig,
		jwtSecret:   jwtSecret,
	}, nil
}

// GetAuthURL generates an OAuth2 authorization URL with state
func (as *AuthService) GetAuthURL() (string, string, error) {
	// Generate random state for CSRF protection
	state := make([]byte, 16)
	if _, err := rand.Read(state); err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	stateStr := hex.EncodeToString(state)
	
	authURL := as.oauthConfig.AuthCodeURL(stateStr, oauth2.AccessTypeOffline)
	return authURL, stateStr, nil
}

// ExchangeCodeForToken exchanges authorization code for access token and user info
func (as *AuthService) ExchangeCodeForToken(ctx context.Context, code string) (*DiscordUser, error) {
	token, err := as.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from Discord API
	client := as.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Discord API returned status %d", resp.StatusCode)
	}

	var user DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &user, nil
}

// GenerateJWT creates a JWT token for the authenticated user
func (as *AuthService) GenerateJWT(user *DiscordUser) (string, error) {
	// Create JWT claims
	claims := Claims{
		UserID:   fmt.Sprintf("discord:%s", user.ID),
		Username: user.Username,
		Avatar:   user.Avatar,
		Provider: "discord",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "alignment-server",
			Subject:   user.ID,
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign the token with our secret
	tokenString, err := token.SignedString(as.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates and parses a JWT token
func (as *AuthService) ValidateJWT(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return as.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Extract and validate claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid JWT token")
}

// GetUserFromJWT extracts user information from a JWT token string
func (as *AuthService) GetUserFromJWT(tokenString string) (*Claims, error) {
	return as.ValidateJWT(tokenString)
}