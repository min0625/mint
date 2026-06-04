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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = ""
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

	var toLang string

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
				Provider:          v.GetString("provider"),
				APIKey:            v.GetString("api_key"),
				BaseURL:           v.GetString("base_url"),
				ModelName:         v.GetString("model_name"),
				PrimaryLanguage:   v.GetString("primary_language"),
				SecondaryLanguage: v.GetString("secondary_language"),
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

			// Determine target language
			targetLang := toLang
			if targetLang == "" {
				if cfg.PrimaryLanguage == "" {
					return errors.New("--to flag is required or MINT_PRIMARY_LANGUAGE environment variable must be set")
				}
				// Use smart language detection
				primaryLang := strings.ToLower(cfg.PrimaryLanguage)

				secondaryLang := cfg.SecondaryLanguage
				if secondaryLang == "" {
					secondaryLang = "zh"
				}

				prompt := fmt.Sprintf(
					"Detect the language of the following text.\n"+
						"If it is already %s, translate it to %s.\n"+
						"Otherwise, translate it to %s.\n"+
						"Output only the translation, nothing else.\n\n%s",
					primaryLang, secondaryLang, primaryLang, text,
				)
				text = prompt
				targetLang = primaryLang // Use primary language as fallback indicator
			}

			// Perform translation
			result, err := t.Translate(context.Background(), text, targetLang)
			if err != nil {
				return fmt.Errorf("translation failed: %w", err)
			}

			fmt.Println(result)

			return nil
		},
	}

	cmd.Flags().StringVar(&toLang, "to", "", "target language (BCP-47 tag, e.g. ja, zh-TW, fr)")

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
