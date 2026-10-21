package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mi991203/TestMoon/backend/testing/mock/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/mock/logic"
	"github.com/mi991203/TestMoon/backend/testing/mock/model"
)

func NewPublicHandler(deps *runtime.Dependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { servePublic(deps, w, r) })
}

func servePublic(deps *runtime.Dependencies, w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 3)
	if len(parts) < 2 || parts[0] != "m" {
		http.NotFound(w, r)
		return
	}
	var endpoint model.Endpoint
	if deps.DB.First(&endpoint, "id = ?", parts[1]).Error != nil {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()
	if deps.Core.Authorize(ctx, endpoint.OrgID, endpoint.ProjectID, "active") != nil {
		http.NotFound(w, r)
		return
	}
	path := "/"
	if len(parts) == 3 {
		path += parts[2]
	}
	body, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if e != nil {
		http.Error(w, "request too large", 413)
		return
	}
	var payload any
	_ = json.Unmarshal(body, &payload)
	response, err := logic.MatchResponse(deps, endpoint, logic.MatchInput{Method: r.Method, Path: path, Headers: r.Header, Query: r.URL.Query(), Body: payload})
	if err != nil {
		http.Error(w, "unavailable", 503)
		return
	}
	if response == nil {
		http.NotFound(w, r)
		return
	}
	timer := time.NewTimer(time.Duration(response.DelayMS) * time.Millisecond)
	select {
	case <-ctx.Done():
		timer.Stop()
		return
	case <-timer.C:
	}
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(response.Status)
	if r.Method != "HEAD" && response.Status != 204 && response.Status != 304 {
		_, _ = io.WriteString(w, response.Body)
	}
}
