package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestGotCpuRoute(t *testing.T) {
// 	r := SetupRouter()

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/gotcpu", nil)
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	// assert.Equal(t, "pong", w.Body.String())
// }

func sendFileRequest(filePath string) *http.Request {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileField, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(fileField, file)
	if err != nil {
		log.Fatal(err)
	}

	writer.Close()

	req, _ := http.NewRequest("POST", "/api/runAFLPlusPlus", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestRunAFLPlusPlusCrashRoute(t *testing.T) {
	t.Parallel()
	r := SetupRouter()

	req := sendFileRequest("c_program/readfile/src/readfile.c")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"message\":\"readfile.c : Crash Found\"}", w.Body.String())
}

func TestRunAFLPlusPlusNotCrashRoute(t *testing.T) {
	t.Parallel()
	r := SetupRouter()

	req := sendFileRequest("c_program/hello/src/hello.c")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"message\":\"hello.c : Crash Not Found\"}", w.Body.String())
}
