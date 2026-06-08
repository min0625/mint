[English](https://github.com/min0625/mint/blob/main/README.md) | **繁體中文**

# 🌿 Mint

> Minimalist AI Translation CLI — 極簡，快速，直覺。

[![GitHub Release](https://img.shields.io/github/v/release/min0625/mint?logo=github)](https://github.com/min0625/mint/releases)
[![PyPI](https://img.shields.io/pypi/v/mint-ai?logo=pypi&logoColor=white)](https://pypi.org/project/mint-ai/)
[![npm](https://img.shields.io/npm/v/mint-ai?logo=npm)](https://www.npmjs.com/package/mint-ai)

Mint 是一款由 LLM 驅動的輕量命令列翻譯工具。
支援 Google Gemini、OpenAI、Anthropic 及本地 Ollama 等多種提供商，
並內建智慧語言偵測，可自動翻譯至指定語言。

---

## ✨ 為什麼是 Mint？

- **極簡** — 單一指令，無多餘選項干擾
- **多提供商** — Google Gemini、OpenAI、Anthropic，或本地 Ollama
- **靈活** — 任何語言對；智慧自動偵測（選用）
- **可組合** — 友善的 stdin/stdout 設計，輕鬆嵌入任何工作流程

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

需要 Go 1.21+。二進位檔會放在 `$GOPATH/bin`（通常是 `~/go/bin`）。

### 手動下載

預編譯的二進位檔位於 [GitHub Releases](https://github.com/min0625/mint/releases)

### 驗證安裝

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

# 本地 Ollama（無需 API 金鑰）
export MINT_PROVIDER=ollama
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=llama2
```

### 2. 翻譯

```bash
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

echo "The quick brown fox" | mint -t fr
cat document.txt | mint -t zh-TW
```

### 3. 智慧語言偵測

```bash
export MINT_TARGET_LANG=en
mint "早安"             # 偵測中文 → 翻譯成 en
mint "Good mooorning"  # 偵測英文 → 修正語法與拼字

# 跨多個目標語言輪換
export MINT_TARGET_LANG=en,zh-TW,ja
mint "Hello"       # → zh-TW
mint "你好"        # → ja
mint "こんにちは"   # → en
```

---

## 🔑 環境變數

| 變數 | 說明 | 預設值 |
|------|------|--------|
| `MINT_PROVIDER` | `google-genai` \| `openai` \| `anthropic` \| `ollama` | — (必填) |
| `MINT_API_KEY` | API 金鑰（`ollama` 除外均需要） | — |
| `MINT_BASE_URL` | 自訂端點（`ollama` 必填） | 提供商預設 |
| `MINT_MODEL_NAME` | 使用的模型 | `gemini-3.1-flash-lite` / `gpt-4o-mini` / `claude-haiku-4-5` / 無 |
| `MINT_TARGET_LANG` | 目標語言，例如 `en` 或 `en,zh-TW,ja` | 系統區域設定或 `en` |

---

## 🎯 設計原則

| 原則 | 說明 |
|------|------|
| 零依賴安裝 | 單一執行檔，開箱即用 |
| 多提供商 | 支援主要 LLM 服務及本地替代方案 |
| 可組合性 | 與 `grep`、`sed`、`xargs` 等工具無縫搭配 |
| 透明輸出 | 結果到 stdout，錯誤到 stderr |
| 環境友善 | 經由環境變數管理 API 金鑰，不污染設定檔 |

---

## 🗺 Roadmap

- [x] 多 LLM 提供商支援（Google Gemini、OpenAI、Anthropic、Ollama）
- [x] 透過 `MINT_TARGET_LANG` 實現智慧語言偵測與多語言輪換
- [x] 透過 `--target` / `-t` 旗標明確指定目標語言
- [x] GoReleaser 多平台二進位檔發布（Linux / macOS / Windows）
- [ ] 批次翻譯模式
- [ ] 術語表 / 自訂詞典支援
- [ ] 輸出格式選項（純文字、JSON、Markdown）
- [ ] 翻譯結果快取

---

## 📄 License

Apache License 2.0 — 詳見 [LICENSE](https://github.com/min0625/mint/blob/main/LICENSE) 文件。
