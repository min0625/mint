# Mint 專案重新設計 - 完成紀錄

## 實施的變更

### 1. Config 結構更新 (internal/provider/config.go)
- 移除 `PrimaryLanguage` 和 `SecondaryLanguage` 字段
- 添加 `TargetLang` 字段支持逗號分隔的語言列表

### 2. 主命令邏輯重構 (cmd/mint/main.go)
- 實現三優先級語言解析：
  1. CLI 標誌 (`--target` / `-t`)
  2. 環境變數 (`MINT_TARGET_LANG`)
  3. 系統區域設定
  4. 預設為 `en`

### 3. 新增輔助函數
- `resolveTargetLangs()` - 優先級解析
- `getSystemLanguage()` - 系統區域設定讀取
- `detectLanguage()` - LLM 驅動的語言檢測
- `determineActualTargetLang()` - 多語言輪換邏輯

### 4. 文檔更新
- README.md - 更新快速開始、環境變數文檔
- README.zh-TW.md - 更新繁體中文文檔
- AGENTS.md - 更新環境變數表格

## 核心功能實現

✅ 優先級解析系統
✅ 語言檢測機制
✅ 多語言輪換邏輯
✅ 語法修正與翻譯邏輯分離
✅ 命令行介面更新 (-t/--target 標誌)
✅ 編譯成功，無錯誤

## 待驗證項目
- 實際翻譯功能需要有效的 LLM API 金鑰測試
- 語言檢測準確性
- 多語言輪換邏輯