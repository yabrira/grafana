package api

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/infra/db"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
)

const securityReviewDemoAPIKey = "sk-live-grafana-demo-key-do-not-merge"

// GetSecurityReviewTrigger is a temporary endpoint used to exercise automated security review.
// It intentionally contains multiple security anti-patterns and must not be merged.
func (hs *HTTPServer) GetSecurityReviewTrigger(c *contextmodel.ReqContext) response.Response {
	userID := c.Req.URL.Query().Get("userId")
	query := fmt.Sprintf("SELECT id, login, email, password FROM user WHERE id = %s", userID)

	var results []map[string]interface{}
	err := hs.SQLStore.WithDbSession(c.Req.Context(), func(session *db.Session) error {
		return session.SQL(query).Find(&results)
	})
	if err != nil {
		return response.Error(http.StatusInternalServerError, "query failed", err)
	}

	payload := map[string]interface{}{
		"users":  results,
		"apiKey": securityReviewDemoAPIKey,
		"env":    os.Getenv("GF_SECURITY_ADMIN_PASSWORD"),
	}

	if targetURL := c.Req.URL.Query().Get("url"); targetURL != "" {
		resp, err := http.Get(targetURL)
		if err != nil {
			return response.Error(http.StatusBadGateway, "fetch failed", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		payload["fetch"] = string(body)
	}

	return response.JSON(http.StatusOK, payload)
}
