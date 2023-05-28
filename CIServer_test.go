package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGotCpuRoute(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	r := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/gotcpu", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// assert.Equal(t, "pong", w.Body.String())
}

func TestRunAFLPlusPlusCrashRoute(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	r := SetupRouter()

	data := url.Values{"target": {"readfile"}}
	body := strings.NewReader(data.Encode())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/runAFLPlusPlus", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"message\":\"Crash Found\"}", w.Body.String())
}

func TestRunAFLPlusPlusNotCrashRoute(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	r := SetupRouter()

	data := url.Values{"target": {"hello"}}
	body := strings.NewReader(data.Encode())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/runAFLPlusPlus", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"message\":\"Crash Not Found\"}", w.Body.String())
}
