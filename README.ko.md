🌐 [다른 언어](https://github.com/min0625/mint/blob/main/LANGUAGES.md)

# 🌿 Mint

> Minimalist AI Translation CLI — 심플하게. 빠르게. 직관적으로.

[![GitHub Release](https://img.shields.io/github/v/release/min0625/mint?logo=github)](https://github.com/min0625/mint/releases)
[![PyPI](https://img.shields.io/pypi/v/mint-ai?logo=pypi&logoColor=white)](https://pypi.org/project/mint-ai/)
[![npm](https://img.shields.io/npm/v/mint-ai?logo=npm)](https://www.npmjs.com/package/mint-ai)
[![codecov](https://codecov.io/gh/min0625/mint/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/mint)

Mint는 단일 실행 파일로 동작하는 LLM 기반 번역 CLI입니다. 환경 변수 두 개만 설정하면 파일, 파이프 출력, 직접 입력한 텍스트 등 무엇이든 명령줄에서 번역할 수 있습니다. 언어 자동 감지, 문법 교정, 스트리밍 출력, 다국어 순환 기능을 기본으로 제공합니다.

```bash
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_key

mint -t ja "Good morning"         # おはようございます
echo "早安" | mint -t en          # Good morning
cat document.txt | mint -t fr     # 파일 전체 번역
```

---

## ✨ 왜 Mint인가?

- **제로 설정** — 단일 실행 파일; API 키는 환경 변수로 관리하여 설정 파일이 지저분해지지 않음
- **멀티 프로바이더** — Google Gemini, OpenAI, Anthropic은 물론, OpenAI 호환 엔드포인트(Ollama, LM Studio, OpenRouter, Groq, DeepSeek, llama.cpp 등)도 지원
- **스마트 감지** — 호출할 때마다 언어를 자동 감지; 숫자·기호 같은 언어 중립적 콘텐츠는 그대로 출력
- **스마트 교정** — 입력 언어와 목표 언어가 같다면 번역 대신 문법과 철자를 자동 교정
- **스트리밍** — 응답을 실시간으로 스트리밍하여 긴 번역도 기다릴 필요 없음
- **조합 가능** — stdin/stdout에 친화적인 설계로 `grep`, `sed`, `xargs` 등과 매끄럽게 연동
- **보안** — system/user 메시지 분리와 요청마다 생성되는 무작위 nonce 구분자를 통해 신뢰할 수 없는 입력을 모델 지시로부터 격리; 악의적인 콘텐츠를 번역해도 LLM의 동작을 탈취할 수 없음

---

## 📋 설치

### 자동 설치 (권장)

**macOS / Linux**

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh)"
```

OS와 아키텍처(Linux/macOS, x86_64/arm64)를 자동 감지하여 `~/.local/bin`에 설치합니다. `MINT_INSTALL_DIR`로 설치 경로를 변경하거나 `MINT_VERSION=v1.0.0`으로 버전을 지정할 수 있습니다.

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/min0625/mint/main/script/install.ps1 | iex
```

아키텍처(x86_64/arm64)를 자동 감지하여 `$HOME\.local\bin`에 설치합니다. `$env:MINT_INSTALL_DIR`로 설치 경로를 변경하거나 `$env:MINT_VERSION = 'v1.0.0'`으로 버전을 지정할 수 있습니다.

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

### 수동 다운로드

[GitHub Releases](https://github.com/min0625/mint/releases)에서 사용 중인 플랫폼용 사전 빌드 바이너리를 다운로드한 뒤 `PATH`에 포함된 디렉터리로 옮기고 다음 명령으로 확인합니다:

```bash
mint --version
```

---

## 🚀 빠른 시작

### 1. 프로바이더 설정

```bash
# Google Gemini (무료 등급 제공 — https://aistudio.google.com/apikey)
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_gemini_api_key

# OpenAI
export MINT_PROVIDER=openai
export MINT_API_KEY=sk-...

# Anthropic
export MINT_PROVIDER=anthropic
export MINT_API_KEY=sk-ant-...

# Ollama (API 키 불필요)
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=qwen2.5:7b  # Ollama에 로드된 원하는 모델로 변경

# LM Studio (API 키 불필요)
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:1234
export MINT_MODEL_NAME=lmstudio-community/Qwen2.5-7B-Instruct-GGUF  # LM Studio에 로드된 원하는 모델로 변경

# llama.cpp의 llama-server (API 키 불필요)
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:8080
export MINT_MODEL_NAME=qwen2.5:7b  # llama-server에 로드된 모델과 일치시키기

# OpenRouter (키 하나로 수백 개 모델 사용 — https://openrouter.ai/models)
export MINT_PROVIDER=openai
export MINT_BASE_URL=https://openrouter.ai/api
export MINT_API_KEY=sk-or-...
export MINT_MODEL_NAME=openai/gpt-4o-mini

# Groq (빠른 추론 속도, 무료 등급 제공)
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

### 2. 번역하기

```bash
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

echo "The quick brown fox" | mint -t fr
cat document.txt | mint -t zh-TW
```

`--verbose` / `-v` (또는 `MINT_VERBOSE=true`)를 사용하면 진단 정보와 토큰 사용량이 stderr로 출력됩니다:

```bash
mint -t ja -v "Good morning"
# [mint] provider: google-genai
# [mint] single target — skipping language detection
# [mint] target language: ja
# おはようございます
# [mint] tokens: 113 in / 2 out
```

**대표적인 토큰 사용량** (`gemini-3.1-flash-lite` 기준 실측):

| 모드 | 입력 | 호출 횟수 | 입력 토큰 | 출력 토큰 |
|------|------|----------|----------|----------|
| 단일 목표 (`-t` 또는 단일 `MINT_TARGET_LANG`) | 짧은 단어/문장 | 1 | ~110–130 | ~1–15 |
| 단일 목표 | 긴 글 (`testdata/sample.txt`) | 1 | ~465–470 | ~450–560 |
| 다중 목표 순환 (쉼표로 구분된 `MINT_TARGET_LANG`) | 짧은 문장 | 2 | ~250–260 | ~2–8 |
| 명시적 소스 `-s` + 순환 | 짧은 문장 | 1 | ~105–120 | ~1–2 |

> 토큰 수는 입력 길이에 비례해 증가합니다. 출력 토큰은 목표 언어에 따라 달라지며, 일본어와 중국어는 동일한 내용이라도 영어보다 더 많은 토큰을 생성하는 경향이 있습니다.

**100만 토큰으로 몇 번이나 번역할 수 있을까?** (입력+출력 합산, 위 실측값에서 산출):

| 입력 | 번역당 약 토큰 수 | 100만 토큰당 번역 횟수 |
|------|-----------------|---------------------|
| 짧은 단어나 구문 | 약 120 | 약 8,000회 |
| 300단어 분량 글 | 약 1,000 | 약 1,000회 |

> 수치는 입력과 출력 토큰을 합산한 것입니다. 프로바이더마다 입력과 출력 요금이 다르며 무료 등급을 제공하는 곳도 많으니, 정확한 요금은 각 프로바이더의 요금 페이지를 확인하세요. Google Gemini는 [Google AI Studio](https://aistudio.google.com/apikey)의 무료 등급을 신용카드 없이 이용할 수 있습니다.

`--source` / `-s`로 **소스 언어를 강제 지정**하면 목표 언어에서도 유효한 입력(언어 간 동형이의어, 로마자 표기)을 번역할 수 있습니다:

```bash
mint -s fr -t en "pain"          # 프랑스어 → bread (-s 없으면 영어 "pain"으로 처리됨)
mint -s ja -t en "konnichiwa"    # 로마자 일본어 → hello
```

### 3. 스마트 언어 감지

**자동 감지 후 번역:**

```bash
export MINT_TARGET_LANG=en

mint "早安"   # 중국어로 감지 → Good morning
```

**문법 및 철자 교정** — 입력 언어와 목표 언어가 같으면 번역 대신 교정합니다:

```bash
export MINT_TARGET_LANG=en

mint "Good mooorning"          # 영어로 감지 → Good morning
mint "She don't know nothing"  # 영어로 감지 → She doesn't know anything
mint "i luv coding"            # 영어로 감지 → I love coding
```

**언어 순환** — 목록의 다음 언어로 번역하며, 마지막에 도달하면 처음으로 돌아갑니다:

```bash
# 언어 2개
export MINT_TARGET_LANG=en,zh-TW
mint "Hello"   # en → zh-TW: 你好
mint "你好"    # zh-TW → en: Hello

# 언어 3개
export MINT_TARGET_LANG=en,zh-TW,ja
mint "Hello"       # en → zh-TW: 你好
mint "你好"        # zh-TW → ja: こんにちは
mint "こんにちは"   # ja → en: Hello
```

---

## 🔑 환경 변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `MINT_PROVIDER` | `google-genai` \| `openai` \| `anthropic` | — (필수) |
| `MINT_API_KEY` | API 키; 기본 엔드포인트 사용 시 필수, `MINT_BASE_URL` 설정 시 선택 (프록시가 인증을 처리하는 경우) | — |
| `MINT_BASE_URL` | 사용자 지정 API base URL (도메인만 입력, 경로는 각 프로바이더가 자동으로 붙임); `openai`와 함께 사용해 Ollama (`http://localhost:11434`), LM Studio (`http://localhost:1234`) 또는 그 밖의 OpenAI 호환 엔드포인트를 지정 가능 | 프로바이더 기본값 |
| `MINT_MODEL_NAME` | 사용할 모델; `MINT_BASE_URL` 설정 시 필수 | `gemini-3.1-flash-lite` / `gpt-4o-mini` / `claude-haiku-4-5` |
| `MINT_TARGET_LANG` | 목표 언어, 예: `en` 또는 `en,zh-TW,ja` | 시스템 로케일, 없으면 `en` |
| `MINT_VERBOSE` | `true`로 설정하면 상세 진단 출력 활성화 (`--verbose`와 동일) | `false` |

---

## 🚩 CLI 플래그

| 플래그 | 축약형 | 설명 |
|------|------|------|
| `--target <lang>` | `-t` | 목표 언어 (BCP-47 태그, 예: `ja`, `zh-TW`, `fr`). `MINT_TARGET_LANG`을 덮어씀. |
| `--source <lang>` | `-s` | 소스 언어 (BCP-47 태그); 자동 감지를 건너뛰고 이 언어로부터 번역하도록 강제함. |
| `--verbose` | `-v` | 진단 정보와 토큰 사용량을 stderr로 출력. `MINT_VERBOSE=true`로도 활성화 가능. |
| `--version` | | 버전을 출력하고 종료. |

---

## 📅 로드맵

- [x] 다중 LLM 프로바이더 지원 (Google Gemini, OpenAI, Anthropic 및 OpenAI 호환 엔드포인트)
- [x] `MINT_TARGET_LANG`을 통한 스마트 언어 감지 및 다국어 순환
- [x] `--target` / `-t` 플래그를 통한 명시적 목표 언어 지정
- [x] `--source` / `-s` 플래그를 통한 명시적 소스 언어 지정
- [x] 스트리밍 출력
- [x] GoReleaser 기반 멀티 플랫폼 바이너리 릴리스 (Linux / macOS / Windows)
- [ ] 배치 번역 모드
- [ ] 용어집 / 사용자 지정 사전 지원
- [ ] 출력 형식 옵션 (일반 텍스트, JSON, Markdown)
- [ ] 반복 번역 캐싱

---

## 📄 라이선스

Apache License 2.0 — 자세한 내용은 [LICENSE](https://github.com/min0625/mint/blob/main/LICENSE) 파일을 참고하세요.
