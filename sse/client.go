package sse

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Event represents a server-sent event.
type Event struct {
	Type string          // event type (e.g., "state", "patch", "action")
	Data json.RawMessage // raw JSON payload
}

// Client connects to an SSE endpoint.
type Client struct {
	URL        string
	HTTPClient *http.Client
	RetryDelay time.Duration // delay between reconnects (default 3s)
}

// Connect establishes an SSE connection and returns a channel of events.
// Automatically reconnects on disconnect.
func (c *Client) Connect(ctx context.Context) (<-chan Event, error) {
	ch := make(chan Event, 32)

	retryDelay := c.RetryDelay
	if retryDelay == 0 {
		retryDelay = 3 * time.Second
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	go func() {
		defer close(ch)
		for {
			err := c.stream(ctx, httpClient, ch)
			if err != nil && ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryDelay):
			}
		}
	}()

	return ch, nil
}

func (c *Client) stream(ctx context.Context, httpClient *http.Client, ch chan<- Event) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var eventType string
	var dataBuf bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event
			if dataBuf.Len() > 0 {
				data := make([]byte, dataBuf.Len())
				copy(data, dataBuf.Bytes())
				evt := Event{
					Type: eventType,
					Data: json.RawMessage(data),
				}
				select {
				case ch <- evt:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			eventType = ""
			dataBuf.Reset()
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			if dataBuf.Len() > 0 {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(strings.TrimSpace(data))
		}
		// Ignore "id:", "retry:", and comments (lines starting with ":")
	}

	return scanner.Err()
}

// PostAction sends an action to the server via HTTP POST.
func PostAction(url string, action string, payload interface{}) error {
	body := map[string]interface{}{
		"action":  action,
		"payload": payload,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}
