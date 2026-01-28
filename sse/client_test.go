package sse

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSSEParsing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		fmt.Fprint(w, "event:state\ndata:{\"count\":1}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event:patch\ndata:{\"op\":\"add\"}\n\n")
		flusher.Flush()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client := &Client{URL: srv.URL, RetryDelay: 100 * time.Millisecond}
	ch, err := client.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Read first event
	evt := <-ch
	if evt.Type != "state" {
		t.Fatalf("expected type 'state', got %q", evt.Type)
	}
	if string(evt.Data) != `{"count":1}` {
		t.Fatalf("unexpected data: %s", evt.Data)
	}

	// Read second event
	evt = <-ch
	if evt.Type != "patch" {
		t.Fatalf("expected type 'patch', got %q", evt.Type)
	}
}

func TestPostAction(t *testing.T) {
	var receivedBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(200)
	}))
	defer srv.Close()

	err := PostAction(srv.URL, "increment", map[string]int{"by": 1})
	if err != nil {
		t.Fatal(err)
	}
	if receivedBody == "" {
		t.Fatal("no body received")
	}
}
