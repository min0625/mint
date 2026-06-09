# 人工測試案例

### 1. 單字翻譯（透過 -t flag 指定目標語言）

```sh
mint -t zh-TW "apple"        # 蘋果
mint -t zh-TW "りんご"        # 蘋果
mint -t en    "蘋果"          # apple
mint -t en    "りんご"        # apple
mint -t ja    "apple"         # りんご
mint -t ja    "蘋果"          # りんご
```

### 2. 句子翻譯（透過 -t flag 指定目標語言）

```sh
mint -t zh-TW "This is an apple."     # 這是一顆蘋果。
mint -t zh-TW "これはリンゴです。"    # 這是一顆蘋果。
mint -t en    "這是一顆蘋果。"        # This is an apple.
mint -t en    "これはリンゴです。"    # This is an apple.
mint -t ja    "This is an apple."     # これはリンゴです。
mint -t ja    "這是一顆蘋果。"        # これはリンゴです。
```

### 3. 透過環境變數指定單一目標語言

```sh
# 目標語言 zh-TW
export MINT_TARGET_LANG=zh-TW
mint "apple"                  # 蘋果
mint "This is an apple."      # 這是一顆蘋果。
mint "これはリンゴです。"      # 這是一顆蘋果。

# 目標語言 en
export MINT_TARGET_LANG=en
mint "蘋果"                    # apple
mint "這是一顆蘋果。"          # This is an apple.
mint "これはリンゴです。"      # This is an apple.

# 目標語言 ja
export MINT_TARGET_LANG=ja
mint "apple"                  # りんご
mint "This is an apple."      # これはリンゴです。
mint "這是一顆蘋果。"          # これはリンゴです。
```

### 4. 同語言 → 拼字／文法校正

```sh
export MINT_TARGET_LANG=zh-TW
mint "這事一科蘋果"            # 這是一顆蘋果（校正）

export MINT_TARGET_LANG=en
mint "This are an apple."     # This is an apple.（校正）

export MINT_TARGET_LANG=ja
mint "これわリンゴです。"       # これはリンゴです。（校正）
```

### 5. 同語系視為相同語言 → 校正而非翻譯

```sh
# zh-HK 與 zh-TW 主語言相同，視為同語言
export MINT_TARGET_LANG=zh-TW
mint "這係一個蘋果"            # 這是一顆蘋果（校正／標準化，非翻譯）
```

### 6. 環境變數語言輪換（兩個語言）

```sh
export MINT_TARGET_LANG=zh-TW,en

mint "This is an apple."      # 這是一顆蘋果。（en 輸入 → 下一個 zh-TW）
mint "これはリンゴです。"      # 這是一顆蘋果。（ja 不在列表中 → 預設 zh-TW）
mint "這是一顆蘋果。"          # This is an apple.（zh-TW 輸入 → 下一個 en）
```

### 7. 環境變數語言輪換（三個語言）

```sh
export MINT_TARGET_LANG=zh-TW,en,ja

mint "This is an apple."      # 這是一顆蘋果。（en 輸入 → 下一個 zh-TW）
mint "這是一顆蘋果。"          # This is an apple.（zh-TW 輸入 → 下一個 en）
mint "これはリンゴです。"      # 這是一顆蘋果。（ja 輸入 → 下一個 zh-TW）
```

### 8. -t flag 優先於環境變數

```sh
export MINT_TARGET_LANG=ja
mint -t zh-TW "This is an apple."   # 這是一顆蘋果。（flag 覆蓋 env）
```
