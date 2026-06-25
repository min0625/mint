// Copyright 2026 The Mint Authors.
package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
)

// mockCompleter is a test double for llm.Completer.
type mockCompleter struct {
	response string
	err      error
}

func (m *mockCompleter) Complete(_ context.Context, _ string, w io.Writer) error {
	if m.err != nil {
		return m.err
	}

	_, _ = io.WriteString(w, m.response)

	return nil
}

const (
	langZhTW   = "zh-TW"
	langZhTw   = "zh-tw"
	argTarget  = "--target"
	inputHello = "Hello"
	inputhello = "hello"
)

func TestLangMatches(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"en", "en", true},
		{"en", "fr", false},
		{langZhTW, "zh-HK", true},
		{langZhTW, "zh", true},
		{"zh", langZhTW, true},
		{"en-US", "en-GB", true},
		{"en", "en-US", true},
		{"", "", true},
		{"en", "", false},
	}
	for _, tt := range tests {
		if got := langMatches(tt.a, tt.b); got != tt.want {
			t.Errorf("langMatches(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestDetermineActualTargetLang(t *testing.T) {
	tests := []struct {
		name        string
		inputLang   string
		targetLangs []string
		want        string
	}{
		{"empty target list defaults to en", "fr", nil, "en"},
		{"single target returns it directly", "fr", []string{"en"}, "en"},
		{"single target same as input (correction mode)", "en", []string{"en"}, "en"},
		{"input matches first returns second", "en", []string{"en", langZhTW}, langZhTW},
		{"input matches middle returns next", langZhTW, []string{"en", langZhTW, "ja"}, "ja"},
		{"input matches last wraps to first", "ja", []string{"en", langZhTW, "ja"}, "en"},
		{"input not in list returns first", "fr", []string{"en", langZhTW}, "en"},
		{"match by primary subtag zh-HK → zh-TW slot", "zh-HK", []string{"en", langZhTW, "ja"}, "ja"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineActualTargetLang(tt.inputLang, tt.targetLangs)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveTargetLangs(t *testing.T) {
	tests := []struct {
		name       string
		flagLang   string
		configLang string
		want       []string
	}{
		{"flag takes priority over config", "ja", "en,zh-TW", []string{"ja"}},
		{"flag with comma uses first part only", "en,zh-TW", "", []string{"en"}},
		{"flag normalized to lowercase", "ZH-TW", "", []string{langZhTw}},
		{"flag trimmed of whitespace", "  fr  ", "", []string{"fr"}},
		{"config single lang", "", "fr", []string{"fr"}},
		{"config multiple langs", "", "en,zh-TW,ja", []string{"en", langZhTw, "ja"}},
		{"config langs trimmed and lowercased", "", " EN , ZH-TW ", []string{"en", langZhTw}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveTargetLangs(tt.flagLang, tt.configLang)
			if !slices.Equal(got, tt.want) {
				t.Errorf("resolveTargetLangs(%q, %q) = %v, want %v", tt.flagLang, tt.configLang, got, tt.want)
			}
		})
	}
}

func TestResolveTargetLangsSystemFallback(t *testing.T) {
	t.Setenv("LANG", "ja_JP.UTF-8")
	t.Setenv("LC_ALL", "")

	got := resolveTargetLangs("", "")
	if !slices.Equal(got, []string{"ja"}) {
		t.Errorf("expected [ja], got %v", got)
	}
}

func TestResolveTargetLangsDefaultEn(t *testing.T) {
	t.Setenv("LANG", "")
	t.Setenv("LC_ALL", "")

	got := resolveTargetLangs("", "")
	if !slices.Equal(got, []string{"en"}) {
		t.Errorf("expected [en], got %v", got)
	}
}

func TestNormalizeDetectedLang(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"plain tag", "en", "en"},
		{"trims whitespace", "  zh-TW \n", langZhTw},
		{"lowercases", "JA", "ja"},
		{"strips quotes", `"fr"`, "fr"},
		{"strips backticks", "`de`", "de"},
		{"strips trailing punctuation", "en.", "en"},
		{"keeps first token only", "en (English)", "en"},
		{"explanatory sentence", "The language is fr", "the"},
		{"neutral passes through", neutralLang, neutralLang},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeDetectedLang(tt.raw); got != tt.want {
				t.Errorf("normalizeDetectedLang(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestResolveInput(t *testing.T) {
	t.Run("positional arg used directly", func(t *testing.T) {
		got, err := resolveInput([]string{inputhello})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != inputhello {
			t.Errorf("got %q, want %q", got, inputhello)
		}
	})

	t.Run("blank positional arg is rejected", func(t *testing.T) {
		if _, err := resolveInput([]string{"   "}); err == nil {
			t.Error("expected error for blank arg, got nil")
		}
	})
}

func TestIsLangNeutral(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"12345", true},
		{"!@#$%", true},
		{"   ", true},
		{"", true},
		{"123 + 456 = 579", true},
		{"hello", false},
		{"你好", false},
		{"hello 123", false},
	}
	for _, tt := range tests {
		if got := isLangNeutral(tt.text); got != tt.want {
			t.Errorf("isLangNeutral(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		response string
		err      error
		wantLang string
		wantErr  bool
	}{
		{"english", "en", nil, "en", false},
		{"chinese traditional", "zh-TW", nil, "zh-tw", false},
		{"neutral returns empty", "neutral", nil, "", false},
		{"model error propagated", "", errors.New("api error"), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCompleter{response: tt.response, err: tt.err}

			lang, err := detectLanguage(context.Background(), mock, "test text")
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if lang != tt.wantLang {
				t.Errorf("lang = %q, want %q", lang, tt.wantLang)
			}
		})
	}
}

func TestGetSystemLanguage(t *testing.T) {
	tests := []struct {
		name  string
		lang  string
		lcAll string
		want  string
	}{
		{"typical LANG", "en_US.UTF-8", "", "en"},
		{"zh locale", "zh_TW.UTF-8", "", "zh"},
		{"lang without region", "en", "", "en"},
		{"C locale skipped, uses LC_ALL", "C", "fr_FR.UTF-8", "fr"},
		{"C.UTF-8 locale skipped, uses LC_ALL", "C.UTF-8", "fr_FR.UTF-8", "fr"},
		{"POSIX locale skipped, uses LC_ALL", "POSIX", "de_DE.UTF-8", "de"},
		{"LC_ALL used when LANG empty", "", "ja_JP.UTF-8", "ja"},
		{"both empty returns empty string", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LANG", tt.lang)
			t.Setenv("LC_ALL", tt.lcAll)

			if got := getSystemLanguage(); got != tt.want {
				t.Errorf("getSystemLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveInputFromStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stdin

	os.Stdin = r
	defer func() { os.Stdin = old; _ = r.Close() }()

	_, _ = io.WriteString(w, "hello from stdin\n")
	_ = w.Close()

	got, err := resolveInput(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "hello from stdin" {
		t.Errorf("got %q, want %q", got, "hello from stdin")
	}
}

func TestResolveInputStdinBlank(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stdin

	os.Stdin = r
	defer func() { os.Stdin = old; _ = r.Close() }()

	_, _ = io.WriteString(w, "   \n")
	_ = w.Close()

	if _, err := resolveInput(nil); err == nil {
		t.Error("expected error for blank stdin input")
	}
}

// captureStdout replaces os.Stdout with a pipe and returns a function that
// restores os.Stdout and returns whatever was written.
func captureStdout(t *testing.T) func() string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	os.Stdout = w

	return func() string {
		_ = w.Close()

		os.Stdout = old

		var sb strings.Builder

		_, _ = io.Copy(&sb, r)
		_ = r.Close()

		return sb.String()
	}
}

func TestNewRootCmdLangNeutral(t *testing.T) {
	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")

	flush := captureStdout(t)

	cmd := newRootCmd()
	cmd.SetArgs([]string{argTarget, "en", "12345"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := flush()
	if !strings.Contains(out, "12345") {
		t.Errorf("expected '12345' in output, got: %q", out)
	}
}

func TestNewRootCmdTranslation(t *testing.T) {
	const sse = "data: {\"choices\":[{\"delta\":{\"content\":\"Bonjour\"}}]}\n\ndata: [DONE]\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, sse)
	}))
	defer srv.Close()

	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")
	t.Setenv("MINT_BASE_URL", srv.URL)
	t.Setenv("MINT_MODEL_NAME", "test-model")

	flush := captureStdout(t)

	cmd := newRootCmd()
	cmd.SetArgs([]string{argTarget, "fr", "--verbose", inputHello})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		_ = flush()

		t.Fatalf("unexpected error: %v", err)
	}

	out := flush()
	if !strings.Contains(out, "Bonjour") {
		t.Errorf("expected 'Bonjour' in output, got: %q", out)
	}
}

func TestNewRootCmdTranslationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"server error"}`)
	}))
	defer srv.Close()

	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")
	t.Setenv("MINT_BASE_URL", srv.URL)
	t.Setenv("MINT_MODEL_NAME", "test-model")

	cmd := newRootCmd()
	cmd.SetArgs([]string{argTarget, "fr", inputHello})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "translation failed") {
		t.Errorf("error %q does not mention 'translation failed'", err.Error())
	}
}

func TestNewRootCmdMultiTarget(t *testing.T) {
	// Both calls (detection + translation) return a valid SSE response.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"en\"}}]}\n\ndata: [DONE]\n")
	}))
	defer srv.Close()

	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")
	t.Setenv("MINT_BASE_URL", srv.URL)
	t.Setenv("MINT_MODEL_NAME", "test-model")
	t.Setenv("MINT_TARGET_LANG", "en,fr")

	flush := captureStdout(t)
	defer flush()

	cmd := newRootCmd()
	cmd.SetArgs([]string{"Hello world"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRootCmdMultiTargetNeutral(t *testing.T) {
	// Detection returns "neutral" — output is passed through without a second call.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"neutral\"}}]}\n\ndata: [DONE]\n")
	}))
	defer srv.Close()

	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")
	t.Setenv("MINT_BASE_URL", srv.URL)
	t.Setenv("MINT_MODEL_NAME", "test-model")
	t.Setenv("MINT_TARGET_LANG", "en,fr")

	flush := captureStdout(t)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"12345"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		_ = flush()

		t.Fatalf("unexpected error: %v", err)
	}

	out := flush()
	if !strings.Contains(out, "12345") {
		t.Errorf("expected '12345' in output, got: %q", out)
	}
}

func TestNewRootCmdDetectLangError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"server error"}`)
	}))
	defer srv.Close()

	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")
	t.Setenv("MINT_BASE_URL", srv.URL)
	t.Setenv("MINT_MODEL_NAME", "test-model")
	t.Setenv("MINT_TARGET_LANG", "en,fr")

	cmd := newRootCmd()
	cmd.SetArgs([]string{inputHello})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "language detection failed") {
		t.Errorf("error %q does not mention 'language detection failed'", err.Error())
	}
}

func TestNewRootCmdProviderError(t *testing.T) {
	t.Setenv("MINT_PROVIDER", "invalid-provider")
	t.Setenv("MINT_API_KEY", "test")

	cmd := newRootCmd()
	cmd.SetArgs([]string{inputHello})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Error("expected error for invalid provider, got nil")
	}
}

func TestNewRootCmdNoInput(t *testing.T) {
	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")

	// Empty stdin so resolveInput returns an error.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stdin
	os.Stdin = r

	_ = w.Close()

	defer func() { os.Stdin = old; _ = r.Close() }()

	cmd := newRootCmd()
	cmd.SetArgs([]string{argTarget, "en"})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Error("expected error for no input, got nil")
	}
}

func TestRunSuccess(t *testing.T) {
	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")

	old := os.Args

	os.Args = []string{"mint", argTarget, "en", "12345"}
	defer func() { os.Args = old }()

	flush := captureStdout(t)
	defer flush()

	if code := run(); code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunError(t *testing.T) {
	t.Setenv("MINT_PROVIDER", "")

	old := os.Args

	os.Args = []string{"mint", argTarget, "en", inputhello}
	defer func() { os.Args = old }()

	// Suppress stderr to keep test output clean.
	rr, ww, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	oldErr := os.Stderr

	os.Stderr = ww
	defer func() {
		_ = ww.Close()

		os.Stderr = oldErr
		_, _ = io.Copy(io.Discard, rr)
		_ = rr.Close()
	}()

	if code := run(); code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestResolveSourceLang(t *testing.T) {
	tests := []struct {
		name string
		flag string
		want string
	}{
		{"empty means auto-detect", "", ""},
		{"normalized to lowercase", "FR", "fr"},
		{"trimmed of whitespace", "  ja  ", "ja"},
		{"comma uses first part only", "fr,en", "fr"},
		{"comma with whitespace trimmed", " fr , en ", "fr"},
		{"bcp-47 variant preserved", "zh-TW", langZhTw},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveSourceLang(tt.flag); got != tt.want {
				t.Errorf("resolveSourceLang(%q) = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestBuildRewritePrompt(t *testing.T) {
	// Without a source language: rewrite (model infers input language).
	noSrc := buildRewritePrompt("", "en", "chat")
	if !strings.Contains(noSrc, "Rewrite the text") {
		t.Errorf("expected rewrite phrasing, got: %q", noSrc)
	}

	// The no-source prompt must instruct the model to translate foreign input
	// (not just proof-read it) so providers like Claude don't pass CJK→en input
	// through unchanged.
	if !strings.Contains(noSrc, "in another language") {
		t.Errorf("expected conditional-translate clause without --source, got: %q", noSrc)
	}

	if !strings.Contains(noSrc, "translate it into en") {
		t.Errorf("expected translate instruction without --source, got: %q", noSrc)
	}

	// With a source language: anchor it and force a translation.
	withSrc := buildRewritePrompt("fr", "en", "chat")
	if !strings.Contains(withSrc, "written in fr") {
		t.Errorf("expected source anchor 'written in fr', got: %q", withSrc)
	}

	if !strings.Contains(withSrc, "Translate it into en") {
		t.Errorf("expected translate instruction, got: %q", withSrc)
	}

	// Source == target (exact same tag): no-op translation, so fall back to
	// the rewrite (correction-only) phrasing rather than "translate en→en".
	sameTag := buildRewritePrompt("en", "en", "helo")
	if !strings.Contains(sameTag, "Rewrite the text") {
		t.Errorf("expected rewrite phrasing when source == target, got: %q", sameTag)
	}

	if strings.Contains(sameTag, "written in") {
		t.Errorf("did not expect a source anchor when source == target, got: %q", sameTag)
	}

	// Distinct tags sharing a primary subtag are a deliberate script
	// conversion (zh-CN → zh-TW) and must keep the source anchor.
	scriptConv := buildRewritePrompt("zh-cn", langZhTw, "汉字")
	if !strings.Contains(scriptConv, "written in zh-cn") {
		t.Errorf("expected source anchor for script conversion, got: %q", scriptConv)
	}
}

// TestNewRootCmdSourceLang verifies that --source anchors the rewrite prompt
// with the source language so the model is told to translate from it.
func TestNewRootCmdSourceLang(t *testing.T) {
	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"cat\"}}]}\n\ndata: [DONE]\n")
	}))
	defer srv.Close()

	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")
	t.Setenv("MINT_BASE_URL", srv.URL)
	t.Setenv("MINT_MODEL_NAME", "test-model")

	flush := captureStdout(t)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--source", "fr", argTarget, "en", "chat"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		_ = flush()

		t.Fatalf("unexpected error: %v", err)
	}

	out := flush()
	if !strings.Contains(out, "cat") {
		t.Errorf("expected 'cat' in output, got: %q", out)
	}

	if !strings.Contains(gotBody, "written in fr") {
		t.Errorf("expected request prompt to anchor source 'fr', got body: %q", gotBody)
	}
}

// TestNewRootCmdSourceRotation verifies that an explicit --source skips the
// language-detection call in rotation mode: only the rewrite call is made.
func TestNewRootCmdSourceRotation(t *testing.T) {
	var calls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"\\u4f60\\u597d\"}}]}\n\ndata: [DONE]\n")
	}))
	defer srv.Close()

	t.Setenv("MINT_PROVIDER", "openai")
	t.Setenv("MINT_API_KEY", "test")
	t.Setenv("MINT_BASE_URL", srv.URL)
	t.Setenv("MINT_MODEL_NAME", "test-model")
	t.Setenv("MINT_TARGET_LANG", "en,zh-TW")

	flush := captureStdout(t)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--source", "en", "Hello"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		_ = flush()

		t.Fatalf("unexpected error: %v", err)
	}

	_ = flush()

	if calls != 1 {
		t.Errorf("expected exactly 1 LLM call (no detection), got %d", calls)
	}
}
