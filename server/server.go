package server

//go:generate sh -c "cd web && npm install && npm run build"

import (
	"bufio"
	"crypto/rand"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/indexer"
	"github.com/asciimoo/hister/server/model"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var (
	sessionStore    *sessions.CookieStore
	errCSRFMismatch = errors.New("CSRF token mismatch")
	storeName       = "hister"
	tokName         = "csrf_token"
)

type historyItem struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Query  string `json:"query"`
	Delete bool   `json:"delete"`
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

type spaHandler struct {
	root http.FileSystem
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cleanPath := path.Clean(r.URL.Path)

	if cleanPath == "/" || cleanPath == "" {
		h.serveIndex(w, r)
		return
	}

	lookupPath := strings.TrimPrefix(cleanPath, "/")

	if served := h.serveCompressed(w, r, lookupPath); served {
		return
	}

	file, err := h.root.Open(lookupPath)
	if err == nil {
		defer file.Close()
		stat, err := file.Stat()
		if err == nil && !stat.IsDir() {
			if ct := mime.TypeByExtension(path.Ext(lookupPath)); ct != "" {
				w.Header().Set("Content-Type", ct)
			}
			http.ServeContent(w, r, lookupPath, stat.ModTime(), file)
			return
		}
	}

	if path.Ext(lookupPath) != "" {
		http.NotFound(w, r)
		return
	}

	file, err = h.root.Open(lookupPath + ".html")
	if err == nil {
		defer file.Close()
		stat, err := file.Stat()
		if err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(w, r, lookupPath+".html", stat.ModTime(), file)
			return
		}
	}

	h.serveIndex(w, r)
}

func (h spaHandler) serveCompressed(w http.ResponseWriter, r *http.Request, lookupPath string) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	supportsBr := strings.Contains(acceptEncoding, "br")
	supportsGzip := strings.Contains(acceptEncoding, "gzip")

	contentType := mime.TypeByExtension(path.Ext(lookupPath))
	if contentType == "" {
		return false
	}

	// try brotli first
	if supportsBr {
		brPath := lookupPath + ".br"
		file, err := h.root.Open(brPath)
		if err == nil {
			defer file.Close()
			stat, err := file.Stat()
			if err == nil && !stat.IsDir() {
				w.Header().Set("Content-Encoding", "br")
				w.Header().Set("Content-Type", contentType)
				w.Header().Set("Vary", "Accept-Encoding")
				http.ServeContent(w, r, lookupPath, stat.ModTime(), file)
				return true
			}
		}
	}

	// try gzip as fallback
	if supportsGzip {
		gzPath := lookupPath + ".gz"
		file, err := h.root.Open(gzPath)
		if err == nil {
			defer file.Close()
			stat, err := file.Stat()
			if err == nil && !stat.IsDir() {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Content-Type", contentType)
				w.Header().Set("Vary", "Accept-Encoding")
				http.ServeContent(w, r, lookupPath, stat.ModTime(), file)
				return true
			}
		}
	}

	return false
}

func (h spaHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	if served := h.serveCompressed(w, r, "index.html"); served {
		return
	}

	f, err := h.root.Open("index.html")
	if err != nil {
		http.NotFound(w, &http.Request{})
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.NotFound(w, &http.Request{})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", stat.ModTime(), f)
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Header() http.Header {
	return lrw.ResponseWriter.Header()
}

func (lrw *loggingResponseWriter) Write(d []byte) (int, error) {
	return lrw.ResponseWriter.Write(d)
}

func (lrw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := lrw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijacking not supported")
	}
	return hj.Hijack()
}

var ws = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type webContext struct {
	Request  *http.Request
	Response http.ResponseWriter
	Config   *config.Config
}

//go:embed web/dist
//go:embed web/dist/*
//go:embed web/dist/**/*
var svelteDistFS embed.FS
var svelteAppFS http.Handler

func respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func parseForm(r *http.Request) error {
	r.Form = make(url.Values)
	r.PostForm = make(url.Values)
	return r.ParseMultipartForm(10 << 20)
}

func checkSafeRequest(r *http.Request, baseURL string) bool {
	origin := r.Header.Get("Origin")
	if origin == "same-origin" || origin == "" {
		return true
	}
	if strings.HasPrefix(origin, baseURL) {
		return true
	}
	return strings.HasPrefix(r.Header.Get("Referer"), baseURL)
}

func isAllowedOrigin(r *http.Request, path string) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "hister://":
		return true
	case "chrome-extension://cciilamhchpmbdnniabclekddabkifhb", "":
		return path == "/add"
	}

	return path == "/add" && strings.HasPrefix(origin, "moz-extension://")
}

func generateCSRFToken(session *sessions.Session) (string, error) {
	tok := rand.Text()
	session.Values[tokName] = tok
	return tok, nil
}

func init() {
	svelteDistFS2, err := fs.Sub(svelteDistFS, "web/dist")
	if err != nil {
		panic(err)
	}
	svelteAppFS = spaHandler{root: http.FS(svelteDistFS2)}
	mime.AddExtensionType(".mjs", "application/javascript")
	mime.AddExtensionType(".js", "application/javascript")
}

func Listen(cfg *config.Config) {
	sessionStore = sessions.NewCookieStore(cfg.SecretKey()[:32])
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		HttpOnly: true,
	}
	handler := createRouter(cfg)
	handler = withLogging(handler)

	log.Info().Str("Address", cfg.Server.Address).Str("URL", cfg.BaseURL("/")).Msg("Starting webserver")
	http.ListenAndServe(cfg.Server.Address, handler)
}

type contextHandler struct {
	cfg *config.Config
}

func (ch *contextHandler) wrap(f func(*webContext)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(&webContext{Request: r, Response: w, Config: ch.cfg})
	}
}

func (ch *contextHandler) wrapCSRF(f func(*webContext)) http.HandlerFunc {
	return withCSRF(ch.cfg, ch.wrap(f))
}

func withCSRF(cfg *config.Config, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isAllowedOrigin(r, r.URL.Path) {
			h(w, r)
			return
		}

		session, err := sessionStore.Get(r, storeName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead && !checkSafeRequest(r, cfg.BaseURL("/")) {
			sToken, ok := session.Values[tokName].(string)
			if !ok || (r.PostFormValue(tokName) != sToken && r.Header.Get("X-CSRF-Token") != sToken) {
				http.Error(w, errCSRFMismatch.Error(), http.StatusInternalServerError)
				return
			}
		}

		tok, err := generateCSRFToken(session)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-CSRF-Token", tok)
		h(w, r)
	}
}

func createRouter(cfg *config.Config) http.Handler {
	ch := &contextHandler{cfg: cfg}
	mux := http.NewServeMux()

	searchHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrade := r.Header.Get("Upgrade")
		if strings.EqualFold(upgrade, "websocket") {
			ch.wrap(serveSearch)(w, r)
			return
		}

		accept := r.Header.Get("Accept")
		hasJSON := strings.Contains(accept, "application/json")
		hasHTML := strings.Contains(accept, "text/html")
		isGenericAccept := accept == "*/*" || accept == "" || strings.HasPrefix(accept, "*/*")

		if hasJSON && !hasHTML && !isGenericAccept {
			ch.wrap(serveSearch)(w, r)
			return
		}
		svelteAppFS.ServeHTTP(w, r)
	})

	mux.Handle("GET /search", searchHandler)
	mux.HandleFunc("GET /api/rules", ch.wrap(serveAPIRules))
	mux.HandleFunc("POST /api/rules", ch.wrapCSRF(serveAPIRules))
	mux.HandleFunc("GET /api/history", ch.wrap(serveAPIHistory))
	mux.HandleFunc("POST /history", ch.wrapCSRF(serveHistory))
	mux.HandleFunc("GET /api/stats", ch.wrap(serveAPIStats))
	mux.HandleFunc("GET /api/csrf", ch.wrap(serveCSRFToken))
	mux.HandleFunc("POST /delete", ch.wrapCSRF(serveDeleteDocument))
	mux.HandleFunc("POST /delete_alias", ch.wrapCSRF(serveDeleteAlias))
	mux.HandleFunc("POST /add_alias", ch.wrapCSRF(serveAddAlias))
	mux.HandleFunc("POST /add", ch.wrapCSRF(serveAdd))
	mux.HandleFunc("GET /readable", ch.wrap(serveReadable))
	mux.Handle("/", svelteAppFS)

	return mux
}

func withLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("path", r.URL.Path).Str("rawquery", r.URL.RawQuery).Msg("Router request")
		start := time.Now()
		lrw := &loggingResponseWriter{w, http.StatusOK}
		h.ServeHTTP(lrw, r)
		log.Info().Str("Method", r.Method).Int("Status", lrw.statusCode).Dur("LoadTimeMS", time.Since(start)).Str("URL", r.RequestURI).Msg("WEB")
	})
}

func serveSearch(c *webContext) {
	q := c.Request.URL.Query().Get("q")
	if q != "" {
		query := &indexer.Query{Text: q}
		for param, field := range map[string]*int64{"date_from": &query.DateFrom, "date_to": &query.DateTo} {
			if v := c.Request.URL.Query().Get(param); v != "" {
				if t, err := time.Parse("2006-01-02", v); err == nil {
					*field = t.Unix()
				}
			}
		}
		r, err := doSearch(query, c.Config)
		if err != nil {
			fmt.Println(err)
			respondError(c.Response, "Internal server error", http.StatusInternalServerError)
			return
		}
		respondJSON(c.Response, r)
		return
	}
	conn, err := ws.Upgrade(c.Response, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade websocket request")
		return
	}
	defer conn.Close()
	for {
		_, q, err := conn.ReadMessage()
		if err != nil {
			log.Error().Err(err).Msg("failed to read websocket message")
			break
		}
		var query *indexer.Query
		err = json.Unmarshal(q, &query)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse query")
			continue
		}
		res, err := doSearch(query, c.Config)
		if err != nil {
			log.Error().Err(err).Msg("search error")
			continue
		}
		jr, err := json.Marshal(res)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal indexer results")
		}
		if err := conn.WriteMessage(websocket.TextMessage, jr); err != nil {
			log.Error().Err(err).Msg("failed to write websocket message")
			break
		}
	}
}

func doSearch(query *indexer.Query, cfg *config.Config) (*indexer.Results, error) {
	start := time.Now()
	oq := query.Text
	query.Text = cfg.Rules.ResolveAliases(query.Text)
	res, err := indexer.Search(cfg, query)
	if err != nil {
		log.Error().Err(err).Msg("failed to get indexer results")
	}
	if res == nil {
		res = &indexer.Results{}
	}
	hr, err := model.GetURLsByQuery(oq)
	if err == nil && len(hr) > 0 {
		res.History = hr
	}
	if oq != "" {
		res.QuerySuggestion = model.GetQuerySuggestion(oq)
	}
	res.SearchDuration = fmt.Sprintf("%v", time.Since(start))
	return res, nil
}

func serveAdd(c *webContext) {
	d := &indexer.Document{}
	if strings.Contains(c.Request.Header.Get("Content-Type"), "json") {
		if err := json.NewDecoder(c.Request.Body).Decode(d); err != nil {
			log.Error().Err(err).Msg("failed to decode JSON")
			respondError(c.Response, "Invalid JSON", http.StatusBadRequest)
			return
		}
	} else {
		if err := parseForm(c.Request); err != nil {
			log.Error().Err(err).Msg("failed to parse form")
			respondError(c.Response, "Invalid form data", http.StatusBadRequest)
			return
		}
		f := c.Request.PostForm
		d.URL, d.Title, d.Text = f.Get("url"), f.Get("title"), f.Get("text")
	}

	if d.URL == "" {
		log.Error().Msg("empty URL provided")
		respondError(c.Response, "URL is required", http.StatusBadRequest)
		return
	}

	if err := d.NormalizeURL(); err != nil {
		log.Error().Err(err).Str("URL", d.URL).Msg("failed to normalize URL")
		respondError(c.Response, fmt.Sprintf("Invalid URL: %v", err), http.StatusBadRequest)
		return
	}

	if d.HTML == "" && (d.Title == "" || d.Text == "") {
		if err := d.Process(); err != nil {
			log.Error().Err(err).Str("URL", d.URL).Msg("failed to process document")
			respondError(c.Response, fmt.Sprintf("Failed to process document: %v", err), http.StatusBadRequest)
			return
		}
	}

	if d.Title == "" {
		d.Title = d.URL
	}
	if d.Text == "" {
		d.Text = d.Title
	}

	if !c.Config.Rules.IsSkip(d.URL) && !strings.HasPrefix(d.URL, c.Config.BaseURL("/")) {
		if err := indexer.Add(d); err != nil {
			log.Error().Err(err).Str("URL", d.URL).Msg("failed to create index")
			respondError(c.Response, fmt.Sprintf("Failed to create index: %v", err), http.StatusInternalServerError)
			return
		}
		c.Response.WriteHeader(http.StatusCreated)
	} else {
		log.Debug().Str("url", d.URL).Msg("skip indexing")
		c.Response.WriteHeader(http.StatusNotAcceptable)
	}
}

func serveHistory(c *webContext) {
	h := &historyItem{}
	if err := json.NewDecoder(c.Request.Body).Decode(h); err != nil {
		respondError(c.Response, "Invalid request", http.StatusBadRequest)
		return
	}
	if h.Delete {
		if err := model.DeleteHistoryItem(h.Query, h.URL); err != nil {
			respondError(c.Response, "Failed to delete history item", http.StatusInternalServerError)
		}
		return
	}

	// Skip internal hister URLs (search pages, etc.)
	if strings.HasPrefix(h.URL, "/") {
		return
	}

	if err := model.UpdateHistory(strings.TrimSpace(h.Query), strings.TrimSpace(h.URL), strings.TrimSpace(h.Title)); err != nil {
		log.Error().Err(err).Msg("failed to update history")
		respondError(c.Response, "Failed to update history", http.StatusInternalServerError)
		return
	}
}

func serveReadable(c *webContext) {
	u := c.Request.URL.Query().Get("url")
	doc := indexer.GetByURL(u)
	if doc == nil {
		respondError(c.Response, "Document not found", http.StatusNotFound)
		return
	}
	pu, err := url.Parse(u)
	if err != nil {
		respondError(c.Response, "Invalid URL", http.StatusBadRequest)
		return
	}
	r, err := readability.FromReader(strings.NewReader(doc.HTML), pu)
	if err != nil {
		respondError(c.Response, "Failed to parse readability", http.StatusInternalServerError)
		return
	}
	r.RenderHTML(c.Response)
}

func serveAddAlias(c *webContext) {
	if err := parseForm(c.Request); err != nil {
		respondError(c.Response, "Invalid form data", http.StatusBadRequest)
		return
	}
	f := c.Request.PostForm
	if kw, val := f.Get("alias-keyword"), f.Get("alias-value"); kw != "" && val != "" {
		c.Config.Rules.Aliases[kw] = val
	}
	if err := c.Config.SaveRules(); err != nil {
		log.Error().Err(err).Msg("failed to save rules")
		respondError(c.Response, "Failed to save rules", http.StatusInternalServerError)
		return
	}
	c.Response.WriteHeader(http.StatusOK)
}

func serveAPIRules(c *webContext) {
	if c.Request.Method == http.MethodPost {
		var rules struct {
			Skip     []string          `json:"skip"`
			Priority []string          `json:"priority"`
			Aliases  map[string]string `json:"aliases"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&rules); err != nil {
			respondError(c.Response, "Invalid JSON", http.StatusBadRequest)
			return
		}
		c.Config.Rules.Skip.ReStrs, c.Config.Rules.Priority.ReStrs, c.Config.Rules.Aliases = rules.Skip, rules.Priority, rules.Aliases
		if err := c.Config.SaveRules(); err != nil {
			log.Error().Err(err).Msg("failed to save rules")
			respondError(c.Response, "Failed to save rules", http.StatusInternalServerError)
			return
		}
		c.Response.WriteHeader(http.StatusOK)
		return
	}
	c.Response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Response).Encode(map[string]any{
		"skip":     c.Config.Rules.Skip.ReStrs,
		"priority": c.Config.Rules.Priority.ReStrs,
		"aliases":  c.Config.Rules.Aliases,
	})
}

func serveAPIHistory(c *webContext) {
	limit := 40
	if l, err := strconv.Atoi(c.Request.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	hs, err := model.GetLatestHistoryItems(limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch history")
		respondError(c.Response, "Failed to fetch history", http.StatusInternalServerError)
		return
	}

	for _, h := range hs {
		if doc := indexer.GetByURL(h.URL); doc != nil && doc.Favicon != "" {
			h.Favicon = doc.Favicon
		}
	}

	c.Response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Response).Encode(hs)
}

func serveAPIStats(c *webContext) {
	pages, domains := indexer.Stats()
	minDate, maxDate := indexer.DateRange()
	respondJSON(c.Response, map[string]any{
		"pagesIndexed": pages,
		"domains":      domains,
		"dateRange":    "Last 30 days",
		"minDate":      minDate,
		"maxDate":      maxDate,
	})
}

func serveCSRFToken(c *webContext) {
	session, err := sessionStore.Get(c.Request, storeName)
	if err != nil {
		respondError(c.Response, err.Error(), http.StatusInternalServerError)
		return
	}
	tok, err := generateCSRFToken(session)
	if err != nil {
		respondError(c.Response, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := session.Save(c.Request, c.Response); err != nil {
		respondError(c.Response, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(c.Response, map[string]string{"token": tok})
}

func serveDeleteDocument(c *webContext) {
	if err := parseForm(c.Request); err != nil {
		respondError(c.Response, "Invalid form data", http.StatusBadRequest)
		return
	}
	url := c.Request.PostForm.Get("url")
	if url == "" {
		respondError(c.Response, "URL is required", http.StatusBadRequest)
		return
	}

	d := &indexer.Document{URL: url}
	if err := d.NormalizeURL(); err != nil {
		log.Error().Err(err).Str("url", url).Msg("failed to normalize URL")
		respondError(c.Response, fmt.Sprintf("Invalid URL: %v", err), http.StatusBadRequest)
		return
	}

	doc := indexer.GetByURL(d.URL)
	if doc == nil {
		log.Error().Str("url", d.URL).Msg("document not found")
		respondError(c.Response, "Document not found", http.StatusNotFound)
		return
	}

	if err := indexer.Delete(doc.URL); err != nil {
		log.Error().Err(err).Str("url", doc.URL).Msg("failed to delete from index")
		respondError(c.Response, fmt.Sprintf("Failed to delete: %v", err), http.StatusInternalServerError)
		return
	}

	log.Debug().Str("url", doc.URL).Msg("document deleted successfully")
	c.Response.WriteHeader(http.StatusOK)
}

func serveDeleteAlias(c *webContext) {
	if err := parseForm(c.Request); err != nil {
		respondError(c.Response, "Invalid form data", http.StatusBadRequest)
		return
	}
	alias := c.Request.PostForm.Get("alias")
	if alias == "" {
		respondError(c.Response, "Alias is required", http.StatusBadRequest)
		return
	}

	if _, ok := c.Config.Rules.Aliases[alias]; !ok {
		c.Response.WriteHeader(http.StatusNotFound)
		return
	}
	delete(c.Config.Rules.Aliases, alias)
	if err := c.Config.SaveRules(); err != nil {
		log.Error().Err(err).Msg("failed to save rules")
		respondError(c.Response, "Failed to save rules", http.StatusInternalServerError)
		return
	}

	c.Response.WriteHeader(http.StatusOK)
}
