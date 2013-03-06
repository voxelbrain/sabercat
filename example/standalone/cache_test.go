package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	DATA_LENGTH = 8 * (1 << 10)
	PAYLOAD     = "data"
)

var (
	data = make([]byte, DATA_LENGTH)
)

func TestCaching(t *testing.T) {
	requestCount := 0
	incrementHandler := func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(PAYLOAD)))
		w.Write([]byte(PAYLOAD))
	}

	handler := Cache(5*time.Hour, http.HandlerFunc(incrementHandler))
	server := httptest.NewServer(handler)
	defer server.Close()
	buf := make([]byte, len(PAYLOAD))

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Could not send request: %s", err)
	}

	n, err := io.ReadFull(resp.Body, buf)
	t.Logf(resp.Header.Get("Content-Length"))
	if err != nil {
		t.Fatalf("Reading from body failed: %s", err)
	}
	if n != len(PAYLOAD) {
		t.Fatalf("Unexpected length of body. Expected %d, got %d", len(PAYLOAD), n)
	}

	resp, err = http.Get(server.URL)
	if err != nil {
		t.Fatalf("Could not send request: %s", err)
	}
	n, err = io.ReadFull(resp.Body, buf)
	if err != nil {
		t.Fatalf("Reading from body failed: %s", err)
	}
	if n != len(PAYLOAD) {
		t.Fatalf("Unexpected length of body. Expected %d, got %d", len(PAYLOAD), n)
	}

	if requestCount != 1 {
		t.Fatalf("Handler was requested %d times, expected 1", requestCount)
	}
}

// Writes 8 kbytes of data to the body
func CompleteDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data)
}

func TestFullCachingOnly(t *testing.T) {
	handler := Cache(5*time.Hour, http.HandlerFunc(CompleteDataHandler))
	server := httptest.NewServer(handler)
	defer server.Close()
	buf := make([]byte, 4)

	// First requests aborts after a few bytes
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Could not send request: %s", err)
	}
	n, err := io.ReadFull(resp.Body, buf)
	if err != nil {
		t.Fatalf("Reading from body failed: %s", err)
	}
	if n != len(buf) {
		t.Fatalf("Unexpected length of body. Expected %d, got %d", len(buf), n)
	}

	resp, err = http.Get(server.URL)
	if err != nil {
		t.Fatalf("Could not send request: %s", err)
	}
	buf = make([]byte, len(data))
	n, err = io.ReadFull(resp.Body, buf)
	if err != nil {
		t.Fatalf("Reading from body failed: %s", err)
	}
	if n != len(buf) {
		t.Fatalf("Unexpected length of body. Expected %d, got %d", len(buf), n)
	}
}

func PartialDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data[:len(data)/2])
}

func TestNoCacheAtPartialResponse(t *testing.T) {
	requestCount := 0
	incrementHandler := func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		PartialDataHandler(w, r)
	}
	handler := Cache(5*time.Hour, http.HandlerFunc(incrementHandler))
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Could not send request: %s", err)
	}
	buf := make([]byte, len(data))
	n, err := io.ReadFull(resp.Body, buf)
	if n != len(data)/2 || (err != nil && err != io.ErrUnexpectedEOF) {
		t.Fatalf("Reading from body failed: %s", err)
	}
	http.DefaultTransport.(*http.Transport).CloseIdleConnections()

	resp, err = http.Get(server.URL)
	if err != nil {
		t.Fatalf("Could not send request: %s", err)
	}
	n, err = io.ReadFull(resp.Body, buf)
	if n != len(data)/2 || (err != nil && err != io.ErrUnexpectedEOF) {
		t.Fatalf("Reading from body failed: %s", err)
	}

	_ = resp

	if requestCount != 2 {
		t.Fatalf("Handler was requested %d times, expected 2", requestCount)
	}
}

func Not200Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(201)
	w.Write(data)
}

func TestNoCacheAtNot200(t *testing.T) {
	requestCount := 0
	incrementHandler := func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		Not200Handler(w, r)
	}
	handler := Cache(5*time.Hour, http.HandlerFunc(incrementHandler))
	server := httptest.NewServer(handler)
	defer server.Close()
	buf := make([]byte, len(data))

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Could not send request: %s", err)
	}
	_, err = io.ReadFull(resp.Body, buf)
	if err != nil {
		t.Fatalf("Reading from body failed: %s", err)
	}

	resp, err = http.Get(server.URL)
	if err != nil {
		t.Fatalf("Could not send request: %s", err)
	}
	_, err = io.ReadFull(resp.Body, buf)
	if err != nil {
		t.Fatalf("Reading from body failed: %s", err)
	}

	if requestCount != 2 {
		t.Fatalf("Handler was requested %d times, expected 2", requestCount)
	}
}
