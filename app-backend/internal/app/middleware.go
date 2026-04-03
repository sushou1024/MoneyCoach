package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const (
	contextUserID contextKey = "user_id"
)

func (s *Server) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = newID("req")
		}
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)
		s.logger.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, wrapped.status, duration)
	})
}

func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unexpected error", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get("Authorization")
		if value == "" {
			next.ServeHTTP(w, r)
			return
		}
		parts := strings.Fields(value)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			next.ServeHTTP(w, r)
			return
		}
		claims, err := s.auth.verifyAccessToken(parts[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), contextUserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := userIDFromContext(r.Context())
		if userID == "" {
			s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid token", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withRateLimit(readLimit, writeLimit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bucket := "read"
			limit := readLimit
			if r.Method != http.MethodGet {
				bucket = "write"
				limit = writeLimit
			}
			if limit > 0 {
				key := s.rateKey(r, bucket)
				count, err := s.redis.incrRate(r.Context(), key, time.Minute)
				if err != nil {
					s.writeError(w, http.StatusInternalServerError, "RATE_LIMIT_ERROR", "rate limiter error", nil)
					return
				}
				if count > int64(limit) {
					s.writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded", nil)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) withIdempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "INVALID_BODY", "unable to read request", nil)
			return
		}
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))
		hash := sha256.Sum256(body)
		hashHex := hex.EncodeToString(hash[:])

		idemKey := s.idempotencyKey(r, key)
		record, err := s.redis.getIdempotency(r.Context(), idemKey)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "IDEMPOTENCY_ERROR", "idempotency lookup failed", nil)
			return
		}
		if record != nil {
			if record.Hash != hashHex {
				s.writeError(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "idempotency key reused with different payload", nil)
				return
			}
			for k, v := range record.Headers {
				w.Header().Set(k, v)
			}
			w.Header().Set("Idempotent-Replay", "true")
			w.WriteHeader(record.StatusCode)
			_, _ = w.Write(record.Body)
			return
		}

		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		if recorder.status >= 200 && recorder.status < 300 {
			record := idempotencyRecord{
				Hash:       hashHex,
				StatusCode: recorder.status,
				Body:       recorder.body.Bytes(),
				Headers:    map[string]string{"Content-Type": recorder.Header().Get("Content-Type")},
			}
			_ = s.redis.setIdempotency(r.Context(), idemKey, record, 24*time.Hour)
		}
	})
}

func (s *Server) idempotencyKey(r *http.Request, key string) string {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		userID = clientIP(r)
	}
	pattern := routePattern(r)
	return "idem:" + userID + ":" + pattern + ":" + key
}

func (s *Server) rateKey(r *http.Request, bucket string) string {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		userID = clientIP(r)
	}
	pattern := routePattern(r)
	if bucket == "" {
		bucket = "read"
	}
	return "rate:" + bucket + ":" + userID + ":" + pattern
}

func routePattern(r *http.Request) string {
	if chiCtx := chi.RouteContext(r.Context()); chiCtx != nil {
		if pattern := chiCtx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}

func userIDFromContext(ctx context.Context) string {
	value := ctx.Value(contextUserID)
	if value == nil {
		return ""
	}
	userID, _ := value.(string)
	return userID
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	r.body.Write(p)
	return r.ResponseWriter.Write(p)
}

func (r *responseRecorder) JSON(data any) {
	encoded, _ := json.Marshal(data)
	_, _ = r.Write(encoded)
}
