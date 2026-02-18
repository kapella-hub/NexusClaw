package respond_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kapella-hub/NexusClaw/internal/platform/respond"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	payload := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{Name: "Alice", Age: 30}

	respond.JSON(rec, http.StatusOK, payload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var got map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got["name"] != "Alice" {
		t.Errorf("expected name Alice, got %v", got["name"])
	}
	if got["age"] != float64(30) {
		t.Errorf("expected age 30, got %v", got["age"])
	}
}

func TestJSONNilPayload(t *testing.T) {
	rec := httptest.NewRecorder()
	respond.JSON(rec, http.StatusNoContent, nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for nil payload, got %q", rec.Body.String())
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	respond.Error(rec, http.StatusBadRequest, "something went wrong")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if got["error"] != "something went wrong" {
		t.Errorf("expected error message 'something went wrong', got %q", got["error"])
	}
}

func TestDecodeValid(t *testing.T) {
	body := `{"name":"Bob"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	var v struct {
		Name string `json:"name"`
	}

	ok := respond.Decode(rec, req, &v)
	if !ok {
		t.Fatal("expected Decode to return true for valid JSON")
	}
	if v.Name != "Bob" {
		t.Errorf("expected name Bob, got %q", v.Name)
	}
}

func TestDecodeInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	var v struct {
		Name string `json:"name"`
	}

	ok := respond.Decode(rec, req, &v)
	if ok {
		t.Fatal("expected Decode to return false for invalid JSON")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
