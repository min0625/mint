// Copyright 2026 The Mint Authors.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/min0625/mint/internal/gemini"
	"github.com/min0625/mint/internal/translator"
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
			apiKey := v.GetString("gemini_api_key")
			if apiKey == "" {
				return errors.New("MINT_GEMINI_API_KEY environment variable is not set")
			}

			text, err := resolveInput(args)
			if err != nil {
				return err
			}

			var t translator.Translator = gemini.New(apiKey)

			result, err := t.Translate(context.Background(), text, toLang)
			if err != nil {
				return fmt.Errorf("translation failed: %w", err)
			}

			fmt.Println(result)

			return nil
		},
	}

	cmd.Flags().StringVar(&toLang, "to", "", "target language (BCP-47 tag, e.g. ja, zh-TW, fr)")
	_ = cmd.MarkFlagRequired("to")

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
