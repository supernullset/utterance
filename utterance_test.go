package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestServer(t *testing.T) {
	// Create a mock AWS service and upload
	server := NewServer()
	server.S3 = s3.New(mock.Session, aws.NewConfig().WithRegion("us-west-2"))

	w := httptest.NewRecorder()

	// GET should return the form
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("failed to create GET: %s", err)
	}

	server.uploadHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(
			"unexpected status from GET (%d): %s",
			w.Code, w.Body.String(),
		)
	}
	// TODO test the body content

	// POST should upload a file
	w = httptest.NewRecorder()

	var content bytes.Buffer
	multi := multipart.NewWriter(&content)
	fw, err := multi.CreateFormFile("file", "plain.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %s", err)
	}
	fw.Write([]byte("hello"))

	req, err = http.NewRequest(http.MethodPost, "/", &content)
	if err != nil {
		t.Fatalf("failed to create POST: %s", err)
	}
	req.Header.Set("Content-Type", multi.FormDataContentType())
	multi.Close()

	server.uploadHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(
			"unexpected status from POST (%d): %s",
			w.Code, w.Body.String(),
		)
	}
	// test the mock s3?
}
