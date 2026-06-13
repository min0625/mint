// Copyright 2026 The Mint Authors.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode"

	"github.com/min0625/mint/internal/llm"
	"github.com/min0625/mint/internal/provider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "unknown"
)

// neutralLang is the sentinel the language-detection prompt returns for
// language-neutral input (numbers, symbols, etc.).
const neutralLang = "neutral"

func main() {
	os.Exit(run())
}

func run() int {
	// No request timeout: the CLI waits as long as the backend needs (handy for
	// slow local models). Ctrl+C / SIGTERM cancels the in-flight request.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			// Interrupted by the user — exit quietly with the conventional code.
			return 130
		}

		fmt.Fprintln(os.Stderr, "Error:", err)

		return 1
	}

	return 0
}

func newRootCmd() *cobra.Command {
	v := viper.New()
	v.SetEnvPrefix("MINT")
	v.AutomaticEnv()

	var (
		targetLangFlag string
		verboseFlag    bool
	)

	cmd := &cobra.Command{
		Use:           "mint [text]",
		Short:         "Minimalist AI translation CLI",
		Long:          "Mint is a lightweight, LLM-powered translation tool for the command line.",
		Version:       fmt.Sprintf("%s (commit: %s)", version, commit),
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true, // main() prints errors so Ctrl+C can exit quietly
		RunE: func(cmd *cobra.Command, args []string) error {
			logv := func(format string, a ...any) {
				if verboseFlag {
					fmt.Fprintf(os.Stderr, "[mint] "+format+"\n", a...)
				}
			}

			// Load configuration from environment variables
			cfg := provider.Config{
				Provider:   v.GetString("provider"),
				APIKey:     v.GetString("api_key"),
				BaseURL:    v.GetString("base_url"),
				ModelName:  v.GetString("model_name"),
				TargetLang: v.GetString("target_lang"),
			}

			logv("provider: %s", cfg.Provider)

			if cfg.ModelName != "" {
				logv("model: %s", cfg.ModelName)
			}

			if cfg.BaseURL != "" {
				logv("base_url: %s", cfg.BaseURL)
			}

			ctx := cmd.Context()

			t, err := provider.NewCompleter(cfg)
			if err != nil {
				return err
			}

			// Get input text
			text, err := resolveInput(args)
			if err != nil {
				return err
			}

			// Resolve target languages based on priority
			targetLangs := resolveTargetLangs(targetLangFlag, cfg.TargetLang)

			// Determine the target language.
			//
			// Single target (the common case — always so with --target): skip the
			// separate detection call; the unified rewrite prompt handles every case
			// without needing to know the input language —
			//   - cross-language (en → zh-TW): translate + correct
			//   - same language (zh-TW → zh-TW): correct in place
			//   - same language, different script (zh-CN → zh-TW): convert + correct
			// Language-neutral content is caught by a local heuristic before any LLM
			// call, so we halve latency and token cost for every translation.
			//
			// Rotation (multiple targets): we must know the input language to
			// pick the *next* tag in the list, so detection still runs here.
			var actualTargetLang string

			if len(targetLangs) == 1 {
				// Short-circuit for language-neutral content (no letters): no LLM
				// call needed — output unchanged, same as the multi-target path.
				if isLangNeutral(text) {
					logv("language-neutral content — outputting unchanged")
					fmt.Println(text)

					return nil
				}

				actualTargetLang = targetLangs[0]

				logv("single target — skipping language detection")
			} else {
				inputLang, err := detectLanguage(ctx, t, text)
				if err != nil {
					return fmt.Errorf("language detection failed: %w", err)
				}

				logv("detected input language: %q", inputLang)

				// Language-neutral content (numbers, symbols): output unchanged,
				// no rewrite call needed.
				if inputLang == "" {
					logv("language-neutral content — outputting unchanged")
					fmt.Println(text)

					return nil
				}

				actualTargetLang = determineActualTargetLang(inputLang, targetLangs)
			}

			logv("target language: %s", actualTargetLang)

			// Rewrite the input in the target language, correcting grammar and
			// spelling along the way. Anchoring on the target tag also pins the
			// output script, so the model can't drift into the wrong variant
			// (e.g. Simplified for zh-TW).
			prompt := fmt.Sprintf(
				"Rewrite the text inside <text> tags in %s, correcting any grammar and spelling errors.\n"+
					"Output ONLY the resulting text — no labels, no explanation, no preamble.\n\n"+
					"<text>\n%s\n</text>",
				actualTargetLang, text,
			)

			if err := t.Complete(ctx, prompt, os.Stdout); err != nil {
				return fmt.Errorf("translation failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&targetLangFlag, "target", "t", "", "target language (BCP-47 tag, e.g. ja, zh-TW, fr)")
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "print diagnostic info to stderr")

	return cmd
}

// resolveInput returns the text to translate from positional args or stdin.
func resolveInput(args []string) (string, error) {
	if len(args) > 0 {
		if strings.TrimSpace(args[0]) == "" {
			return "", errors.New("no input text provided")
		}

		return args[0], nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	text := strings.TrimRight(string(data), "\n")
	if strings.TrimSpace(text) == "" {
		return "", errors.New("no input text provided")
	}

	return text, nil
}

// resolveTargetLangs resolves target languages based on priority:
// 1. Flag Lang Code (--target / -t) - single language only
// 2. Config Lang Code (MINT_TARGET_LANG) - single or comma-separated languages
// 3. System Lang Code (OS locale) - single language only
// 4. Default to "en" - single language only
func resolveTargetLangs(flagLang, configLang string) []string {
	// Priority 1: Flag Lang Code (single language only)
	if flagLang != "" {
		// Remove any whitespace and normalize
		flagLang = strings.ToLower(strings.TrimSpace(flagLang))
		// Flag should not contain commas - use only the first part if present
		if first, _, found := strings.Cut(flagLang, ","); found {
			flagLang = first
		}

		return []string{flagLang}
	}

	// Priority 2: Config Lang Code (supports multiple comma-separated languages)
	if configLang != "" {
		langs := strings.Split(configLang, ",")
		for i, lang := range langs {
			langs[i] = strings.ToLower(strings.TrimSpace(lang))
		}

		return langs
	}

	// Priority 3: System Lang Code (single language only)
	systemLang := getSystemLanguage()
	if systemLang != "" {
		return []string{strings.ToLower(systemLang)}
	}

	// Priority 4: Default to "en" (single language only)
	return []string{"en"}
}

// getSystemLanguage gets the system language from the OS locale.
func getSystemLanguage() string {
	for _, env := range []string{"LANG", "LC_ALL"} {
		if lang := os.Getenv(env); lang != "" {
			// Strip encoding suffix before cutting on "_":
			// "C.UTF-8" → "C"; "en_US.UTF-8" → "en_US" (no change here)
			lang, _, _ = strings.Cut(lang, ".")
			// Extract primary language subtag: "en_US" → "en"; ignore "C" / "POSIX"
			code, _, _ := strings.Cut(lang, "_")
			if code == "" || code == "C" || code == "POSIX" {
				continue
			}

			return code
		}
	}

	return ""
}

// isLangNeutral reports whether text contains no letters (pure numbers, symbols,
// punctuation, whitespace, etc.) and therefore needs no translation.
func isLangNeutral(text string) bool {
	for _, r := range text {
		if unicode.IsLetter(r) {
			return false
		}
	}

	return true
}

// detectLanguage detects the language of the input text.
// Returns empty string if the input is language-neutral (e.g., numbers, symbols).
func detectLanguage(ctx context.Context, t llm.Completer, text string) (string, error) {
	prompt := "Detect the dominant language of the text inside <text> tags.\n" +
		"Reply with ONLY the BCP-47 language tag (e.g. en, zh-TW, ja) — no quotes, no punctuation, no explanation.\n" +
		"If the text contains only numbers, symbols, or other language-neutral content, reply with: neutral\n\n" +
		"<text>\n" + text + "\n</text>"

	var buf bytes.Buffer
	if err := t.Complete(ctx, prompt, &buf); err != nil {
		return "", err
	}

	lang := normalizeDetectedLang(buf.String())
	if lang == neutralLang {
		return "", nil
	}

	return lang, nil
}

// normalizeDetectedLang coerces the model's free-form reply into a bare
// language tag. Models occasionally wrap the tag in quotes, add trailing
// punctuation, or append an explanation; we keep only the first token and
// strip surrounding noise so downstream tag comparisons stay reliable.
func normalizeDetectedLang(raw string) string {
	lang := strings.ToLower(strings.TrimSpace(raw))

	// Keep only the first whitespace-delimited token (drops any explanation).
	if i := strings.IndexFunc(lang, unicode.IsSpace); i != -1 {
		lang = lang[:i]
	}

	// Strip surrounding quotes/backticks and trailing punctuation.
	return strings.Trim(lang, "\"'`.,!?;:")
}

// langMatches reports whether two BCP-47 tags should occupy the same slot during
// target-language rotation. Exact matches aside, tags sharing the same primary
// subtag (the part before the first "-") are treated as equivalent so that, for
// example, "zh-HK" and "zh-TW" rotate as one slot. This is purely a rotation
// concern — the actual rewrite always targets the configured tag, correcting
// grammar regardless of how close the input language already is.
func langMatches(a, b string) bool {
	if a == b {
		return true
	}

	primaryA, _, _ := strings.Cut(a, "-")
	primaryB, _, _ := strings.Cut(b, "-")

	return primaryA == primaryB
}

// determineActualTargetLang determines the actual target language based on input language and configured targets.
// If input language matches a language in the list:
//   - translate into the next language in the list (wraps around to the first if at the end)
//
// If input language does not match any language in the list:
//   - translate into the first language in the list
//
// If the resolved target language matches the input language:
//   - return for grammar correction
func determineActualTargetLang(inputLang string, targetLangs []string) string {
	if len(targetLangs) == 0 {
		return "en"
	}

	inputLang = strings.ToLower(inputLang)

	// Single target language: use it directly
	if len(targetLangs) == 1 {
		return targetLangs[0]
	}

	// Multiple target languages: find the next one after input language.
	// Use langMatches so that BCP-47 variants sharing the same primary subtag
	// (e.g. "zh-HK" and "zh-TW") are treated as equivalent.
	for i, lang := range targetLangs {
		if langMatches(lang, inputLang) {
			// Found input language, return the next one (wrap around if necessary)
			return targetLangs[(i+1)%len(targetLangs)]
		}
	}

	// Input language not in the list, return the first target language
	return targetLangs[0]
}
