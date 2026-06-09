[English](https://github.com/min0625/mint/blob/main/README.md) | **繁體中文**

# 🌿 Mint

> Minimalist AI Translation CLI — 極簡，快速，直覺。

[![GitHub Release](https://img.shields.io/github/v/release/min0625/mint?logo=github)](https://github.com/min0625/mint/releases)
[![PyPI](https://img.shields.io/pypi/v/mint-ai?logo=pypi&logoColor=white)](https://pypi.org/project/mint-ai/)
[![npm](https://img.shields.io/npm/v/mint-ai?logo=npm)](https://www.npmjs.com/package/mint-ai)

Mint 是一款單一執行檔的 LLM 驅動命令列翻譯工具。設定兩個環境變數，即可翻譯任何內容 — 檔案、管道輸出或直接輸入文字。內建語言偵測、語法修正、串流輸出與多語言輪換功能。

```bash
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_key

mint -t ja "Good morning"         # おはようございます
echo "早安" | mint -t en          # Good morning
cat document.txt | mint -t fr     # 翻譯整個檔案
```

---

## ✨ 為什麼是 Mint？

- **零設定** — 單一執行檔；API 金鑰透過環境變數管理，不污染設定檔
- **多提供商** — Google Gemini、OpenAI、Anthropic，或本地 Ollama / LM Studio
- **智慧偵測** — 每次呼叫皆自動偵測語言；語言中性的內容（數字、符號）原樣輸出
- **智慧修正** — 輸入語言與目標語言相同？自動修正語法與拼字，而非翻譯
- **串流輸出** — 即時串流回應，翻譯長文不需等待
- **可組合** — 友善的 stdin/stdout 設計；與 `grep`、`sed`、`xargs` 等工具無縫搭配

---

## 📋 安裝

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

### 自動安裝

**macOS / Linux**

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh)"
```

自動偵測作業系統與架構（Linux/macOS、x86_64/arm64），安裝到 `~/.local/bin`。可透過 `MINT_INSTALL_DIR` 覆蓋，或用 `MINT_VERSION=v1.0.0` 指定版本。

**Windows（PowerShell）**

```powershell
irm https://raw.githubusercontent.com/min0625/mint/main/script/install.ps1 | iex
```

安裝到 `$HOME\.local\bin`。可透過 `$env:MINT_INSTALL_DIR` 覆蓋，或用 `$env:MINT_VERSION = 'v1.0.0'` 指定版本。

### go install

```bash
go install github.com/min0625/mint/cmd/mint@latest
```

需要 Go 1.26+。二進位檔會放在 `$GOPATH/bin`（通常是 `~/go/bin`）。

### 手動下載

預編譯的二進位檔位於 [GitHub Releases](https://github.com/min0625/mint/releases)

```bash
mint --version
```

---

## 🚀 快速上手

### 1. 設定提供商

```bash
# Google Gemini（有免費層級 — https://aistudio.google.com/apikey）
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_gemini_api_key

# OpenAI
export MINT_PROVIDER=openai
export MINT_API_KEY=sk-...

# Anthropic
export MINT_PROVIDER=anthropic
export MINT_API_KEY=sk-ant-...

# Ollama（無需 API 金鑰）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=qwen2.5:7b  # 替換為 Ollama 中已載入的任意模型

# LM Studio（無需 API 金鑰）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:1234
export MINT_MODEL_NAME=lmstudio-community/Qwen2.5-7B-Instruct-GGUF  # 替換為 LM Studio 中已載入的任意模型
```

### 2. 翻譯

```bash
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

echo "The quick brown fox" | mint -t fr
cat document.txt | mint -t zh-TW
```

使用 `--verbose` / `-v` 將診斷資訊（偵測語言、提供商、模型）輸出至 stderr：

```bash
mint -t ja -v "Good morning"
# [mint] provider=google-genai model=gemini-3.1-flash-lite detected=en target=ja
# おはようございます
```

### 3. 智慧語言偵測

**自動偵測並翻譯：**

```bash
export MINT_TARGET_LANG=en

mint "早安"   # 偵測中文 → Good morning
```

**語法與拼字修正** — 當輸入語言與目標語言相同時，Mint 自動修正而非翻譯：

```bash
export MINT_TARGET_LANG=en

mint "Good mooorning"          # 偵測英文 → Good morning
mint "She don't know nothing"  # 偵測英文 → She doesn't know anything
mint "i luv coding"            # 偵測英文 → I love coding
```

**語言輪換** — 依序翻譯至列表中的下一個語言，循環進行：

```bash
# 兩個語言
export MINT_TARGET_LANG=en,zh-TW
mint "Hello"   # en → zh-TW: 你好
mint "你好"    # zh-TW → en: Hello

# 三個語言
export MINT_TARGET_LANG=en,zh-TW,ja
mint "Hello"       # en → zh-TW: 你好
mint "你好"        # zh-TW → ja: こんにちは
mint "こんにちは"   # ja → en: Hello
```

---

## 🔑 環境變數

| 變數 | 說明 | 預設值 |
|------|------|--------|
| `MINT_PROVIDER` | `google-genai` \| `openai` \| `anthropic` | — (必填) |
| `MINT_API_KEY` | API 金鑰；使用預設 endpoint 時必填；設定 `MINT_BASE_URL` 時選填（由代理處理認證） | — |
| `MINT_BASE_URL` | 自訂 API base URL（僅填 domain，各提供商自行附加路徑）；搭配 `openai` 可指向 Ollama（`http://localhost:11434`）、LM Studio（`http://localhost:1234`）或任何 OpenAI 相容端點 | 提供商預設 |
| `MINT_MODEL_NAME` | 使用的模型 | `gemini-3.1-flash-lite` / `gpt-4o-mini` / `claude-haiku-4-5` |
| `MINT_TARGET_LANG` | 目標語言，例如 `en` 或 `en,zh-TW,ja` | 系統區域設定 |

---

## 🗺 Roadmap

- [x] 多 LLM 提供商支援（Google Gemini、OpenAI、Anthropic，本地透過 Ollama / LM Studio）
- [x] 透過 `MINT_TARGET_LANG` 實現智慧語言偵測與多語言輪換
- [x] 透過 `--target` / `-t` 旗標明確指定目標語言
- [x] 串流輸出
- [x] GoReleaser 多平台二進位檔發布（Linux / macOS / Windows）
- [ ] 批次翻譯模式
- [ ] 術語表 / 自訂詞典支援
- [ ] 輸出格式選項（純文字、JSON、Markdown）
- [ ] 翻譯結果快取

---

## 📄 License

Apache License 2.0 — 詳見 [LICENSE](https://github.com/min0625/mint/blob/main/LICENSE) 文件。
