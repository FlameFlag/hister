package mcp

import _ "embed"

// LLM-facing prose (tool descriptions, server instructions, prompt bodies)
// lives as markdown files under descriptions/ so it's easy to edit with real
// syntax highlighting and diffs cleanly. Iterate the wording before iterating
// the tool set. Everything in here is sent to the model; operator-facing
// strings (errors, logs) stay inline in the Go source.

//go:embed descriptions/search.md
var searchDesc string

//go:embed descriptions/fetch_document.md
var fetchDocumentDesc string

//go:embed descriptions/index_url.md
var indexURLDesc string

//go:embed descriptions/forget.md
var forgetDesc string

//go:embed descriptions/server_instructions.md
var serverInstructions string

//go:embed descriptions/prompt_what_do_i_know.md
var promptWhatDoIKnowTemplate string

//go:embed descriptions/prompt_summarize_recent.md
var promptSummarizeRecentTemplate string
