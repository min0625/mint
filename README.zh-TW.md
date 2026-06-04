[English](./README.md) | **繁體中文**

# 🌿 Mint

> Minimalist AI Translation CLI — 極簡，快速，直覺。

---

Mint 是一個輕量的命令列翻譯工具，由 LLM 驅動。
沒有複雜的設定，沒有多餘的介面——只需一行指令，即可獲得流暢、自然的翻譯結果。

---

## ✨ 為什麼是 Mint？

現有的翻譯工具，不是太笨重，就是太依賴特定平台。
Mint 的設計哲學只有一句話：**做最少的事，做到最好。**

- **極簡** — 單一指令，無多餘選項干擾
- **快速** — 直接呼叫 LLM API，無中間層延遲
- **靈活** — 支援多種語言對，自由指定目標語言
- **可組合** — 友善的 stdin/stdout 設計，輕鬆嵌入任何工作流程

---

## 🚀 快速上手

### 1. 設定 API 金鑰

```bash
export MINT_GEMINI_API_KEY=your_api_key_here
```

至 [Google AI Studio](https://aistudio.google.com/apikey) 免費申請 API 金鑰。

### 2. 開始翻譯

```bash
# 指定目標語言（BCP-47 語言標籤）
mint --to ja "Good morning"

# 從 stdin 管線輸入
echo "The quick brown fox" | mint --to zh-TW

# 翻譯整個檔案
cat document.txt | mint --to fr
```

---

## 🔑 環境變數

| 變數 | 說明 |
|------|------|
| `MINT_GEMINI_API_KEY` | Google Gemini API 金鑰 **（必填）** |

---

## 🎯 設計原則

Mint 遵循 Unix 哲學——**只做一件事，並把它做好。**

| 原則 | 說明 |
|------|------|
| 零依賴安裝 | 單一執行檔，開箱即用 |
| 可組合性 | 與 `grep`、`sed`、`xargs` 等工具無縫搭配 |
| 透明輸出 | 結果直接輸出至 stdout，錯誤訊息至 stderr |
| 尊重環境 | 透過環境變數管理 API 金鑰，不污染設定檔 |

---

## 🗺 Roadmap

- [x] Gemini LLM 後端
- [ ] 更多 LLM 後端支援（OpenAI、Anthropic、Ollama 本地模型）
- [ ] 自動語言偵測
- [ ] 批次翻譯模式
- [ ] 術語表 / 自訂詞典支援
- [ ] 輸出格式選項（純文字、JSON、Markdown）

---

## 📄 License

Apache License 2.0 — 詳見 [LICENSE](./LICENSE) 文件。
