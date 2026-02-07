package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"gamba/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountInactive    = errors.New("account is deactivated")
	ErrAccountRestricted  = errors.New("account is restricted")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenRevoked       = errors.New("token has been revoked")
)

func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:              db,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  15 * time.Minute,
		refreshTokenTTL: 7 * 24 * time.Hour,
	}
}

func (s *AuthService) Register(username, password string) (*models.User, error) {
	var existing models.User
	err := s.db.Where("username = ?", username).First(&existing).Error
	if err == nil {
		return nil, ErrUsernameTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         models.RolePlayer,
		Balance:      0,
		IsActive:     true,
		IsRestricted: false,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(username, password string) (*models.User, *TokenPair, error) {
	var user models.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, nil, ErrAccountInactive
	}

	if user.IsRestricted {
		return nil, nil, ErrAccountRestricted
	}

	tokens, err := s.generateTokenPair(&user)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (s *AuthService) Refresh(rawRefreshToken string) (*TokenPair, error) {
	tokenHash := hashToken(rawRefreshToken)

	var storedToken models.RefreshToken
	err := s.db.Where("token_hash = ? AND is_revoked = ?", tokenHash, false).First(&storedToken).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	now := time.Now()
	s.db.Model(&storedToken).Updates(map[string]interface{}{
		"is_revoked": true,
		"revoked_at": &now,
	})

	var user models.User
	if err := s.db.First(&user, "id = ?", storedToken.UserID).Error; err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	tokens, err := s.generateTokenPair(&user)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Logout(rawRefreshToken string) error {
	tokenHash := hashToken(rawRefreshToken)

	now := time.Now()
	result := s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND is_revoked = ?", tokenHash, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": &now,
		})

	if result.RowsAffected == 0 {
		return ErrInvalidToken
	}

	return result.Error
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// --- private helpers ---

func (s *AuthService) generateTokenPair(user *models.User) (*TokenPair, error) {
	// Access token
	now := time.Now()
	accessClaims := AccessTokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh token
	rawRefreshToken, err := generateRandomToken()
	if err != nil {
		return nil, err
	}

	tokenHash := hashToken(rawRefreshToken)

	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(s.refreshTokenTTL),
		IsRevoked: false,
	}

	if err := s.db.Create(&refreshToken).Error; err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
	}, nil
}

func generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
