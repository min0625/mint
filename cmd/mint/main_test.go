// Copyright 2026 The Mint Authors.
package main

import (
	"slices"
	"testing"
)

const (
	langZhTW = "zh-TW"
	langZhTw = "zh-tw"
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
		got, err := resolveInput([]string{"hello"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})

	t.Run("blank positional arg is rejected", func(t *testing.T) {
		if _, err := resolveInput([]string{"   "}); err == nil {
			t.Error("expected error for blank arg, got nil")
		}
	})
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
