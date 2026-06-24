[English](https://github.com/min0625/mint/blob/main/README.md) | [繁體中文](https://github.com/min0625/mint/blob/main/README.zh-TW.md) | **日本語**

# 🌿 Mint

> ミニマリスト AI 翻訳 CLI — シンプル、高速、直感的。

[![GitHub Release](https://img.shields.io/github/v/release/min0625/mint?logo=github)](https://github.com/min0625/mint/releases)
[![PyPI](https://img.shields.io/pypi/v/mint-ai?logo=pypi&logoColor=white)](https://pypi.org/project/mint-ai/)
[![npm](https://img.shields.io/npm/v/mint-ai?logo=npm)](https://www.npmjs.com/package/mint-ai)
[![codecov](https://codecov.io/gh/min0625/mint/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/mint)

Mint はシングルバイナリの LLM 駆動翻訳 CLI ツールです。環境変数を 2 つ設定するだけで、ファイル、パイプ出力、インラインテキストなど、コマンドラインから何でも翻訳できます。言語自動検出、文法修正、ストリーミング出力、多言語ローテーション機能を内蔵しています。

```bash
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_key

mint -t ja "Good morning"         # おはようございます
echo "早安" | mint -t en          # Good morning
cat document.txt | mint -t fr     # ファイル全体を翻訳
```

---

## ✨ なぜ Mint？

- **ゼロ設定** — シングルバイナリ；API キーは環境変数で管理、設定ファイル不要
- **マルチプロバイダー** — Google Gemini、OpenAI、Anthropic、またはローカルの Ollama / LM Studio
- **スマート検出** — 毎回呼び出し時に言語を自動検出；数字や記号などの言語中立なコンテンツはそのまま出力
- **スマート修正** — 入力言語がターゲット言語と同じ場合、翻訳ではなく文法・スペルを自動修正
- **ストリーミング** — リアルタイムでストリーミング出力；長文翻訳でも待ち時間なし
- **コンポーザブル** — パイプフレンドリーな stdin/stdout 設計；`grep`、`sed`、`xargs` などのツールとシームレスに連携

---

## 📋 インストール

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

### 自動インストール

**macOS / Linux**

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh)"
```

OS とアーキテクチャ（Linux/macOS、x86_64/arm64）を自動検出し、`~/.local/bin` にインストールします。`MINT_INSTALL_DIR` で上書き、`MINT_VERSION=v1.0.0` でバージョンを固定できます。

**Windows（PowerShell）**

```powershell
irm https://raw.githubusercontent.com/min0625/mint/main/script/install.ps1 | iex
```

アーキテクチャ（x86_64/arm64）を自動検出し、`$HOME\.local\bin` にインストールします。`$env:MINT_INSTALL_DIR` で上書き、`$env:MINT_VERSION = 'v1.0.0'` でバージョンを固定できます。

### go install

```bash
go install github.com/min0625/mint/cmd/mint@latest
```

Go 1.26 以上が必要です。バイナリは `$GOPATH/bin`（通常 `~/go/bin`）に配置されます。

### 手動ダウンロード

ビルド済みバイナリは [GitHub Releases](https://github.com/min0625/mint/releases) からダウンロードできます。

```bash
mint --version
```

---

## 🚀 クイックスタート

### 1. プロバイダーを設定する

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

# Ollama（API キー不要）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=qwen2.5:7b  # Ollama に読み込んだ任意のモデルを使用

# LM Studio（API キー不要）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:1234
export MINT_MODEL_NAME=lmstudio-community/Qwen2.5-7B-Instruct-GGUF  # LM Studio に読み込んだ任意のモデルを使用
```

### 2. 翻訳する

```bash
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

echo "The quick brown fox" | mint -t fr
cat document.txt | mint -t zh-TW
```

`--verbose` / `-v` を使用すると、診断情報（検出言語、プロバイダー、モデル）を stderr に出力します：

```bash
mint -t ja -v "Good morning"
# [mint] provider=google-genai model=gemini-3.1-flash-lite detected=en target=ja
# おはようございます
```

### 3. スマート言語検出

**自動検出して翻訳：**

```bash
export MINT_TARGET_LANG=en

mint "早安"   # 中国語を検出 → Good morning
```

**文法・スペル修正** — 入力言語がターゲット言語と一致する場合、Mint は翻訳ではなく修正を行います：

```bash
export MINT_TARGET_LANG=en

mint "Good mooorning"          # 英語を検出 → Good morning
mint "She don't know nothing"  # 英語を検出 → She doesn't know anything
mint "i luv coding"            # 英語を検出 → I love coding
```

**言語ローテーション** — リスト内の次の言語へ順番に翻訳し、循環します：

```bash
# 2 言語
export MINT_TARGET_LANG=en,zh-TW
mint "Hello"   # en → zh-TW: 你好
mint "你好"    # zh-TW → en: Hello

# 3 言語
export MINT_TARGET_LANG=en,zh-TW,ja
mint "Hello"       # en → zh-TW: 你好
mint "你好"        # zh-TW → ja: こんにちは
mint "こんにちは"   # ja → en: Hello
```

---

## 🔑 環境変数

| 変数 | 説明 | デフォルト |
|------|------|-----------|
| `MINT_PROVIDER` | `google-genai` \| `openai` \| `anthropic` | — (必須) |
| `MINT_API_KEY` | API キー；デフォルトエンドポイント使用時は必須；`MINT_BASE_URL` 設定時は省略可（プロキシが認証を処理） | — |
| `MINT_BASE_URL` | カスタム API ベース URL（ドメインのみ；各プロバイダーが自身のパスを付加）；`openai` と組み合わせて Ollama（`http://localhost:11434`）、LM Studio（`http://localhost:1234`）、または任意の OpenAI 互換エンドポイントを指定可 | プロバイダーデフォルト |
| `MINT_MODEL_NAME` | 使用するモデル | `gemini-3.1-flash-lite` / `gpt-4o-mini` / `claude-haiku-4-5` |
| `MINT_TARGET_LANG` | ターゲット言語、例：`en` または `en,zh-TW,ja` | システムロケール |

---

## 🗺 ロードマップ

- [x] マルチ LLM プロバイダーサポート（Google Gemini、OpenAI、Anthropic、Ollama / LM Studio 経由のローカル）
- [x] `MINT_TARGET_LANG` によるスマート言語検出と多言語ローテーション
- [x] `--target` / `-t` フラグによる明示的なターゲット言語指定
- [x] ストリーミング出力
- [x] GoReleaser マルチプラットフォームバイナリリリース（Linux / macOS / Windows）
- [ ] バッチ翻訳モード
- [ ] 用語集 / カスタム辞書サポート
- [ ] 出力フォーマットオプション（プレーンテキスト、JSON、Markdown）
- [ ] 翻訳結果のキャッシュ

---

## 📄 License

Apache License 2.0 — 詳細は [LICENSE](https://github.com/min0625/mint/blob/main/LICENSE) を参照してください。
