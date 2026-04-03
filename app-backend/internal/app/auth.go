package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
	minPasswordLen  = 8
	maxPasswordLen  = 72 // bcrypt ignores bytes after 72, so we reject longer inputs.
)

type authService struct {
	cfg   Config
	db    *gorm.DB
	redis *redisStore
}

func newAuthService(cfg Config, db *gorm.DB, redis *redisStore) *authService {
	return &authService{cfg: cfg, db: db, redis: redis}
}

type accessClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (a *authService) createSession(ctx context.Context, userID string) (string, string, error) {
	accessToken, err := a.createAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	session := AuthSession{
		ID:               newID("sess"),
		UserID:           userID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        time.Now().UTC().Add(refreshTokenTTL),
		CreatedAt:        time.Now().UTC(),
	}
	if err := a.db.WithContext(ctx).Create(&session).Error; err != nil {
		return "", "", fmt.Errorf("store refresh session: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (a *authService) createAccessToken(userID string) (string, error) {
	claims := accessClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(accessTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.JWTSigningSecret))
}

func (a *authService) verifyAccessToken(token string) (*accessClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &accessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(a.cfg.JWTSigningSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (a *authService) refreshSession(ctx context.Context, refreshToken string) (string, error) {
	refreshHash := hashRefreshToken(refreshToken)
	var session AuthSession
	if err := a.db.WithContext(ctx).Where("refresh_token_hash = ?", refreshHash).First(&session).Error; err != nil {
		return "", err
	}
	if session.RevokedAt != nil || time.Now().UTC().After(session.ExpiresAt) {
		return "", errors.New("refresh token expired")
	}
	return a.createAccessToken(session.UserID)
}

func (a *authService) revokeSession(ctx context.Context, refreshToken string) error {
	refreshHash := hashRefreshToken(refreshToken)
	now := time.Now().UTC()
	return a.db.WithContext(ctx).Model(&AuthSession{}).Where("refresh_token_hash = ?", refreshHash).Update("revoked_at", &now).Error
}

func generateRefreshToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	refresh := base64.RawURLEncoding.EncodeToString(buf)
	return refresh, hashRefreshToken(refresh), nil
}

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func hashPassword(password string) (string, error) {
	encoded, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(encoded), nil
}

func verifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func randomDigits(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("invalid length")
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, b := range buf {
		digit := int(b) % 10
		builder.WriteByte(byte('0' + digit))
	}
	return builder.String(), nil
}
