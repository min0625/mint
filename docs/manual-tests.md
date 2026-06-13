# Manual Test Cases

AI agents and humans can use these cases to verify correct behavior.
Assumes `MINT_PROVIDER` and `MINT_API_KEY` are already set in the environment.

**Debugging tip:** add `-v` / `--verbose` to any command to print diagnostic info to stderr.
Use it first whenever output is unexpected:

Single-target mode (the common case with `--target` or a single `MINT_TARGET_LANG`):

```
[mint] provider: google-genai
[mint] single target — skipping language detection
[mint] target language: zh-TW
```

Multi-target mode (language rotation with `MINT_TARGET_LANG=zh-TW,en,...`):

```
[mint] provider: google-genai
[mint] detected input language: "en"
[mint] target language: zh-TW
```

---

## 1. Single-word translation (`-t` flag)

```sh
mint -t zh-TW "apple"    # 蘋果
mint -t zh-TW "りんご"   # 蘋果
mint -t en    "蘋果"      # apple
mint -t en    "りんご"    # apple
mint -t ja    "apple"     # りんご
mint -t ja    "蘋果"      # りんご
```

## 2. Sentence translation (`-t` flag)

```sh
mint -t zh-TW "This is an apple."     # 這是一顆蘋果。
mint -t zh-TW "これはリンゴです。"    # 這是一顆蘋果。
mint -t en    "這是一顆蘋果。"        # This is an apple.
mint -t en    "これはリンゴです。"    # This is an apple.
mint -t ja    "This is an apple."     # これはリンゴです。
mint -t ja    "這是一顆蘋果。"        # これはリンゴです。
```

## 3. Long-form flags

`--target` and `--verbose` are the long forms of `-t` and `-v`:

```sh
mint --target zh-TW "apple"                    # 蘋果
mint --target zh-TW --verbose "apple"          # 蘋果  (diagnostics on stderr)
```

`-t` accepts only a single language tag; a comma in the value is silently truncated to the first tag:

```sh
mint -t "zh-TW,en" "apple"    # 蘋果  (only zh-TW is used; rotation requires MINT_TARGET_LANG)
```

## 4. Single target language via `MINT_TARGET_LANG`

```sh
export MINT_TARGET_LANG=zh-TW
mint "apple"                   # 蘋果
mint "This is an apple."       # 這是一顆蘋果。
mint "これはリンゴです。"      # 這是一顆蘋果。

export MINT_TARGET_LANG=en
mint "蘋果"                    # apple
mint "這是一顆蘋果。"          # This is an apple.
mint "これはリンゴです。"      # This is an apple.

export MINT_TARGET_LANG=ja
mint "apple"                   # りんご
mint "This is an apple."       # これはリンゴです。
mint "這是一顆蘋果。"          # これはリンゴです。
```

## 5. Same-language: spelling and grammar correction

When the detected input language matches the target, the tool corrects rather than translates.
`-v` shows `single target — skipping language detection` and `target language: en`.

```sh
export MINT_TARGET_LANG=en
mint "This are an apple."      # This is an apple.

export MINT_TARGET_LANG=zh-TW
mint "這事一科蘋果"            # 這是一顆蘋果。

export MINT_TARGET_LANG=ja
mint "これわリンゴです。"      # これはリンゴです。
```

## 6. Same primary subtag → correction / script conversion

BCP-47 tags sharing the same primary subtag (e.g. `zh-HK` and `zh-TW`, both `zh`) are
treated as one slot in language rotation. With a single target the text is always rewritten
in the target tag's standard form.

```sh
# zh-HK input, zh-TW target → standardize to Traditional Chinese (not a translation)
export MINT_TARGET_LANG=zh-TW
mint "這係一個蘋果"            # 這是一個蘋果

# zh-CN (Simplified Chinese) input, zh-TW target → convert script to Traditional Chinese
mint "这是一个苹果"            # 這是一個蘋果
```

`-v` shows `single target — skipping language detection` and `target language: zh-tw` — the
rewrite prompt handles standardization without needing to detect the input language first.

> **Rotation note:** in a multi-language list (e.g. `zh-TW,en`), zh-HK input occupies
> the zh-TW slot and rotates to `en`, not to `zh-TW`. In multi-target mode `-v` confirms:
> `detected input language: "zh-hk"` → `target language: en`. See section 9 below.

## 7. Language-neutral pass-through

Numbers, symbols, and other language-agnostic content are printed unchanged with no LLM call.
`-v` confirms: `language-neutral content — outputting unchanged`.

- **Single-target mode** (`-t` or single `MINT_TARGET_LANG`): a local heuristic (no letters
  in the text) detects neutral content before any LLM call is made.
- **Multi-target mode** (rotation list): language detection runs first; if the model returns
  `neutral`, the text is output immediately with no second LLM call.

```sh
mint -t zh-TW "42"        # 42
mint -t zh-TW "3.14"      # 3.14
mint -t zh-TW "!@#$%"     # !@#$%
mint -t ja    "123-456"   # 123-456
```

## 8. Stdin / pipe input

Text can be piped from stdin instead of passed as a positional argument.

```sh
echo "apple"              | mint -t zh-TW   # 蘋果
echo "This is an apple."  | mint -t en      # This is an apple.
cat file.txt              | mint -t ja      # file contents translated to Japanese

# Multiline input
printf "First line.\nSecond line." | mint -t zh-TW   # two-line Traditional Chinese output
```

## 9. Language rotation — two languages

```sh
export MINT_TARGET_LANG=zh-TW,en

mint "This is an apple."     # 這是一顆蘋果。  (en matched at index 1 → next: zh-TW at index 0)
mint "これはリンゴです。"    # 這是一顆蘋果。  (ja not in list → first: zh-TW)
mint "這是一顆蘋果。"        # This is an apple.  (zh-TW matched at index 0 → next: en at index 1)
```

> **zh-HK edge case:** zh-HK input matches the zh-TW slot (same primary subtag `zh`) and
> therefore rotates to `en`. Use `-v` to confirm:
> `detected input language: "zh-hk"` → `target language: en`.

## 10. Language rotation — three languages (wrap-around)

```sh
export MINT_TARGET_LANG=zh-TW,en,ja

mint "This is an apple."     # これはリンゴです。  (en at index 1 → next: ja at index 2)
mint "這是一顆蘋果。"        # This is an apple.  (zh-TW at index 0 → next: en at index 1)
mint "これはリンゴです。"    # 這是一顆蘋果。  (ja at index 2 → wraps to: zh-TW at index 0)
```

## 11. `-t` flag overrides `MINT_TARGET_LANG`

```sh
export MINT_TARGET_LANG=ja
mint -t zh-TW "This is an apple."    # 這是一顆蘋果。  (flag wins over env var)
```

## 12. Error cases

All errors go to stderr; the process exits with code 1.

```sh
# Empty or whitespace-only input
mint -t zh-TW ""           # Error: no input text provided
mint -t zh-TW "   "        # Error: no input text provided
echo "" | mint -t zh-TW    # Error: no input text provided

# Missing or invalid provider
unset MINT_PROVIDER
mint -t zh-TW "apple"      # Error: MINT_PROVIDER environment variable is required

MINT_PROVIDER=invalid mint -t zh-TW "apple"
# Error: unsupported provider: invalid. Supported: google-genai, openai, anthropic

# Missing API key (no MINT_BASE_URL set)
unset MINT_API_KEY
mint -t zh-TW "apple"      # Error: MINT_API_KEY is required for provider: <provider>
```

Ctrl+C / SIGTERM cancels any in-flight request and exits with code 130.
