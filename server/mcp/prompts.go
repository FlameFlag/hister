package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	promptWhatDoIKnow     = "what_do_i_know_about"
	promptSummarizeRecent = "summarize_recent"
)

func registerPrompts(srv *mcpsdk.Server) {
	srv.AddPrompt(&mcpsdk.Prompt{
		Name:        promptWhatDoIKnow,
		Title:       "What do I know about…",
		Description: "Surface everything the user's local index has on a topic.",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "topic", Description: "the topic to look up", Required: true},
		},
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		topic := strings.TrimSpace(req.Params.Arguments["topic"])
		if topic == "" {
			return nil, fmt.Errorf("argument %q is required", "topic")
		}
		text := fmt.Sprintf(promptWhatDoIKnowTemplate, topic, topic)
		return userPromptResult("Summarise the user's index coverage on a topic.", text), nil
	})

	srv.AddPrompt(&mcpsdk.Prompt{
		Name:        promptSummarizeRecent,
		Title:       "Summarise recent browsing",
		Description: "Summarise what the user indexed recently, optionally filtered by topic.",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "days", Description: "how many days back to cover (default 7)"},
			{Name: "topic", Description: "optional topic filter"},
		},
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		days := 7
		if v := strings.TrimSpace(req.Params.Arguments["days"]); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("days must be a positive integer, got %q", v)
			}
			days = n
		}
		topic := strings.TrimSpace(req.Params.Arguments["topic"])
		searchHint := `text="*"` // broad sweep; rely on date filter to bound it.
		if topic != "" {
			searchHint = fmt.Sprintf("text=%q", topic)
		}
		text := fmt.Sprintf(promptSummarizeRecentTemplate, days, topicSuffix(topic), days, searchHint)
		return userPromptResult("Summarise recent browsing.", text), nil
	})
}

// userPromptResult wraps a single user-role text message in a GetPromptResult.
// Every prompt in this package is shaped the same way: one message to the
// LLM, so the wire shape lives here and the handlers only produce text.
func userPromptResult(description, text string) *mcpsdk.GetPromptResult {
	return &mcpsdk.GetPromptResult{
		Description: description,
		Messages: []*mcpsdk.PromptMessage{{
			Role:    "user",
			Content: &mcpsdk.TextContent{Text: text},
		}},
	}
}

func topicSuffix(topic string) string {
	if topic == "" {
		return ""
	}
	return fmt.Sprintf(" about %q", topic)
}
