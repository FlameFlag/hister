// Package extractor provides HTML content extraction for documents.
package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/extractor/extractors/godoc"
	"github.com/asciimoo/hister/server/extractor/extractors/stackoverflow"
	"github.com/asciimoo/hister/server/types"
)

// Extractor extracts content from a Document.
type Extractor interface {
	// Name returns a human-readable identifier for the extractor.
	Name() string

	// Match reports whether this extractor is applicable to the given document.
	// Extract and Preview will only be called when Match returns true.
	Match(*document.Document) bool

	// Extract rewrites documents before the documents are added to the index.
	// The returned bool signals whether the caller should continue trying
	// subsequent extractors: true means this attempt was inconclusive and the
	// next matching extractor should be tried
	Extract(*document.Document) (bool, error)

	// Preview returns a rendered representation of the document suitable for
	// display (e.g. readable HTML or plain text).
	// The returned bool signals whether the caller should continue trying
	// subsequent extractors: true means this attempt was inconclusive and the
	// next matching extractor should be tried
	Preview(*document.Document) (types.PreviewResponse, bool, error)

	// GetConfig returns the extractor's current configuration. Before
	// SetConfig is called, implementations must return their default config.
	GetConfig() *config.Extractor

	// SetConfig applies cfg to the extractor, overwriting defaults.
	// Implementations should return an error for unrecognised option keys.
	SetConfig(*config.Extractor) error
}

// ErrNoExtractor is returned when no extractor can handle the document.
var ErrNoExtractor = errors.New("no extractor found")

var extractors = []Extractor{
	&stackoverflow.StackoverflowExtractor{},
	&godoc.GoDocExtractor{},
	&readabilityExtractor{},
	&defaultExtractor{},
}

// Init applies user-supplied extractor configurations on top of each
// extractor's defaults. It must be called before Extract or Preview.
// cfgs is keyed by lowercased extractor name (as Viper lowercases YAML keys).
func Init(cfgs map[string]*config.Extractor) error {
	for _, e := range extractors {
		def := e.GetConfig()
		merged := &config.Extractor{
			Enable:  def.Enable,
			Options: make(map[string]any, len(def.Options)),
		}
		maps.Copy(merged.Options, def.Options)
		if user, ok := cfgs[strings.ToLower(e.Name())]; ok && user != nil {
			merged.Enable = user.Enable
			maps.Copy(merged.Options, user.Options)
		}
		if err := e.SetConfig(merged); err != nil {
			return fmt.Errorf("extractor %s: %w", e.Name(), err)
		}
	}
	return nil
}

// Extract tries each registered extractor in order and returns the first
// successful result. Returns ErrNoExtractor if none succeed.
func Extract(d *document.Document) error {
	for _, e := range extractors {
		if !e.GetConfig().Enable {
			continue
		}
		if e.Match(d) {
			cont, err := e.Extract(d)
			log.Debug().Str("URL", d.URL).Str("Extractor", e.Name()).Msg("Extracting data")
			if err != nil {
				log.Warn().Err(err).Str("URL", d.URL).Str("Extractor", e.Name()).Msg("Failed to extract content")
			} else if !cont {
				return nil
			}
		}
	}
	return ErrNoExtractor
}

// Preview returns a rendered preview of the document using the first matching
// extractor. Returns ErrNoExtractor if none match.
func Preview(d *document.Document) (types.PreviewResponse, error) {
	for _, e := range extractors {
		if !e.GetConfig().Enable {
			continue
		}
		if e.Match(d) {
			log.Debug().Str("URL", d.URL).Str("Extractor", e.Name()).Msg("Creating preview")
			resp, cont, err := e.Preview(d)
			if err != nil {
				log.Warn().Err(err).Str("URL", d.URL).Str("Extractor", e.Name()).Msg("Failed to preview content")
			} else {
				return resp, nil
			}
			if !cont {
				break
			}
		}
	}
	return types.PreviewResponse{}, ErrNoExtractor
}

type defaultExtractor struct {
	cfg *config.Extractor
}

type readabilityExtractor struct {
	cfg *config.Extractor
}

func (e *defaultExtractor) GetConfig() *config.Extractor {
	if e.cfg == nil {
		return &config.Extractor{Enable: true, Options: map[string]any{}}
	}
	return e.cfg
}

func (e *defaultExtractor) SetConfig(c *config.Extractor) error {
	for k := range c.Options {
		return fmt.Errorf("unknown option %q", k)
	}
	e.cfg = c
	return nil
}

func (e *readabilityExtractor) GetConfig() *config.Extractor {
	if e.cfg == nil {
		return &config.Extractor{Enable: true, Options: map[string]any{}}
	}
	return e.cfg
}

func (e *readabilityExtractor) SetConfig(c *config.Extractor) error {
	for k := range c.Options {
		return fmt.Errorf("unknown option %q", k)
	}
	e.cfg = c
	return nil
}

func (e *defaultExtractor) Name() string {
	return "Default"
}

func (e *defaultExtractor) Match(_ *document.Document) bool {
	return true
}

func (e *defaultExtractor) Extract(d *document.Document) (bool, error) {
	d.Title = ""
	r := bytes.NewReader([]byte(d.HTML))
	doc := html.NewTokenizer(r)
	inBody := false
	skip := false
	var text strings.Builder
	var currentTag string
out:
	for {
		tt := doc.Next()
		switch tt {
		case html.ErrorToken:
			err := doc.Err()
			if errors.Is(err, io.EOF) {
				break out
			}
			return false, errors.New("failed to parse html: " + err.Error())
		case html.SelfClosingTagToken, html.StartTagToken:
			tn, _ := doc.TagName()
			currentTag = string(tn)
			switch currentTag {
			case "body":
				inBody = true
			case "script", "style", "noscript":
				skip = true
			}
		case html.TextToken:
			if currentTag == "title" {
				d.Title += strings.TrimSpace(string(doc.Text()))
			}
			if inBody && !skip {
				text.Write(doc.Text())
			}
		case html.EndTagToken:
			tn, _ := doc.TagName()
			switch string(tn) {
			case "body":
				inBody = false
			case "script", "style", "noscript":
				skip = false
			}
		}
	}
	d.Text = strings.TrimSpace(text.String())
	if d.Text == "" && d.Title == "" {
		return false, errors.New("no content found")
	}
	return false, nil
}

func (e *defaultExtractor) Preview(d *document.Document) (types.PreviewResponse, bool, error) {
	return types.PreviewResponse{Content: d.Text}, false, nil
}

func (e *readabilityExtractor) Name() string {
	return "Readability"
}

func (e *readabilityExtractor) Match(_ *document.Document) bool {
	return true
}

func (e *readabilityExtractor) Extract(d *document.Document) (bool, error) {
	r := bytes.NewReader([]byte(d.HTML))

	u, err := url.Parse(d.URL)
	if err != nil {
		return false, err
	}
	a, err := readability.FromReader(r, u)
	if err != nil {
		return true, err
	}
	buf := bytes.NewBuffer(nil)
	if err := a.RenderText(buf); err != nil {
		return true, err
	}
	d.Text = buf.String()
	d.Title = a.Title()
	d.SetFaviconURL(a.Favicon())
	return false, nil
}

func (e *readabilityExtractor) Preview(d *document.Document) (types.PreviewResponse, bool, error) {
	r := bytes.NewReader([]byte(d.HTML))
	u, err := url.Parse(d.URL)
	if err != nil {
		return types.PreviewResponse{}, false, err
	}
	a, err := readability.FromReader(r, u)
	if err != nil {
		return types.PreviewResponse{}, true, err
	}
	var htmlContent strings.Builder
	if err := a.RenderHTML(&htmlContent); err != nil {
		return types.PreviewResponse{}, true, err
	}
	return types.PreviewResponse{Content: htmlContent.String()}, false, nil
}
