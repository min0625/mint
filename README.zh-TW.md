[English](https://github.com/min0625/mint/blob/main/README.md) | **繁體中文**

# 🌿 Mint

> Minimalist AI Translation CLI — 極簡，快速，直覺。

---

Mint 是一個輕量的命令列翻譯工具，由 LLM 驅動。
支援多個 LLM 服務提供商（Google Gemini、OpenAI、Anthropic）及本地 Ollama 模型，
並提供智慧語言偵測自動翻譯功能。

---

## ✨ 為什麼是 Mint？

現有的翻譯工具，不是太笨重，就是太依賴特定平台。
Mint 的設計哲學只有一句話：**做最少的事，做到最好。**

- **極簡** — 單一指令，無多餘選項干擾
- **快速** — 直接呼叫 LLM API，無中間層延遲
- **多提供商** — 支援 Google Gemini、OpenAI、Anthropic、本地 Ollama 模型
- **靈活** — 支援多種語言對，自由指定目標語言
- **智慧偵測** — 自動偵測輸入語言，根據偏好自動翻譯
- **可組合** — 友善的 stdin/stdout 設計，輕鬆嵌入任何工作流程

---

## 📋 安裝

### pipx（Python 套件索引）

最簡單的安裝方式——Mint 已上架到 PyPI：

```bash
pipx install mint-ai
```

安裝後，命令是 `mint`：

```bash
mint --version
```

### npm（Node 套件管理器）

如果偏好使用 npm：

```bash
npm install -g mint-ai
```

然後使用：

```bash
mint --version
```

### 自動安裝（單行指令）

自動下載最新的二進位檔：

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh)"
```

功能特性：
- 自動偵測您的作業系統和架構（Linux/macOS、x86_64/arm64）
- 驗證 SHA256 校驗和
- 預設安裝到 `~/.local/bin`（可透過 `MINT_INSTALL_DIR` 覆蓋）
- 在需要時顯示 PATH 設定提示
- 支援指定特定版本：`MINT_VERSION=v1.0.0 bash script/install.sh`

### go install

如果已安裝 Go 1.21+：

```bash
go install github.com/min0625/mint/cmd/mint@latest
```

二進位檔會被放在 `$GOPATH/bin` 目錄中（通常是 `~/go/bin`）。

### 從 GitHub Releases 手動下載

直接從 [GitHub Releases](https://github.com/min0625/mint/releases) 下載預編譯的二進位檔：

```bash
# Linux x86_64
curl -L https://github.com/min0625/mint/releases/latest/download/mint_Linux_x86_64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# macOS arm64 (Apple Silicon)
curl -L https://github.com/min0625/mint/releases/latest/download/mint_Darwin_arm64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# macOS x86_64 (Intel)
curl -L https://github.com/min0625/mint/releases/latest/download/mint_Darwin_x86_64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# Windows x86_64 (PowerShell)
# 從 releases 頁面下載 mint_Windows_x86_64.zip 並解壓縮到 PATH 中的目錄
```

### 驗證安裝

```bash
mint --version
```

---

## 🚀 快速上手

### 1. 選擇 LLM 提供商

```bash
# Google Gemini（有免費層級）
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_gemini_api_key
# 申請免費 API 金鑰：https://aistudio.google.com/apikey

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

### 2. 指定目標語言翻譯

```bash
# 指定目標語言（BCP-47 語言標籤），使用 --target 或 -t 旗標
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

# 從 stdin 管線輸入
echo "The quick brown fox" | mint -t fr

# 翻譯整個檔案
cat document.txt | mint -t zh-TW
```

### 3. 智慧語言偵測（選用）

使用 `MINT_TARGET_LANG` 設定目標語言偏好：

```bash
# 單一目標語言
export MINT_TARGET_LANG=en
mint "早安"             # 偵測中文 → 翻譯成 en
mint "Good mooorning"  # 偵測英文 → 進行語法與拼字修正

# 多個目標語言（語言輪換）
export MINT_TARGET_LANG=en,zh-TW,ja

mint "Hello"         # 英文輸入 → 翻譯成 zh-TW（輪換中的下一個）
mint "你好"          # 中文輸入 → 翻譯成 ja（輪換中的下一個）
mint "こんにちは"     # 日文輸入 → 翻譯成 en（環繞回開始）
```

工具會自動偵測輸入語言並應用適當的轉換。

---

## 🔑 環境變數

| 變數 | 說明 | 必填 | 預設值 |
|------|------|------|--------|
| `MINT_PROVIDER` | LLM 提供商：`google-genai`、`openai`、`anthropic`、`ollama` | 是 | — |
| `MINT_API_KEY` | 所選提供商的 API 金鑰 | 條件式* | — |
| `MINT_BASE_URL` | 自訂 API 端點；`ollama` 必填（例如自架或本地服務） | 條件式* | 提供商預設 |
| `MINT_MODEL_NAME` | 指定要使用的 LLM 模型名稱 | 否 | 提供商預設** |
| `MINT_TARGET_LANG` | 目標語言 - 單一或逗號分隔（如 `en`、`en,zh-TW,ja`） | 否 | 系統區域設定或 `en` |

**條件式:* `MINT_API_KEY` 對 `google-genai`、`openai`、`anthropic` 必填；`ollama` 不需要。`MINT_BASE_URL` 對 `ollama` 必填。*
**預設模型:* `google-genai`: `gemini-3.1-flash-lite`，`openai`: `gpt-4o-mini`，`anthropic`: `claude-haiku-4-5`；`ollama` 無預設（須指定）。*

### 語言解析優先順序

工具使用以下優先順序來決定目標語言：

1. **旗標**：`--target` / `-t` CLI 旗標（最高優先順序）
2. **設定**：`MINT_TARGET_LANG` 環境變數
3. **系統**：作業系統區域設定
4. **預設**：`en`（最低優先順序）

---

## 🎯 設計原則

Mint 遵循 Unix 哲學——**只做一件事，並把它做好。**

| 原則 | 說明 |
|------|------|
| 零依賴安裝 | 單一執行檔，開箱即用 |
| 多提供商支援 | 支援主要 LLM 服務及本地替代方案 |
| 可組合性 | 與 `grep`、`sed`、`xargs` 等工具無縫搭配 |
| 透明輸出 | 結果直接輸出至 stdout，錯誤訊息至 stderr |
| 尊重環境 | 透過環境變數管理 API 金鑰，不污染設定檔 |

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
