🌐 [その他の言語](https://github.com/min0625/mint/blob/main/LANGUAGES.md)

# 🌿 Mint

> Minimalist AI Translation CLI — シンプル、高速、直感的。

[![GitHub Release](https://img.shields.io/github/v/release/min0625/mint?logo=github)](https://github.com/min0625/mint/releases)
[![PyPI](https://img.shields.io/pypi/v/mint-ai?logo=pypi&logoColor=white)](https://pypi.org/project/mint-ai/)
[![npm](https://img.shields.io/npm/v/mint-ai?logo=npm)](https://www.npmjs.com/package/mint-ai)
[![codecov](https://codecov.io/gh/min0625/mint/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/mint)

Mintは、単一の実行ファイルで動作するLLM駆動のコマンドライン翻訳ツールです。環境変数を2つ設定するだけで、ファイル、パイプ出力、直接入力されたテキストなど、あらゆる内容を翻訳できます。言語自動検出、文法修正、ストリーミング出力、多言語ローテーション機能を内蔵しています。

```bash
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_key

mint -t ja "Good morning"         # おはようございます
echo "早安" | mint -t en          # Good morning
cat document.txt | mint -t fr     # ファイル全体を翻訳
```

---

## ✨ Mintの特徴

- **ゼロ設定** — 単一の実行ファイル。APIキーは環境変数で管理するため、設定ファイルを汚しません。
- **マルチプロバイダー** — Google Gemini、OpenAI、Anthropicのほか、ローカルのOllamaやLM Studioにも対応。
- **スマート検出** — 実行のたびに言語を自動検出。言語に依存しない内容（数字、記号）はそのまま出力します。
- **スマート修正** — 入力言語とターゲット言語が同じ場合は、翻訳ではなく文法やスペルの修正を行います。
- **ストリーミング出力** — レスポンスをリアルタイムでストリーミングするため、長文翻訳でも待たされません。
- **コンポーザブル** — stdin/stdoutを重視した設計により、`grep`、`sed`、`xargs`などのツールとシームレスに連携可能です。
- **セキュア** — system/userメッセージの分離とリクエストごとのランダムnonceデリミタにより、信頼できない入力はモデルの指示から隔離されます。悪意あるコンテンツを翻訳してもLLMの動作をハイジャックすることはできません。

---

## 📋 インストール

### 自動インストール（推奨）

**macOS / Linux**

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh)"
```

OSとアーキテクチャ（Linux/macOS、x86_64/arm64）を自動検出し、`~/.local/bin`にインストールします。`MINT_INSTALL_DIR`でインストール先を変更したり、`MINT_VERSION=v1.0.0`でバージョンを指定したりすることも可能です。

**Windows（PowerShell）**

```powershell
irm https://raw.githubusercontent.com/min0625/mint/main/script/install.ps1 | iex
```

アーキテクチャ（x86_64/arm64）を自動検出し、`$HOME\.local\bin`にインストールします。`$env:MINT_INSTALL_DIR`でインストール先を変更したり、`$env:MINT_VERSION = 'v1.0.0'`でバージョンを指定したりすることも可能です。

### Homebrew (macOS / Linux)

```bash
brew install min0625/tap/mint-ai
```

### pipx

```bash
pipx install mint-ai
```

### npm

```bash
npm install -g mint-ai
```

### 手動ダウンロード

[GitHub Releases](https://github.com/min0625/mint/releases) からお使いのプラットフォーム向けのビルド済みバイナリをダウンロードし、`PATH` の通ったディレクトリに移動してから、動作を確認します：

```bash
mint --version
```

---

## 🚀 クイックスタート

### 1. プロバイダーの設定

```bash
# Google Gemini（無料枠あり — https://aistudio.google.com/apikey）
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_gemini_api_key

# OpenAI
export MINT_PROVIDER=openai
export MINT_API_KEY=sk-...

# Anthropic
export MINT_PROVIDER=anthropic
export MINT_API_KEY=sk-ant-...

# Ollama（APIキー不要）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=qwen2.5:7b  # Ollamaでロード済みのモデル名に変更

# LM Studio（APIキー不要）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:1234
export MINT_MODEL_NAME=lmstudio-community/Qwen2.5-7B-Instruct-GGUF  # LM Studioでロード済みのモデル名に変更
```

### 2. 翻訳の実行

```bash
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

echo "The quick brown fox" | mint -t fr
cat document.txt | mint -t zh-TW
```

`--verbose` / `-v`（または `MINT_VERBOSE=true`）を使用すると、診断情報とトークン使用量がstderrに出力されます。

```bash
mint -t ja -v "Good morning"
# [mint] provider: google-genai
# [mint] model: gemini-3.1-flash-lite
# [mint] single target — skipping language detection
# [mint] target language: ja
# おはようございます
# [mint] tokens: 113 in / 2 out
```

**典型的なトークン使用量**（`gemini-3.1-flash-lite` での実測値）：

| モード | 入力 | API呼び出し回数 | 入力トークン | 出力トークン |
|--------|------|---------------|------------|------------|
| 単一ターゲット（`-t` または単一 `MINT_TARGET_LANG`） | 短い単語・文章 | 1 | ~110–130 | ~1–15 |
| 単一ターゲット | 長い文章（`testdata/sample.txt`） | 1 | ~465–470 | ~450–560 |
| 多言語ローテーション（カンマ区切り `MINT_TARGET_LANG`） | 短い文章 | 2 | ~250–260 | ~2–8 |
| 明示的ソース `-s` + ローテーション | 短い文章 | 1 | ~105–120 | ~1–2 |

> トークン数は入力の長さによって変わります。また出力トークンは言語によっても異なり、日本語や中国語は英語より多くなる傾向があります。

**100万トークンで何回翻訳できる？**（入力＋出力の合計、上記の実測使用量から算出）：

| 入力 | 1回あたりのトークン数 | 100万トークンでの翻訳回数 |
|------|---------------------|------------------------|
| 短い単語・文章 | 約120 | 約8,000回 |
| 300語の文章 | 約1,000 | 約1,000件 |

> 回数は入力と出力のトークンを合計したものです。各プロバイダーは入力と出力を別々に課金し、多くが無料枠を提供しています——実際の料金は各プロバイダーの価格ページをご確認ください。GoogleのGeminiは[Google AI Studio](https://aistudio.google.com/apikey)の無料枠をクレジットカードなしで利用できます。

`--source` / `-s` で**ソース言語を強制指定**すると、ターゲット言語としても有効な入力（言語をまたぐ同形異義語、ローマ字表記のテキスト）を翻訳できます：

```bash
mint -s fr -t en "pain"          # フランス語 → bread（-s なしでは英語の "pain" として扱われる）
mint -s ja -t en "konnichiwa"    # ローマ字の日本語 → hello
```

### 3. インテリジェント言語検出

**自動検出と翻訳：**

```bash
export MINT_TARGET_LANG=en

mint "早安"   # 中国語を検出 → Good morning
```

**文法とスペルの修正** — 入力言語とターゲット言語が同じ場合、Mintは翻訳ではなく修正を行います：

```bash
export MINT_TARGET_LANG=en

mint "Good mooorning"          # 英語を検出 → Good morning
mint "She don't know nothing"  # 英語を検出 → She doesn't know anything
mint "i luv coding"            # 英語を検出 → I love coding
```

**言語ローテーション** — 指定した言語リストを順次切り替えて翻訳します：

```bash
# 2言語の場合
export MINT_TARGET_LANG=en,zh-TW
mint "Hello"   # en → zh-TW: 你好
mint "你好"    # zh-TW → en: Hello

# 3言語の場合
export MINT_TARGET_LANG=en,zh-TW,ja
mint "Hello"       # en → zh-TW: 你好
mint "你好"        # zh-TW → ja: こんにちは
mint "こんにちは"   # ja → en: Hello
```

---

## 🔑 環境変数

| 変数 | 説明 | デフォルト値 |
|------|------|--------|
| `MINT_PROVIDER` | `google-genai` \| `openai` \| `anthropic` | — (必須) |
| `MINT_API_KEY` | APIキー。デフォルトのエンドポイント使用時は必須。`MINT_BASE_URL`設定時は任意（プロキシ側で認証処理する場合） | — |
| `MINT_BASE_URL` | カスタムAPIベースURL（ドメインのみ指定、パスは各プロバイダーが自動付与）。`openai`と組み合わせることで、Ollama（`http://localhost:11434`）、LM Studio（`http://localhost:1234`）、またはOpenAI互換エンドポイントを指定可能 | プロバイダーのデフォルト |
| `MINT_MODEL_NAME` | 使用するモデル名 | `gemini-3.1-flash-lite` / `gpt-4o-mini` / `claude-haiku-4-5` |
| `MINT_TARGET_LANG` | ターゲット言語（例: `en` または `en,zh-TW,ja`） | システムのロケール設定 |
| `MINT_VERBOSE` | `true`に設定すると詳細な診断出力が有効になります（`--verbose`相当） | `false` |

---

## 🚩 CLIフラグ

| フラグ | 省略形 | 説明 |
|--------|--------|------|
| `--target <lang>` | `-t` | ターゲット言語（BCP-47タグ、例: `ja`、`zh-TW`、`fr`）。`MINT_TARGET_LANG`を上書きします。 |
| `--source <lang>` | `-s` | ソース言語（BCP-47タグ）。自動検出をスキップし、この言語からの翻訳を強制します。 |
| `--verbose` | `-v` | 診断情報とトークン使用量をstderrに出力します。`MINT_VERBOSE=true`でも有効にできます。 |
| `--version` | | バージョンを表示して終了します。 |

---

## 📅 ロードマップ

- [x] 複数のLLMプロバイダー対応（Google Gemini、OpenAI、Anthropic、ローカルのOllama / LM Studio）
- [x] `MINT_TARGET_LANG` による言語自動検出と多言語ローテーション
- [x] `--target` / `-t` フラグによるターゲット言語の明示的指定
- [x] `--source` / `-s` フラグによるソース言語の明示的指定
- [x] ストリーミング出力
- [x] GoReleaserによるマルチプラットフォームバイナリ配布（Linux / macOS / Windows）
- [ ] バッチ翻訳モード
- [ ] 用語集 / カスタム辞書サポート
- [ ] 出力フォーマットオプション（プレーンテキスト、JSON、Markdown）
- [ ] 翻訳結果のキャッシュ

---

## 📄 ライセンス

Apache License 2.0 — 詳細は [LICENSE](https://github.com/min0625/mint/blob/main/LICENSE) ファイルを参照してください。
