🌐 [其他语言](https://github.com/min0625/mint/blob/main/LANGUAGES.md)

# 🌿 Mint

> Minimalist AI Translation CLI — 极简，快速，直观。

[![GitHub Release](https://img.shields.io/github/v/release/min0625/mint?logo=github)](https://github.com/min0625/mint/releases)
[![PyPI](https://img.shields.io/pypi/v/mint-ai?logo=pypi&logoColor=white)](https://pypi.org/project/mint-ai/)
[![npm](https://img.shields.io/npm/v/mint-ai?logo=npm)](https://www.npmjs.com/package/mint-ai)
[![codecov](https://codecov.io/gh/min0625/mint/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/mint)

Mint 是一款单一可执行文件、由 LLM 驱动的命令行翻译工具。只需配置两个环境变量，即可在命令行中翻译任意内容——文件、管道输出或直接输入的文本。内置语言检测、语法纠正、流式输出与多语言轮换功能。

```bash
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_key

mint -t ja "Good morning"         # おはようございます
echo "早安" | mint -t en          # Good morning
cat document.txt | mint -t fr     # 翻译整个文件
```

---

## ✨ 为什么选择 Mint？

- **零配置** — 单一可执行文件；API 密钥通过环境变量管理，不产生配置文件污染
- **多提供商** — Google Gemini、OpenAI、Anthropic，或任何 OpenAI 兼容端点（Ollama、LM Studio、OpenRouter、Groq、DeepSeek、llama.cpp 等）
- **智能检测** — 每次调用自动检测语言；语言中性的内容（数字、符号）原样输出
- **智能纠正** — 输入语言与目标语言相同？自动纠正语法与拼写，而非翻译
- **流式输出** — 响应实时流式返回，翻译长文无需等待
- **可组合** — 对 stdin/stdout 友好的设计；与 `grep`、`sed`、`xargs` 等工具无缝配合
- **安全** — 通过 system/user 消息分离与每次请求的随机 nonce 分隔符，将不可信输入与模型指令隔离；翻译恶意内容也无法劫持 LLM 的行为

---

## 📋 安装

### 自动安装（推荐）

**macOS / Linux**

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh)"
```

自动检测操作系统与架构（Linux/macOS、x86_64/arm64），安装到 `~/.local/bin`。可通过 `MINT_INSTALL_DIR` 覆盖安装目录，或用 `MINT_VERSION=v1.0.0` 指定版本。

**Windows（PowerShell）**

```powershell
irm https://raw.githubusercontent.com/min0625/mint/main/script/install.ps1 | iex
```

自动检测架构（x86_64/arm64），安装到 `$HOME\.local\bin`。可通过 `$env:MINT_INSTALL_DIR` 覆盖安装目录，或用 `$env:MINT_VERSION = 'v1.0.0'` 指定版本。

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

### 手动下载

从 [GitHub Releases](https://github.com/min0625/mint/releases) 下载对应平台的预编译二进制文件，移动到 `PATH` 中的目录，然后验证：

```bash
mint --version
```

---

## 🚀 快速上手

### 1. 配置提供商

```bash
# Google Gemini（提供免费额度 — https://aistudio.google.com/apikey）
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_gemini_api_key

# OpenAI
export MINT_PROVIDER=openai
export MINT_API_KEY=sk-...

# Anthropic
export MINT_PROVIDER=anthropic
export MINT_API_KEY=sk-ant-...

# Ollama（无需 API 密钥）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=qwen2.5:7b  # 替换为 Ollama 中已加载的任意模型

# LM Studio（无需 API 密钥）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:1234
export MINT_MODEL_NAME=lmstudio-community/Qwen2.5-7B-Instruct-GGUF  # 替换为 LM Studio 中已加载的任意模型

# llama.cpp 的 llama-server（无需 API 密钥）
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:8080
export MINT_MODEL_NAME=qwen2.5:7b  # 替换为 llama-server 中已加载的模型

# OpenRouter（一个密钥，数百种模型 — https://openrouter.ai/models）
export MINT_PROVIDER=openai
export MINT_BASE_URL=https://openrouter.ai/api
export MINT_API_KEY=sk-or-...
export MINT_MODEL_NAME=openai/gpt-4o-mini

# Groq（推理速度快，提供免费额度）
export MINT_PROVIDER=openai
export MINT_BASE_URL=https://api.groq.com/openai
export MINT_API_KEY=gsk_...
export MINT_MODEL_NAME=llama-3.1-8b-instant

# DeepSeek
export MINT_PROVIDER=openai
export MINT_BASE_URL=https://api.deepseek.com
export MINT_API_KEY=sk-...
export MINT_MODEL_NAME=deepseek-chat
```

### 2. 翻译

```bash
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

echo "The quick brown fox" | mint -t fr
cat document.txt | mint -t zh-TW
```

使用 `--verbose` / `-v`（或 `MINT_VERBOSE=true`）将诊断信息与 token 用量输出到 stderr：

```bash
mint -t ja -v "Good morning"
# [mint] provider: google-genai
# [mint] single target — skipping language detection
# [mint] target language: ja
# おはようございます
# [mint] tokens: 113 in / 2 out
```

**典型 token 用量**（基于 `gemini-3.1-flash-lite` 实测）：

| 模式 | 输入 | API 调用次数 | 输入 tokens | 输出 tokens |
|------|------|-------------|------------|------------|
| 单一目标（`-t` 或单个 `MINT_TARGET_LANG`） | 短词/短句 | 1 | ~110–130 | ~1–15 |
| 单一目标 | 长文章（`testdata/sample.txt`） | 1 | ~465–470 | ~450–560 |
| 多目标轮换（逗号分隔 `MINT_TARGET_LANG`） | 短句 | 2 | ~250–260 | ~2–8 |
| 显式指定来源 `-s` + 轮换 | 短句 | 1 | ~105–120 | ~1–2 |

> Token 数量随输入长度变化；输出 token 也因目标语言而异——日语与中文通常比英语产生更多 token。

**100 万 token 能翻译多少次？**（输入+输出合计，由上述实测用量推算）：

| 输入 | 每次约用 token | 每 100 万 token 可翻译次数 |
|------|---------------|--------------------------|
| 短词或短句 | 约 120 | 约 8,000 次 |
| 300 字文章 | 约 1,000 | 约 1,000 篇 |

> 次数为输入与输出 token 合计。各提供商对输入和输出分别计价，且多数提供免费额度——具体费率请查阅提供商的定价页面。Google Gemini 在 [Google AI Studio](https://aistudio.google.com/apikey) 的免费额度无需绑定信用卡。

使用 `--source` / `-s` **强制指定来源语言**，可翻译那些在目标语言中同样合法的输入（跨语言同形词、罗马音文本）：

```bash
mint -s fr -t en "pain"          # 法语 → bread（不加 -s 会被当作英语的 "pain"）
mint -s ja -t en "konnichiwa"    # 日语罗马音 → hello
```

### 3. 智能语言检测

**自动检测并翻译：**

```bash
export MINT_TARGET_LANG=en

mint "早安"   # 检测到中文 → Good morning
```

**语法与拼写纠正** — 当输入语言与目标语言相同时，Mint 会纠正而非翻译：

```bash
export MINT_TARGET_LANG=en

mint "Good mooorning"          # 检测到英语 → Good morning
mint "She don't know nothing"  # 检测到英语 → She doesn't know anything
mint "i luv coding"            # 检测到英语 → I love coding
```

**语言轮换** — 依次翻译为列表中的下一个语言，循环进行：

```bash
# 两种语言
export MINT_TARGET_LANG=en,zh-TW
mint "Hello"   # en → zh-TW: 你好
mint "你好"    # zh-TW → en: Hello

# 三种语言
export MINT_TARGET_LANG=en,zh-TW,ja
mint "Hello"       # en → zh-TW: 你好
mint "你好"        # zh-TW → ja: こんにちは
mint "こんにちは"   # ja → en: Hello
```

---

## 🔑 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MINT_PROVIDER` | `google-genai` \| `openai` \| `anthropic` | — （必填） |
| `MINT_API_KEY` | API 密钥；使用默认端点时必填；设置 `MINT_BASE_URL` 时可选（由代理处理鉴权） | — |
| `MINT_BASE_URL` | 自定义 API base URL（仅填域名，路径由各提供商自动附加）；搭配 `openai` 可指向 Ollama（`http://localhost:11434`）、LM Studio（`http://localhost:1234`）或任何其他 OpenAI 兼容端点 | 提供商默认值 |
| `MINT_MODEL_NAME` | 使用的模型；设置 `MINT_BASE_URL` 时必填 | `gemini-3.1-flash-lite` / `gpt-4o-mini` / `claude-haiku-4-5` |
| `MINT_TARGET_LANG` | 目标语言，例如 `en` 或 `en,zh-TW,ja` | 系统区域设置，否则为 `en` |
| `MINT_VERBOSE` | 设为 `true` 可启用详细诊断输出（等同于 `--verbose`） | `false` |

---

## 🚩 CLI 参数

| 参数 | 缩写 | 说明 |
|------|------|------|
| `--target <lang>` | `-t` | 目标语言（BCP-47 标签，例如 `ja`、`zh-TW`、`fr`）。覆盖 `MINT_TARGET_LANG`。 |
| `--source <lang>` | `-s` | 来源语言（BCP-47 标签）；跳过自动检测，强制从此语言翻译。 |
| `--verbose` | `-v` | 将诊断信息与 token 用量输出到 stderr。也可通过 `MINT_VERBOSE=true` 启用。 |
| `--version` | | 显示版本并退出。 |

---

## 📅 路线图

- [x] 多 LLM 提供商支持（Google Gemini、OpenAI、Anthropic，或任何 OpenAI 兼容端点）
- [x] 通过 `MINT_TARGET_LANG` 实现智能语言检测与多语言轮换
- [x] 通过 `--target` / `-t` 参数明确指定目标语言
- [x] 通过 `--source` / `-s` 参数明确指定来源语言
- [x] 流式输出
- [x] GoReleaser 多平台二进制发布（Linux / macOS / Windows）
- [x] 批量翻译模式 — 长输入按段落边界切块、逐块翻译
- [ ] 术语表 / 自定义词典支持
- [ ] 输出格式选项（纯文本、JSON、Markdown）
- [ ] 翻译结果缓存

---

## 📄 许可证

Apache License 2.0 — 详见 [LICENSE](https://github.com/min0625/mint/blob/main/LICENSE) 文件。
