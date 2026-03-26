package crawler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// HeadlessFetcher uses a headless browser (chromedp) to render pages with JavaScript.
type HeadlessFetcher struct {
	mu          sync.Mutex
	allocCtx    context.Context
	allocCancel context.CancelFunc
	Visible     bool // launch visible browser with user's real profile
}

// browserInfo describes a browser installation found on the system.
type browserInfo struct {
	name    string
	dataDir string   // user profile / data directory
	procs   []string // process name patterns for detecting running instances
}

func (f *HeadlessFetcher) ensureBrowser() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.allocCtx != nil {
		return nil
	}

	if f.Visible {
		return f.launchVisible()
	}

	// Headless mode: find a Chromium binary and use default options
	binPath := findChromiumBinary("Chrome")
	if binPath == "" {
		// Try common binary names in PATH
		for _, name := range []string{"google-chrome", "chromium", "chromium-browser"} {
			if p, err := exec.LookPath(name); err == nil {
				binPath = p
				break
			}
		}
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(binPath),
	)

	f.allocCtx, f.allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	return nil
}

// launchVisible launches a browser in visible mode using the user's real profile
// (cookies, sessions, extensions). The browser must be fully closed first since
// it locks its profile directory.
func (f *HeadlessFetcher) launchVisible() error {
	bi, err := detectBrowser()
	if err != nil {
		return err
	}

	if isBrowserRunning(bi.procs) {
		return fmt.Errorf("%s is currently running — please close all %s windows and try again", bi.name, bi.name)
	}

	binPath := findChromiumBinary(bi.name)
	if binPath == "" {
		return fmt.Errorf("%s binary not found", bi.name)
	}

	// Start from defaults, then override for visible mode
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(binPath),
		chromedp.UserDataDir(bi.dataDir),
		chromedp.Flag("headless", false),
		chromedp.Flag("no-startup-window", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("use-mock-keychain", false),
	)

	f.allocCtx, f.allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	return nil
}

func (f *HeadlessFetcher) Fetch(ctx context.Context, url string, headers map[string]string) (string, error) {
	if err := f.ensureBrowser(); err != nil {
		return "", err
	}

	tabCtx, tabCancel := chromedp.NewContext(f.allocCtx)
	defer tabCancel()

	// Visible mode gets a longer timeout so the user can interact (e.g. solve CAPTCHAs)
	if f.Visible {
		var cancel context.CancelFunc
		tabCtx, cancel = context.WithTimeout(tabCtx, 5*time.Minute)
		defer cancel()
	}

	actions := make([]chromedp.Action, 0, 4)

	// Set extra headers if provided
	if len(headers) > 0 {
		h := make(network.Headers, len(headers))
		for k, v := range headers {
			h[k] = v
		}
		actions = append(actions, network.SetExtraHTTPHeaders(h))
	}

	actions = append(actions,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		waitNetworkIdle(2*time.Second),
	)

	if err := chromedp.Run(tabCtx, actions...); err != nil {
		return "", fmt.Errorf("failed to load page: %w", err)
	}

	var html string
	if err := chromedp.Run(tabCtx, chromedp.OuterHTML("html", &html, chromedp.ByQuery)); err != nil {
		return "", fmt.Errorf("failed to get HTML: %w", err)
	}

	return html, nil
}

func (f *HeadlessFetcher) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.allocCancel != nil {
		f.allocCancel()
	}
	return nil
}

// waitNetworkIdle waits until there are no in-flight network requests for the given duration.
func waitNetworkIdle(quiesce time.Duration) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		done := make(chan struct{}, 1)
		inflight := 0
		var mu sync.Mutex
		var timer *time.Timer

		resetTimer := func() {
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(quiesce, func() {
				mu.Lock()
				defer mu.Unlock()
				if inflight == 0 {
					select {
					case done <- struct{}{}:
					default:
					}
				}
			})
		}

		chromedp.ListenTarget(cctx, func(ev interface{}) {
			mu.Lock()
			defer mu.Unlock()
			switch ev.(type) {
			case *network.EventRequestWillBeSent:
				inflight++
			case *network.EventLoadingFinished, *network.EventLoadingFailed:
				inflight--
				if inflight <= 0 {
					inflight = 0
					resetTimer()
				}
			}
		})

		resetTimer()
		select {
		case <-done:
			return nil
		case <-cctx.Done():
			return nil // timeout is acceptable — page likely loaded enough
		}
	}
}

// browserCandidates returns the supported browsers in preference order for the current OS.
// Each candidate includes the expected user-data directory path and process name patterns.
func browserCandidates() []browserInfo {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		appSupport := filepath.Join(home, "Library", "Application Support")
		return []browserInfo{
			{name: "Chrome", dataDir: filepath.Join(appSupport, "Google", "Chrome"), procs: []string{"Google Chrome"}},
			{name: "Chromium", dataDir: filepath.Join(appSupport, "Chromium"), procs: []string{"Chromium"}},
			{name: "Edge", dataDir: filepath.Join(appSupport, "Microsoft Edge"), procs: []string{"Microsoft Edge"}},
			{name: "Brave", dataDir: filepath.Join(appSupport, "BraveSoftware", "Brave-Browser"), procs: []string{"Brave Browser"}},
		}
	case "linux":
		return []browserInfo{
			{name: "Chrome", dataDir: filepath.Join(home, ".config", "google-chrome"), procs: []string{"chrome", "google-chrome"}},
			{name: "Chromium", dataDir: filepath.Join(home, ".config", "chromium"), procs: []string{"chromium"}},
			{name: "Edge", dataDir: filepath.Join(home, ".config", "microsoft-edge"), procs: []string{"msedge", "microsoft-edge"}},
			{name: "Brave", dataDir: filepath.Join(home, ".config", "BraveSoftware", "Brave-Browser"), procs: []string{"brave"}},
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		return []browserInfo{
			{name: "Chrome", dataDir: filepath.Join(localAppData, "Google", "Chrome", "User Data"), procs: []string{"chrome.exe"}},
			{name: "Chromium", dataDir: filepath.Join(localAppData, "Chromium", "User Data"), procs: []string{"chromium.exe"}},
			{name: "Edge", dataDir: filepath.Join(localAppData, "Microsoft", "Edge", "User Data"), procs: []string{"msedge.exe"}},
			{name: "Brave", dataDir: filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data"), procs: []string{"brave.exe"}},
		}
	}
	return nil
}

// detectBrowser returns the first browser whose data directory exists on disk.
func detectBrowser() (browserInfo, error) {
	for _, bi := range browserCandidates() {
		if info, err := os.Stat(bi.dataDir); err == nil && info.IsDir() {
			return bi, nil
		}
	}
	return browserInfo{}, fmt.Errorf("no supported browser profile found (looked for Chrome, Chromium, Edge, Brave)")
}

// isBrowserRunning checks whether any of the given process name patterns are running.
func isBrowserRunning(procs []string) bool {
	return slices.ContainsFunc(procs, processExists)
}

func processExists(name string) bool {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("pgrep", "-f", name).Output()
		return err == nil && len(strings.TrimSpace(string(out))) > 0
	case "linux":
		out, err := exec.Command("pgrep", "-f", name).Output()
		return err == nil && len(strings.TrimSpace(string(out))) > 0
	case "windows":
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", name), "/NH").Output()
		return err == nil && strings.Contains(string(out), name)
	}
	return false
}

// findChromiumBinary returns the path to the given Chromium-variant browser binary.
// Falls back to common binary names in PATH if the specific browser isn't found.
func findChromiumBinary(name string) string {
	var candidates []string

	switch runtime.GOOS {
	case "darwin":
		switch name {
		case "Chrome":
			candidates = []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"}
		case "Chromium":
			candidates = []string{"/Applications/Chromium.app/Contents/MacOS/Chromium"}
		case "Edge":
			candidates = []string{"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge"}
		case "Brave":
			candidates = []string{"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser"}
		}
	case "linux":
		switch name {
		case "Chrome":
			candidates = []string{"google-chrome-stable", "google-chrome"}
		case "Chromium":
			candidates = []string{"chromium-browser", "chromium"}
		case "Edge":
			candidates = []string{"microsoft-edge-stable", "microsoft-edge"}
		case "Brave":
			candidates = []string{"brave-browser-stable", "brave-browser"}
		}
	case "windows":
		pf := os.Getenv("ProgramFiles")
		pf86 := os.Getenv("ProgramFiles(x86)")
		switch name {
		case "Chrome":
			candidates = []string{
				filepath.Join(pf, "Google", "Chrome", "Application", "chrome.exe"),
				filepath.Join(pf86, "Google", "Chrome", "Application", "chrome.exe"),
			}
		case "Chromium":
			candidates = []string{
				filepath.Join(pf, "Chromium", "Application", "chrome.exe"),
			}
		case "Edge":
			candidates = []string{
				filepath.Join(pf, "Microsoft", "Edge", "Application", "msedge.exe"),
				filepath.Join(pf86, "Microsoft", "Edge", "Application", "msedge.exe"),
			}
		case "Brave":
			candidates = []string{
				filepath.Join(pf, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
				filepath.Join(pf86, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			}
		}
	}

	for _, c := range candidates {
		if runtime.GOOS == "linux" {
			// On Linux, candidates are command names — look up in PATH
			if p, err := exec.LookPath(c); err == nil {
				return p
			}
		} else {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	}

	// Fall back to common binary names in PATH
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}
