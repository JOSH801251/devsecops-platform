package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/josh801251/devsecops-platform/services/api/internal/jobs"
)

func TestLiveOK(t *testing.T) {
	rr := httptest.NewRecorder()
	Live(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d", rr.Code)
	}
}

type okPinger struct{}

func (okPinger) PingCheck() error { return nil }

type badPinger struct{ err error }

func (b badPinger) PingCheck() error { return b.err }

func TestReady_OK_WhenPingerHealthy(t *testing.T) {
	rr := httptest.NewRecorder()
	Ready(okPinger{})(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestReady_503_WhenPingerFails(t *testing.T) {
	rr := httptest.NewRecorder()
	Ready(badPinger{err: io.EOF})(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestJobs_PostInsertsAndReturnsID(t *testing.T) {
	repo := jobs.NewFakeRepo()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jobs", strings.NewReader(`{"x":1}`))
	Jobs(logger, repo)(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"id"`) {
		t.Fatalf("missing id in %q", rr.Body.String())
	}
	if n, _ := repo.Count(req.Context(), jobs.StatusPending); n != 1 {
		t.Fatalf("expected 1 pending, got %d", n)
	}
}

func TestJobs_GetRejected(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rr := httptest.NewRecorder()
	Jobs(logger, jobs.NewFakeRepo())(rr, httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil))

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestMetrics_ExposesJobCounts(t *testing.T) {
	repo := jobs.NewFakeRepo()
	_, _ = repo.Enqueue(httptest.NewRequest(http.MethodPost, "/", nil).Context(), []byte("x"))

	rr := httptest.NewRecorder()
	Metrics(repo)(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rr.Body.String()

	for _, want := range []string{"api_requests_total", "api_jobs_pending 1", "api_jobs_done 0"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %q", want, body)
		}
	}
}

// Middleware tests

func TestWithRequestID_PreservesExistingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "existing-123")

	var receivedID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedID = r.Header.Get("X-Request-ID")
	})

	rr := httptest.NewRecorder()
	WithRequestID(next).ServeHTTP(rr, req)

	if receivedID != "existing-123" {
		t.Fatalf("expected existing request ID, got %q", receivedID)
	}

	if rr.Header().Get("X-Request-ID") != "existing-123" {
		t.Fatal("response request ID was not preserved")
	}
}

func TestWithRequestID_GeneratesMissingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	var receivedID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedID = r.Header.Get("X-Request-ID")
	})

	rr := httptest.NewRecorder()
	WithRequestID(next).ServeHTTP(rr, req)

	if len(receivedID) != 16 {
		t.Fatalf("expected 16-character request ID, got %q", receivedID)
	}

	for _, ch := range receivedID {
		if !strings.ContainsRune("0123456789abcdef", ch) {
			t.Fatalf("request ID contains invalid hexadecimal character: %q", ch)
		}
	}

	if rr.Header().Get("X-Request-ID") != receivedID {
		t.Fatal("response request ID does not match generated request ID")
	}
}

func TestWithLogging_LogsRequestAndIncrementsCounter(t *testing.T) {
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, nil))

	before := requestsTotal.Load()
	nextCalled := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "test-123")

	rr := httptest.NewRecorder()
	WithLogging(logger, next).ServeHTTP(rr, req)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}

	if requestsTotal.Load() != before+1 {
		t.Fatal("request counter was not incremented")
	}

	output := logOutput.String()

	for _, want := range []string{
		"method=GET",
		"path=/test",
		"request_id=test-123",
		"duration_ms=",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("log output missing %q: %s", want, output)
		}
	}
}
