package app

import (
	"net/http"
	"strings"
)

func (s *Server) handleLocalUpload(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ObjectStorageMode != "local" {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "local uploads not enabled", nil)
		return
	}
	key := strings.TrimPrefix(chiURLParam(r, "*"), "/")
	if strings.TrimSpace(key) == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "missing storage key", nil)
		return
	}
	storage, ok := s.storage.(*localStorageClient)
	if !ok {
		s.writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "local storage unavailable", nil)
		return
	}
	if r.Body == nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "missing body", nil)
		return
	}
	defer r.Body.Close()
	if err := storage.putObject(r.Context(), key, r.Body); err != nil {
		s.writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "failed to store upload", nil)
		return
	}
	w.WriteHeader(http.StatusOK)
}
