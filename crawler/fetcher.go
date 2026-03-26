package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Fetcher defines an interface for fetching page content.
type Fetcher interface {
	Fetch(ctx context.Context, url string, headers map[string]string) (html string, err error)
	Close() error
}

// HTTPFetcher fetches pages using net/http.
type HTTPFetcher struct {
	Client *http.Client
}

func (f *HTTPFetcher) Fetch(ctx context.Context, url string, headers map[string]string) (_ string, err error) {
	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid response code: %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "html") {
		return "", errors.New("invalid content type: " + contentType)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}
	return string(body), nil
}

func (f *HTTPFetcher) Close() error {
	return nil
}
