package app

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestChiURLParamDecodesEscapedSegments(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/intelligence/assets/stock%3Amic%3AXNAS%3ATSLA", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("asset_key", "stock%3Amic%3AXNAS%3ATSLA")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	if got := chiURLParam(req, "asset_key"); got != "stock:mic:XNAS:TSLA" {
		t.Fatalf("chiURLParam() = %q, want %q", got, "stock:mic:XNAS:TSLA")
	}
}
