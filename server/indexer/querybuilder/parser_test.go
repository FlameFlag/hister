package querybuilder

import "testing"

func Test_tokenize_arabic_word(t *testing.T) {
	tokens, err := Tokenize("سلام")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenWord {
		t.Fatalf("expected TokenWord, got %v", tokens[0].Type)
	}
	if tokens[0].Value != "سلام" {
		t.Fatalf("expected token value %q, got %q", "سلام", tokens[0].Value)
	}
}

func Test_tokenize_english_word(t *testing.T) {
	tokens, err := Tokenize("hello")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenWord {
		t.Fatalf("expected TokenWord, got %v", tokens[0].Type)
	}
	if tokens[0].Value != "hello" {
		t.Fatalf("expected token value %q, got %q", "hello", tokens[0].Value)
	}
}

func Test_tokenize_english_quoted_phrase(t *testing.T) {
	tokens, err := Tokenize(`"hello world"`)
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenQuoted {
		t.Fatalf("expected TokenQuoted, got %v", tokens[0].Type)
	}
	if tokens[0].Value != "hello world" {
		t.Fatalf("expected token value %q, got %q", "hello world", tokens[0].Value)
	}
}

func Test_tokenize_english_alternation(t *testing.T) {
	tokens, err := Tokenize("(hello|world)")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenAlternation {
		t.Fatalf("expected TokenAlternation, got %v", tokens[0].Type)
	}
	if len(tokens[0].Parts) != 2 {
		t.Fatalf("expected 2 alternation parts, got %d", len(tokens[0].Parts))
	}
	if tokens[0].Parts[0].Value != "hello" {
		t.Fatalf("expected first alternation part %q, got %q", "hello", tokens[0].Parts[0].Value)
	}
	if tokens[0].Parts[1].Value != "world" {
		t.Fatalf("expected second alternation part %q, got %q", "world", tokens[0].Parts[1].Value)
	}
}

func Test_tokenize_alternation_keeps_last_arabic_part(t *testing.T) {
	tokens, err := Tokenize("(hello|مرحبا)")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenAlternation {
		t.Fatalf("expected TokenAlternation, got %v", tokens[0].Type)
	}
	if len(tokens[0].Parts) != 2 {
		t.Fatalf("expected 2 alternation parts, got %d", len(tokens[0].Parts))
	}
	if tokens[0].Parts[1].Value != "مرحبا" {
		t.Fatalf("expected last alternation part %q, got %q", "مرحبا", tokens[0].Parts[1].Value)
	}
}
