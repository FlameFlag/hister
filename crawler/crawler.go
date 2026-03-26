package crawler

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Options configures the crawling behavior.
type Options struct {
	Headers      map[string]string
	Cookies      string
	UserAgent    string
	Timeout      time.Duration
	Depth        int
	Include      []string
	Exclude      []string
	SameDomain   bool
	MaxPages     int
	Delay        time.Duration
	Concurrency  int
	IgnoreRobots bool
	SkipExisting bool
	Headless     bool
	Browser      bool

	// Fetcher is used for content extraction. If nil, HTTPFetcher is used.
	Fetcher Fetcher

	// ExistsFunc checks whether a URL is already indexed.
	ExistsFunc func(url string) (bool, error)
}

// DefaultOptions returns sensible defaults matching the original behavior.
func DefaultOptions() Options {
	return Options{
		Timeout:      5 * time.Second,
		Depth:        1,
		SameDomain:   true,
		Delay:        1 * time.Second,
		Concurrency:  5,
		SkipExisting: true,
	}
}

// Crawler crawls URLs and calls OnPage for each page found.
type Crawler struct {
	Options Options
	OnPage  func(url, html string) error
}

// Run crawls the given seed URLs.
func (c *Crawler) Run(urls []string) error {
	if c.Options.Depth <= 1 {
		return c.runDirect(urls)
	}
	return c.runColly(urls)
}

// runDirect fetches each URL without link-following (depth==1 behavior).
func (c *Crawler) runDirect(urls []string) (err error) {
	fetcher := c.Options.Fetcher
	if fetcher == nil {
		fetcher = &HTTPFetcher{}
	}
	defer func() {
		err = errors.Join(err, fetcher.Close())
	}()

	for _, u := range urls {
		if u == "" {
			continue
		}
		if c.Options.SkipExisting && c.Options.ExistsFunc != nil {
			exists, err := c.Options.ExistsFunc(u)
			if err == nil && exists {
				continue
			}
		}
		html, err := fetcher.Fetch(context.Background(), u, c.buildHeaders())
		if err != nil {
			return fmt.Errorf("failed to fetch %s: %w", u, err)
		}
		if err := c.OnPage(u, html); err != nil {
			return err
		}
	}
	return nil
}

// runColly uses colly for recursive crawling.
func (c *Crawler) runColly(urls []string) (err error) {
	opts := []colly.CollectorOption{
		colly.MaxDepth(c.Options.Depth),
		colly.Async(true),
	}

	if c.Options.IgnoreRobots {
		opts = append(opts, colly.IgnoreRobotsTxt())
	}

	if c.Options.SameDomain {
		var domains []string
		for _, u := range urls {
			if d := extractDomain(u); d != "" {
				domains = append(domains, d)
			}
		}
		if len(domains) > 0 {
			opts = append(opts, colly.AllowedDomains(domains...))
		}
	}

	if len(c.Options.Include) > 0 {
		var filters []*regexp.Regexp
		for _, pattern := range c.Options.Include {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("invalid include pattern %q: %w", pattern, err)
			}
			filters = append(filters, re)
		}
		opts = append(opts, colly.URLFilters(filters...))
	}

	if len(c.Options.Exclude) > 0 {
		var filters []*regexp.Regexp
		for _, pattern := range c.Options.Exclude {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
			}
			filters = append(filters, re)
		}
		opts = append(opts, colly.DisallowedURLFilters(filters...))
	}

	if c.Options.UserAgent != "" {
		opts = append(opts, colly.UserAgent(c.Options.UserAgent))
	}

	col := colly.NewCollector(opts...)

	col.SetRequestTimeout(c.Options.Timeout)

	if err := col.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       c.Options.Delay,
		RandomDelay: c.Options.Delay / 2,
		Parallelism: c.Options.Concurrency,
	}); err != nil {
		return fmt.Errorf("failed to set limit rule: %w", err)
	}

	// Set up cookie jar
	if c.Options.Cookies != "" {
		jar, _ := cookiejar.New(nil)
		col.SetCookieJar(jar)
	}

	// Set custom headers
	headers := c.buildHeaders()
	if len(headers) > 0 {
		col.OnRequest(func(r *colly.Request) {
			for k, v := range headers {
				r.Headers.Set(k, v)
			}
			if c.Options.Cookies != "" {
				r.Headers.Set("Cookie", c.Options.Cookies)
			}
		})
	} else if c.Options.Cookies != "" {
		col.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Cookie", c.Options.Cookies)
		})
	}

	// Track page count
	pageCount := 0
	var crawlErr error

	// Use headless fetcher for content if configured
	var headlessFetcher Fetcher
	if c.Options.Headless || c.Options.Browser {
		headlessFetcher = &HeadlessFetcher{Visible: c.Options.Browser}
		defer func() {
			err = errors.Join(err, headlessFetcher.Close())
		}()
	}

	col.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if c.Options.MaxPages > 0 && pageCount >= c.Options.MaxPages {
			return
		}
		link := e.Attr("href")
		_ = e.Request.Visit(e.Request.AbsoluteURL(link))
	})

	col.OnResponse(func(r *colly.Response) {
		if crawlErr != nil {
			return
		}
		if c.Options.MaxPages > 0 && pageCount >= c.Options.MaxPages {
			return
		}

		contentType := r.Headers.Get("Content-Type")
		if !strings.Contains(contentType, "html") {
			return
		}

		u := r.Request.URL.String()

		if c.Options.SkipExisting && c.Options.ExistsFunc != nil {
			exists, err := c.Options.ExistsFunc(u)
			if err == nil && exists {
				return
			}
		}

		var html string
		if headlessFetcher != nil {
			var err error
			html, err = headlessFetcher.Fetch(context.Background(), u, c.buildHeaders())
			if err != nil {
				// Fall back to the response body
				html = string(r.Body)
			}
		} else {
			html = string(r.Body)
		}

		if err := c.OnPage(u, html); err != nil {
			crawlErr = err
			return
		}
		pageCount++
	})

	for _, u := range urls {
		if err := col.Visit(u); err != nil {
			return fmt.Errorf("failed to visit %s: %w", u, err)
		}
	}

	col.Wait()

	return crawlErr
}

func (c *Crawler) buildHeaders() map[string]string {
	headers := make(map[string]string)
	maps.Copy(headers, c.Options.Headers)
	return headers
}

func extractDomain(rawURL string) string {
	// Simple domain extraction: strip scheme, take host part
	u := rawURL
	if idx := strings.Index(u, "://"); idx >= 0 {
		u = u[idx+3:]
	}
	if idx := strings.Index(u, "/"); idx >= 0 {
		u = u[:idx]
	}
	if idx := strings.Index(u, ":"); idx >= 0 {
		u = u[:idx]
	}
	return u
}

// ParseHeaders parses curl-style header strings ("Key: Value") into a map.
func ParseHeaders(rawHeaders []string) map[string]string {
	headers := make(map[string]string)
	for _, h := range rawHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

// BuildHTTPClient creates an http.Client configured with the given options.
func BuildHTTPClient(opts Options) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Timeout: opts.Timeout,
		Jar:     jar,
	}
}
