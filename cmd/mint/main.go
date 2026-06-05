// Copyright 2026 The Mint Authors.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/min0625/mint/internal/provider"
	"github.com/min0625/mint/internal/translator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	v := viper.New()
	v.SetEnvPrefix("MINT")
	v.AutomaticEnv()

	var targetLangFlag string

	cmd := &cobra.Command{
		Use:          "mint [text]",
		Short:        "Minimalist AI translation CLI",
		Long:         "Mint is a lightweight, LLM-powered translation tool for the command line.",
		Version:      fmt.Sprintf("%s (commit: %s)", version, commit),
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, args []string) error {
			// Load configuration from environment variables
			cfg := provider.Config{
				Provider:   v.GetString("provider"),
				APIKey:     v.GetString("api_key"),
				BaseURL:    v.GetString("base_url"),
				ModelName:  v.GetString("model_name"),
				TargetLang: v.GetString("target_lang"),
			}

			// Create translator
			t, err := provider.NewTranslator(context.Background(), cfg)
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

			// Detect input language
			inputLang, err := detectLanguage(context.Background(), t, text)
			if err != nil {
				return fmt.Errorf("language detection failed: %w", err)
			}

			// If language-neutral content, output unchanged
			if inputLang == "" {
				fmt.Println(text)
				return nil
			}

			// Determine actual target language
			actualTargetLang := determineActualTargetLang(inputLang, targetLangs)

			// Generate appropriate prompt and perform translation
			var prompt string
			if inputLang == actualTargetLang {
				// Same language: grammar and spelling correction
				prompt = fmt.Sprintf(
					"Fix any grammar and spelling errors in the text inside <text> tags.\n"+
						"Output ONLY the corrected text — no labels, no explanation, no preamble.\n\n"+
						"<text>\n%s\n</text>",
					text,
				)
			} else {
				// Different languages: translation
				prompt = fmt.Sprintf(
					"Translate the text inside <text> tags from %s to %s.\n"+
						"Output ONLY the translated text — no labels, no explanation, no preamble.\n\n"+
						"<text>\n%s\n</text>",
					inputLang, actualTargetLang, text,
				)
			}

			// Perform translation
			result, err := t.Translate(context.Background(), prompt, actualTargetLang)
			if err != nil {
				return fmt.Errorf("translation failed: %w", err)
			}

			fmt.Println(result)

			return nil
		},
	}

	cmd.Flags().StringVarP(&targetLangFlag, "target", "t", "", "target language (BCP-47 tag, e.g. ja, zh-TW, fr)")

	return cmd
}

// resolveInput returns the text to translate from positional args or stdin.
func resolveInput(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	text := strings.TrimRight(string(data), "\n")
	if text == "" {
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
		if idx := strings.Index(flagLang, ","); idx != -1 {
			flagLang = flagLang[:idx]
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
	// Try LANG environment variable
	if lang := os.Getenv("LANG"); lang != "" {
		// Extract language code (e.g., "en_US.UTF-8" -> "en")
		if parts := strings.Split(lang, "_"); len(parts) > 0 {
			return parts[0]
		}
	}

	// Try LC_ALL environment variable
	if lang := os.Getenv("LC_ALL"); lang != "" {
		if parts := strings.Split(lang, "_"); len(parts) > 0 {
			return parts[0]
		}
	}

	return ""
}

// detectLanguage detects the language of the input text.
// Returns empty string if the input is language-neutral (e.g., numbers, symbols).
func detectLanguage(ctx context.Context, t translator.Translator, text string) (string, error) {
	// Use LLM to detect language
	prompt := "Detect the dominant language of the text inside <text> tags.\n" +
		"Reply with ONLY the BCP-47 language tag (e.g. en, zh-TW, ja) — no quotes, no punctuation, no explanation.\n" +
		"If the text contains only numbers, symbols, or other language-neutral content, reply with: neutral\n\n" +
		"<text>\n" + text + "\n</text>"

	result, err := t.Translate(ctx, prompt, "en")
	if err != nil {
		return "", err
	}

	lang := strings.ToLower(strings.TrimSpace(result))
	if lang == "neutral" {
		return "", nil
	}

	return lang, nil
}

// langMatches returns true if two BCP-47 language tags are considered equivalent.
// Exact matches are preferred; otherwise tags sharing the same primary language
// subtag (the part before the first "-") are treated as equivalent so that, for
// example, "zh-HK" matches "zh-TW" because both share the primary subtag "zh".
func langMatches(a, b string) bool {
	if a == b {
		return true
	}

	primaryA := strings.SplitN(a, "-", 2)[0]
	primaryB := strings.SplitN(b, "-", 2)[0]

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
