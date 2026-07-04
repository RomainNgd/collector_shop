package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMiddlewareExportsTemplatedHTTPMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Middleware())
	router.GET("/products/:id", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodGet, "/products/42", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	metricsResponse := httptest.NewRecorder()
	Handler().ServeHTTP(metricsResponse, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body, err := io.ReadAll(metricsResponse.Body)
	if err != nil {
		t.Fatalf("failed to read metrics response: %v", err)
	}
	output := string(body)

	if !strings.Contains(output, `collector_http_requests_total{method="GET",route="/products/:id",status="201"} 1`) {
		t.Fatalf("expected request counter with route template, got:\n%s", output)
	}
	if !strings.Contains(output, `collector_http_request_duration_seconds_count{method="GET",route="/products/:id",status="201"} 1`) {
		t.Fatalf("expected request duration histogram, got:\n%s", output)
	}
	if strings.Contains(output, "/products/42") {
		t.Fatal("raw request path must not be used as a Prometheus label")
	}
}
