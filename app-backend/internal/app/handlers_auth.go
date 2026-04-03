package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
)

type authEmailRegisterStartRequest struct {
	Email string `json:"email"`
}

type authEmailRegisterRequest struct {
	Email string `json:"email"`
	// Password is plaintext over TLS and stored as bcrypt hash server-side.
	Password string `json:"password"`
	Code     string `json:"code"`
}

type authOAuthRequest struct {
	Provider string `json:"provider"`
	IDToken  string `json:"id_token"`
}

type authEmailLoginRequest struct {
	Email string `json:"email"`
	// Password is plaintext over TLS and verified against bcrypt hash.
	Password string `json:"password"`
}

type authRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authLogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type emailOTP struct {
	Hash      string    `json:"hash"`
	ExpiresAt time.Time `json:"expires_at"`
}

func normalizeEmail(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func isValidEmail(email string) bool {
	if strings.Count(email, "@") != 1 {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(addr.Address), email)
}

func validateRegisterPassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}
	if len(password) < minPasswordLen {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > maxPasswordLen {
		return errors.New("password must be at most 72 bytes")
	}
	if !utf8.ValidString(password) {
		return errors.New("password must be valid UTF-8")
	}
	return nil
}

func validateLoginPassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}
	if len(password) > maxPasswordLen {
		return errors.New("password must be at most 72 bytes")
	}
	if !utf8.ValidString(password) {
		return errors.New("password must be valid UTF-8")
	}
	return nil
}

func emailRegisterOTPKey(email string) string {
	return "auth:email:register:" + email
}

func shouldReturnEmailOTP(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "debug", "local", "test":
		return true
	default:
		return false
	}
}

func hashOTP(code string) string {
	value := sha256.Sum256([]byte(code))
	return base64.RawURLEncoding.EncodeToString(value[:])
}

func (s *Server) loadEmailRegisterOTP(ctx context.Context, email string) (emailOTP, error) {
	var otp emailOTP
	found, err := s.redis.getJSON(ctx, emailRegisterOTPKey(email), &otp)
	if err != nil || !found {
		if err == nil {
			err = errors.New("otp not found")
		}
		return emailOTP{}, err
	}
	return otp, nil
}

func (s *Server) handleAuthOAuth(w http.ResponseWriter, r *http.Request) {
	var req authOAuthRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	provider := strings.TrimSpace(strings.ToLower(req.Provider))
	if provider == "" || strings.TrimSpace(req.IDToken) == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "provider and id_token are required", nil)
		return
	}

	switch provider {
	case "google":
		claims, err := s.verifyGoogleIDToken(r.Context(), req.IDToken)
		if err != nil {
			if errors.Is(err, errGoogleAuthNotConfigured) {
				s.writeError(w, http.StatusInternalServerError, "CONFIG_ERROR", "google oauth not configured", nil)
				return
			}
			s.writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid google token", nil)
			return
		}

		var email *string
		if claims.EmailVerified {
			trimmed := strings.TrimSpace(strings.ToLower(claims.Email))
			if trimmed != "" {
				email = &trimmed
			}
		}

		user, err := s.upsertAuthIdentity(r.Context(), "google", claims.Subject, email)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create session", nil)
			return
		}

		accessToken, refreshToken, err := s.auth.createSession(r.Context(), user.ID)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create session", nil)
			return
		}

		s.writeJSON(w, http.StatusOK, authResponse{AccessToken: accessToken, RefreshToken: refreshToken, UserID: user.ID})
	case "apple":
		claims, err := s.verifyAppleIDToken(r.Context(), req.IDToken)
		if err != nil {
			if errors.Is(err, errAppleAuthNotConfigured) {
				s.writeError(w, http.StatusInternalServerError, "CONFIG_ERROR", "apple oauth not configured", nil)
				return
			}
			s.writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid apple token", nil)
			return
		}

		var email *string
		if claims.EmailVerified {
			trimmed := strings.TrimSpace(strings.ToLower(claims.Email))
			if trimmed != "" {
				email = &trimmed
			}
		}

		user, err := s.upsertAuthIdentity(r.Context(), "apple", claims.Subject, email)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create session", nil)
			return
		}

		accessToken, refreshToken, err := s.auth.createSession(r.Context(), user.ID)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create session", nil)
			return
		}

		s.writeJSON(w, http.StatusOK, authResponse{AccessToken: accessToken, RefreshToken: refreshToken, UserID: user.ID})
	default:
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "provider must be google or apple", nil)
	}
}

func (s *Server) handleAuthEmailRegisterStart(w http.ResponseWriter, r *http.Request) {
	var req authEmailRegisterStartRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	email := normalizeEmail(req.Email)
	if !isValidEmail(email) {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "valid email is required", nil)
		return
	}

	var existing AuthIdentity
	if err := s.db.DB().WithContext(r.Context()).Where("provider = ? AND provider_user_id = ?", "email", email).First(&existing).Error; err == nil {
		s.writeError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "email already registered", nil)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to start registration", nil)
		return
	}

	code, err := randomDigits(6)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "OTP_ERROR", "failed to generate code", nil)
		return
	}
	otp := emailOTP{
		Hash:      hashOTP(code),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
	}
	if err := s.redis.setJSON(r.Context(), emailRegisterOTPKey(email), otp, 10*time.Minute); err != nil {
		s.writeError(w, http.StatusInternalServerError, "OTP_ERROR", "failed to store code", nil)
		return
	}
	if shouldReturnEmailOTP(s.cfg.EmailOTPMode) {
		s.writeJSON(w, http.StatusOK, map[string]any{"sent": true, "code": code})
		return
	}
	if err := s.mailer.sendVerificationCode(r.Context(), email, code); err != nil {
		s.writeError(w, http.StatusInternalServerError, "EMAIL_ERROR", "failed to send email", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"sent": true})
}

func (s *Server) handleAuthEmailRegister(w http.ResponseWriter, r *http.Request) {
	var req authEmailRegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	email := normalizeEmail(req.Email)
	if !isValidEmail(email) {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "valid email is required", nil)
		return
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "code is required", nil)
		return
	}
	if err := validateRegisterPassword(req.Password); err != nil {
		s.writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error(), nil)
		return
	}

	otp, err := s.loadEmailRegisterOTP(r.Context(), email)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "INVALID_CODE", "invalid or expired code", nil)
		return
	}
	if time.Now().UTC().After(otp.ExpiresAt) || otp.Hash != hashOTP(code) {
		s.writeError(w, http.StatusUnauthorized, "INVALID_CODE", "invalid or expired code", nil)
		return
	}

	var existing AuthIdentity
	if err := s.db.DB().WithContext(r.Context()).Where("provider = ? AND provider_user_id = ?", "email", email).First(&existing).Error; err == nil {
		s.writeError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "email already registered", nil)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create account", nil)
		return
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create account", nil)
		return
	}

	user, err := s.createEmailAuthUser(r.Context(), email, passwordHash)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create account", nil)
		return
	}
	_ = s.redis.del(r.Context(), emailRegisterOTPKey(email))

	accessToken, refreshToken, err := s.auth.createSession(r.Context(), user.ID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create session", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, authResponse{AccessToken: accessToken, RefreshToken: refreshToken, UserID: user.ID})
}

func (s *Server) handleAuthEmailLogin(w http.ResponseWriter, r *http.Request) {
	var req authEmailLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	email := normalizeEmail(req.Email)
	if !isValidEmail(email) {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "valid email is required", nil)
		return
	}
	if err := validateLoginPassword(req.Password); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	var identity AuthIdentity
	err := s.db.DB().WithContext(r.Context()).Where("provider = ? AND provider_user_id = ?", "email", email).First(&identity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password", nil)
			return
		}
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to authenticate", nil)
		return
	}
	if identity.PasswordHash == nil || !verifyPassword(req.Password, *identity.PasswordHash) {
		s.writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password", nil)
		return
	}

	var user User
	if err := s.db.DB().WithContext(r.Context()).First(&user, "id = ?", identity.UserID).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to authenticate", nil)
		return
	}
	if user.Email == nil {
		emailCopy := email
		_ = s.db.DB().WithContext(r.Context()).Model(&user).Updates(map[string]any{
			"email":      &emailCopy,
			"updated_at": time.Now().UTC(),
		}).Error
	}

	accessToken, refreshToken, err := s.auth.createSession(r.Context(), user.ID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to create session", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, authResponse{AccessToken: accessToken, RefreshToken: refreshToken, UserID: user.ID})
}

func (s *Server) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	var req authRefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "refresh_token is required", nil)
		return
	}
	accessToken, err := s.auth.refreshSession(r.Context(), req.RefreshToken)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "refresh token invalid", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"access_token": accessToken})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	var req authLogoutRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "refresh_token is required", nil)
		return
	}
	if err := s.auth.revokeSession(r.Context(), req.RefreshToken); err != nil {
		s.writeError(w, http.StatusInternalServerError, "AUTH_ERROR", "failed to revoke session", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"revoked": true})
}

func (s *Server) upsertAuthIdentity(ctx context.Context, provider, providerUserID string, email *string) (User, error) {
	var identity AuthIdentity
	if err := s.db.DB().WithContext(ctx).Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&identity).Error; err == nil {
		var user User
		if err := s.db.DB().WithContext(ctx).First(&user, "id = ?", identity.UserID).Error; err != nil {
			return User{}, err
		}
		if email != nil && user.Email == nil {
			_ = s.db.DB().WithContext(ctx).Model(&user).Update("email", email).Error
		}
		return user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, err
	}

	user := User{
		ID:        newID("usr"),
		Email:     email,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	identity = AuthIdentity{
		ID:             newID("aid"),
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		Email:          email,
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Create(&identity).Error; err != nil {
			return err
		}
		return s.ensureUserProfileTx(ctx, tx, user.ID)
	}); err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *Server) createEmailAuthUser(ctx context.Context, email, passwordHash string) (User, error) {
	emailCopy := email
	passwordHashCopy := passwordHash
	user := User{
		ID:        newID("usr"),
		Email:     &emailCopy,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	identity := AuthIdentity{
		ID:             newID("aid"),
		UserID:         user.ID,
		Provider:       "email",
		ProviderUserID: email,
		Email:          &emailCopy,
		PasswordHash:   &passwordHashCopy,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Create(&identity).Error; err != nil {
			return err
		}
		return s.ensureUserProfileTx(ctx, tx, user.ID)
	}); err != nil {
		return User{}, err
	}
	return user, nil
}
