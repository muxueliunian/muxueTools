# 可翻译文本提取 (i18n Text Extraction)

**目的**: 本文档列出所有前端页面的硬编码文本，供翻译模型参考创建完整的翻译文件。

**翻译文件位置**:
- `web/src/i18n/locales/zh-CN.json` - 简体中文
- `web/src/i18n/locales/en-US.json` - 英语
- `web/src/i18n/locales/ja-JP.json` - 日语

---

## 📋 翻译文件结构

```json
{
  "common": { /* 通用词汇 */ },
  "sidebar": { /* 侧边栏 */ },
  "chat": { /* 对话页面 */ },
  "dashboard": { /* 仪表盘 */ },
  "keys": { /* API Keys 管理 */ },
  "stats": { /* 统计页面 */ },
  "settings": { /* 设置页面 */ }
}
```

---

## 🔤 需要翻译的文本

### common (通用词汇)

| Key | 中文 | English | 日本語 |
|-----|------|---------|--------|
| save | 保存 | Save | 保存 |
| cancel | 取消 | Cancel | キャンセル |
| confirm | 确认 | Confirm | 確認 |
| delete | 删除 | Delete | 削除 |
| loading | 加载中... | Loading... | 読み込み中... |
| success | 成功 | Success | 成功 |
| error | 错误 | Error | エラー |
| copy | 复制 | Copy | コピー |
| retry | 重试 | Retry | 再試行 |
| back | 返回 | Back | 戻る |
| next | 下一步 | Next | 次へ |
| create | 创建 | Create | 作成 |
| import | 导入 | Import | インポート |
| export | 导出 | Export | エクスポート |
| search | 搜索 | Search | 検索 |
| close | 关闭 | Close | 閉じる |
| active | 活跃 | Active | アクティブ |
| disabled | 已禁用 | Disabled | 無効 |
| neverUsed | 从未使用 | Never used | 未使用 |
| copiedToClipboard | 已复制到剪贴板 | Copied to clipboard | クリップボードにコピーしました |

---

### sidebar (侧边栏)

| Key | 中文 | English | 日本語 |
|-----|------|---------|--------|
| chat | 对话 | Chat | チャット |
| dashboard | 仪表盘 | Dashboard | ダッシュボード |
| keys | API Keys | API Keys | API Keys |
| stats | 统计 | Statistics | 統計 |
| settings | 设置 | Settings | 設定 |
| newChat | 新对话 | New Chat | 新規チャット |
| darkMode | 深色模式 | Dark Mode | ダークモード |
| lightMode | 浅色模式 | Light Mode | ライトモード |

---

### chat (对话页面)

| Key | 中文 | English | 日本語 |
|-----|------|---------|--------|
| placeholder | 有什么可以帮到你的？ | How can I help you? | 何かお手伝いできることはありますか？ |
| sendHint | 按 Enter 发送，Shift + Enter 换行 | Press Enter to send, Shift + Enter for new line | Enterで送信、Shift + Enterで改行 |
| stopGeneration | 已停止生成 | Generation stopped | 生成を停止しました |
| welcomeBack | Back at it, {name} | Back at it, {name} | おかえりなさい、{name} |
| video | 视频 | Video | 動画 |
| deleteSession | 删除会话 | Delete session | セッションを削除 |
| loadingSessions | 加载中... | Loading... | 読み込み中... |
| noSessions | 暂无会话 | No sessions | セッションがありません |
| deleteChat | 删除对话 | Delete Chat | チャットを削除 |
| deleteChatConfirm | 确定要删除这个对话吗？此操作无法撤销。 | Are you sure you want to delete this chat? This action cannot be undone. | このチャットを削除してもよろしいですか？この操作は元に戻せません。 |
| justNow | 刚刚 | Just now | たった今 |
| minutesAgo | {n} 分钟前 | {n} minutes ago | {n}分前 |
| hoursAgo | {n} 小时前 | {n} hours ago | {n}時間前 |
| daysAgo | {n} 天前 | {n} days ago | {n}日前 |
| newChatTitle | New Chat | New Chat | 新規チャット |

---

### dashboard (仪表盘)

| Key | 中文 | English | 日本語 |
|-----|------|---------|--------|
| title | 仪表盘 | Dashboard | ダッシュボード |
| subtitle | MuxueTools 代理服务 - OpenAI 兼容网关 | MuxueTools Proxy Service - OpenAI Compatible Gateway | MuxueTools プロキシサービス - OpenAI 互換ゲートウェイ |
| apiEndpoint | API 端点 | API Endpoint | APIエンドポイント |
| baseUrl | 基础 URL | Base URL | ベースURL |
| apiKey | API Key | API Key | APIキー |
| running | 运行中 | Running | 実行中 |
| degraded | 降级 | Degraded | 低下 |
| keysActive | {active} / {total} 个密钥活跃 | {active} of {total} keys active | {active} / {total} キーがアクティブ |
| uptime | 运行时间 | Uptime | 稼働時間 |
| quickStart | 快速开始 | Quick Start | クイックスタート |
| tip | 💡 提示 | 💡 Tip | 💡 ヒント |
| noApiKeyNeeded | 无需 API Key！本地反代已配置密钥池，可直接使用。 | No API Key needed! The local proxy has a key pool configured. | APIキー不要！ローカルプロキシにはキープールが設定されています。 |
| totalKeys | 总密钥数 | Total Keys | 総キー数 |
| activeKeys | 活跃 | Active | アクティブ |
| rateLimited | 限速中 | Rate Limited | レート制限中 |
| disabledKeys | 已禁用 | Disabled | 無効 |
| connectionError | 连接错误 | Connection Error | 接続エラー |
| pythonComment | # 本地反代无需 Key | # No key needed for local proxy | # ローカルプロキシはキー不要 |

---

### keys (API Keys 管理)

| Key | 中文 | English | 日本語 |
|-----|------|---------|--------|
| title | API Keys | API Keys | APIキー |
| subtitle | 管理 AI 模型的认证密钥 | Manage authentication keys for your AI models. | AIモデルの認証キーを管理します。 |
| searchPlaceholder | 搜索密钥... | Search keys... | キーを検索... |
| createKey | 创建密钥 | Create Key | キーを作成 |
| importKeys | 导入密钥 | Import Keys | キーをインポート |
| addNewKey | 添加新密钥 | Add New API Key | 新しいAPIキーを追加 |
| status | 状态 | STATUS | ステータス |
| name | 名称 | NAME | 名前 |
| key | 密钥 | KEY | キー |
| tags | 标签 | TAGS | タグ |
| usage24h | 用量 (24小时) | USAGE (24H) | 使用量 (24時間) |
| actions | 操作 | ACTIONS | 操作 |
| untitledKey | 未命名密钥 | Untitled Key | 無題のキー |
| requests | {n} 请求 | {n} reqs | {n} リクエスト |
| revokeKey | 撤销密钥 | Revoke Key | キーを取り消す |
| revokeConfirm | 确定要撤销此 API 密钥吗？此操作无法撤销。 | Are you sure you want to revoke this API key? This action cannot be undone. | このAPIキーを取り消してもよろしいですか？この操作は元に戻せません。 |
| revoke | 撤销 | Revoke | 取り消す |
| keyRevokedSuccess | 密钥已成功撤销 | Key revoked successfully | キーが正常に取り消されました |
| keyCreatedSuccess | 密钥已成功创建 | Key created successfully | キーが正常に作成されました |
| testingKey | 正在测试密钥连接... | Testing key connection... | キー接続をテスト中... |
| connectionSuccess | 连接成功 ({latency}ms) | Connection successful ({latency}ms) | 接続成功 ({latency}ms) |
| connectionFailed | 连接失败或密钥无效 | Connection failed or key invalid | 接続失敗またはキーが無効です |
| wizardStep1 | 提供商和密钥 | Provider & Key | プロバイダーとキー |
| wizardStep2 | 选择模型 | Select Model | モデルを選択 |
| wizardStep3 | 详情 | Details | 詳細 |
| wizardStep4 | 确认 | Confirm | 確認 |
| provider | 提供商 | Provider | プロバイダー |
| selectProvider | 选择提供商 | Select Provider | プロバイダーを選択 |
| googleAistudio | Google AI Studio | Google AI Studio | Google AI Studio |
| geminiApi | Gemini API | Gemini API | Gemini API |
| enterApiKey | 输入您的 API Key（例如 AIzaSy...） | Enter your API Key (e.g., AIzaSy...) | APIキーを入力してください（例: AIzaSy...） |
| validateAndFetch | 验证并获取模型 | Validate & Fetch Models | 検証してモデルを取得 |
| validating | 验证中... | Validating... | 検証中... |
| keyValidatedSuccess | 密钥验证成功！延迟: {latency}ms | Key validated successfully! Latency: {latency}ms | キー検証成功！レイテンシ: {latency}ms |
| foundModels | 发现 {count} 个可用模型。您可以跳过模型选择。 | Found {count} available models. You can skip model selection if preferred. | {count}個の利用可能なモデルが見つかりました。モデル選択をスキップできます。 |
| defaultModel | 默认模型 | Default Model | デフォルトモデル |
| selectDefaultModel | 选择默认模型 | Select a default model | デフォルトモデルを選択 |
| keyNameOptional | 密钥名称（可选） | Key Name (Optional) | キー名（オプション） |
| keyNamePlaceholder | 例如 Production Key, Dev Team Key | e.g., Production Key, Dev Team Key | 例: Production Key, Dev Team Key |
| tagsOptional | 标签（可选） | Tags (Optional) | タグ（オプション） |
| tagsPlaceholder | production, high-priority（逗号分隔） | production, high-priority (comma separated) | production, high-priority（カンマ区切り） |
| notSelected | 未选择 | Not selected | 未選択 |
| untitled | 未命名 | Untitled | 無題 |
| none | 无 | None | なし |
| importDescription | 逐行粘贴密钥，或提供 JSON 数组。 | Paste keys line by line, or provide a JSON array. | 行ごとにキーを貼り付けるか、JSON配列を入力してください。 |
| importedSuccess | 已导入 {imported} 个密钥（跳过 {skipped} 个） | Imported {imported} keys ({skipped} skipped) | {imported}個のキーをインポート（{skipped}個スキップ） |
| importFailed | 导入失败 | Import failed | インポートに失敗しました |
| noValidKeys | 未找到有效的密钥 | No valid keys found to import | 有効なキーが見つかりませんでした |
| someKeysFailed | 部分密钥失败: {count} 个错误 | Some keys failed: {count} errors | 一部のキーが失敗しました: {count}個のエラー |

---

### stats (统计页面)

| Key | 中文 | English | 日本語 |
|-----|------|---------|--------|
| title | 统计 | Statistics | 統計 |
| subtitle | 监控 API 使用、趋势和模型分布 | Monitor API usage, trends, and model distribution. | API使用量、トレンド、モデル分布を監視します。 |
| last24Hours | 过去 24 小时 | Last 24 Hours | 過去24時間 |
| last7Days | 过去 7 天 | Last 7 Days | 過去7日間 |
| last30Days | 过去 30 天 | Last 30 Days | 過去30日間 |
| totalRequests | 总请求数 | Total Requests | 総リクエスト数 |
| totalTokens | 总 Token 数 | Total Tokens | 総トークン数 |
| errorRate | 错误率 | Error Rate | エラー率 |
| requestTrend | 请求趋势 | Request Trend | リクエストトレンド |
| modelDistribution | 模型分布 | Model Distribution | モデル分布 |
| noModelUsageData | 暂无模型使用数据 | No model usage data | モデル使用データがありません |
| requests | 请求 | Requests | リクエスト |
| errors | 错误 | Errors | エラー |

---

### settings (设置页面)

| Key | 中文 | English | 日本語 |
|-----|------|---------|--------|
| title | 设置 | Settings | 設定 |
| subtitle | 配置系统行为和性能 | Configure system behavior and performance. | システムの動作とパフォーマンスを設定します。 |
| general | 通用 | General | 一般 |
| security | 安全 | Security | セキュリティ |
| advanced | 高级 | Advanced | 詳細 |
| model | 模型 | Model | モデル |
| keyManagement | 密钥管理 | Key Management | キー管理 |
| selectionStrategy | 选择策略 | Selection Strategy | 選択戦略 |
| strategyDescription | 用于选择下一个可用 API 密钥的算法 | Algorithm used to select the next available API key. | 次に利用可能なAPIキーを選択するためのアルゴリズム。 |
| roundRobin | 轮询（顺序） | Round Robin (Sequential) | ラウンドロビン（順次） |
| randomSelection | 随机选择 | Random Selection | ランダム選択 |
| leastUsedFirst | 最少使用优先 | Least Used First | 使用頻度が低い順 |
| weightedRandom | 加权随机 | Weighted Random | 重み付けランダム |
| loggingAndUpdates | 日志和更新 | Logging & Updates | ログと更新 |
| logLevel | 日志级别 | Log Level | ログレベル |
| debugVerbose | 调试（详细） | Debug (Verbose) | デバッグ（詳細） |
| infoStandard | 信息（标准） | Info (Standard) | 情報（標準） |
| warning | 警告 | Warning | 警告 |
| errorCritical | 错误（仅关键） | Error (Critical only) | エラー（重大のみ） |
| automaticUpdates | 自动更新 | Automatic Updates | 自動更新 |
| checkOnStartup | 启动时检查新版本 | Check for new versions on startup. | 起動時に新しいバージョンを確認します。 |
| checkNow | 立即检查 | Check Now | 今すぐ確認 |
| updateSource | 更新源 | Update Source | 更新ソース |
| mxlnServer | mxln 服务器（推荐） | mxln Server (Recommended) | mxlnサーバー（推奨） |
| github | GitHub | GitHub | GitHub |
| mxlnDescription | 使用 mxln 服务器检查更新（中国用户推荐） | Use mxln server for updates (recommended for China) | mxlnサーバーで更新を確認（中国のユーザーに推奨） |
| githubDescription | 使用 GitHub Releases 检查更新 | Use GitHub Releases for updates | GitHub Releasesで更新を確認 |
| updateAvailable | 有新版本可用: v{version} | Update Available: v{version} | 新しいバージョンが利用可能: v{version} |
| latestVersion | 您正在使用最新版本 | You are using the latest version | 最新バージョンを使用しています |
| downloadUpdate | 下载更新 → | Download Update → | 更新をダウンロード → |
| accessControl | 访问控制 | Access Control | アクセス制御 |
| ipWhitelist | IP 白名单 | IP Whitelist | IPホワイトリスト |
| ipWhitelistDescription | 仅允许来自特定 IP 地址的请求 | Only allow requests from specific IP address. | 特定のIPアドレスからのリクエストのみを許可します。 |
| allowedIpAddress | 允许的 IP 地址 | Allowed IP Address | 許可されたIPアドレス |
| localhostAlwaysAllowed | 本地主机 (127.0.0.1) 始终允许，以防止锁定 | Localhost (127.0.0.1) is always allowed to prevent lockout. | ローカルホスト（127.0.0.1）は常に許可され、ロックアウトを防止します。 |
| proxyApiKey | 代理 API Key | Proxy API Key | プロキシAPIキー |
| proxyKeyDescription | 用于验证发送到此代理的请求。仅与授权用户共享。 | Used to authenticate requests to this proxy. Share with authorized users only. | このプロキシへのリクエストを認証するために使用されます。認可されたユーザーのみと共有してください。 |
| regenerate | 重新生成 | Regenerate | 再生成 |
| keyRegenerated | 代理密钥已成功重新生成 | Proxy key regenerated successfully | プロキシキーが正常に再生成されました |
| performanceTuning | 性能调优 | Performance Tuning | パフォーマンスチューニング |
| cooldownTime | 冷却时间 | Cooldown Time | クールダウン時間 |
| cooldownDescription | 限速冷却 | Rate limit cooldown | レート制限クールダウン |
| maxRetries | 最大重试次数 | Max Retries | 最大リトライ回数 |
| retryOnFailure | 失败时重试 | Retry on failure | 失敗時にリトライ |
| requestTimeout | 请求超时 | Request Timeout | リクエストタイムアウト |
| apiRequestTimeout | API 请求超时 | API request timeout | APIリクエストタイムアウト |
| debugMode | 调试模式 | Debug Mode | デバッグモード |
| verboseLogging | 启用详细日志输出 | Enable verbose logging output. | 詳細なログ出力を有効にします。 |
| dataManagement | 数据管理 | Data Management | データ管理 |
| databaseLocation | 数据库位置 | Database Location | データベースの場所 |
| databaseLocationDescription | SQLite 数据库文件路径（只读） | SQLite database file path (read-only) | SQLiteデータベースファイルパス（読み取り専用） |
| dangerZone | 危险区域 | Danger Zone | 危険ゾーン |
| deleteChatHistory | 删除聊天记录 | Delete Chat History | チャット履歴を削除 |
| deleteChatDescription | 删除所有聊天会话和消息 | Remove all chat sessions and messages. | すべてのチャットセッションとメッセージを削除します。 |
| resetStatistics | 重置统计 | Reset Statistics | 統計をリセット |
| resetStatsDescription | 清除所有 API 密钥的使用统计 | Clear all API key usage statistics. | すべてのAPIキー使用統計をクリアします。 |
| deleteAllChatHistory | 删除所有聊天记录 | Delete All Chat History | すべてのチャット履歴を削除 |
| deleteAllChatsConfirm | 此操作将永久删除所有聊天会话和消息，无法撤销。 | This action will permanently delete all chat sessions and messages. This cannot be undone. | この操作はすべてのチャットセッションとメッセージを完全に削除します。元に戻すことはできません。 |
| deleteAll | 全部删除 | Delete All | すべて削除 |
| deletedSessions | 已成功删除 {count} 个会话 | Deleted {count} sessions successfully | {count}個のセッションを正常に削除しました |
| resetAllStatistics | 重置所有统计 | Reset All Statistics | すべての統計をリセット |
| resetAllStatsConfirm | 此操作将重置所有 API 密钥的使用统计（请求次数、Token 用量等），无法撤销。 | This action will reset all API key usage statistics (request counts, token usage, etc.). This cannot be undone. | この操作はすべてのAPIキー使用統計（リクエスト数、トークン使用量など）をリセットします。元に戻すことはできません。 |
| resetAll | 全部重置 | Reset All | すべてリセット |
| resetKeysAffected | 已重置 {count} 个密钥的统计 | Reset statistics for {count} keys | {count}個のキーの統計をリセットしました |
| saveChanges | 保存更改 | Save Changes | 変更を保存 |
| configSavedSuccess | 配置已成功保存 | Configuration saved successfully | 設定が正常に保存されました |
| systemPrompt | 系统提示词 | System Prompt | システムプロンプト |
| defaultSystemPrompt | 默认系统提示词 | Default System Prompt | デフォルトシステムプロンプト |
| systemPromptPlaceholder | 输入用于所有请求的系统提示词... | Enter a system prompt to be used for all requests... | すべてのリクエストに使用するシステムプロンプトを入力... |
| systemPromptDescription | 此提示词将添加到所有聊天请求的前面 | This prompt will be prepended to all chat requests. | このプロンプトはすべてのチャットリクエストの先頭に追加されます。 |
| generationParameters | 生成参数 | Generation Parameters | 生成パラメータ |
| temperature | 温度 | Temperature | 温度 |
| temperatureDescription | 控制随机性。较低 = 更确定性，较高 = 更有创意 | Controls randomness. Lower = more deterministic, Higher = more creative. | ランダム性を制御します。低い = より決定的、高い = より創造的。 |
| topP | Top-P | Top-P | Top-P |
| nucleusSampling | 核采样阈值 | Nucleus sampling threshold | 核サンプリング閾値 |
| topK | Top-K | Top-K | Top-K |
| topKSampling | Top-K 采样 | Top-K sampling | Top-Kサンプリング |
| maxOutputTokens | 最大输出 Token 数 | Max Output Tokens | 最大出力トークン数 |
| maxOutputDescription | 生成的最大 Token 数 | Maximum number of tokens to generate. | 生成する最大トークン数。 |
| advancedFeatures | 高级功能 (Gemini 2.5+) | Advanced Features (Gemini 2.5+) | 高度な機能（Gemini 2.5+） |
| thinkingLevel | 思考级别 | Thinking Level | 思考レベル |
| thinkingLevelDescription | 控制支持模型的推理深度 | Controls reasoning depth for supported models. | サポートされているモデルの推論深度を制御します。 |
| thinkingDisabled | 禁用 | Disabled | 無効 |
| thinkingLow | 低 | Low | 低 |
| thinkingMedium | 中 | Medium | 中 |
| thinkingHigh | 高 | High | 高 |
| mediaResolution | 媒体分辨率 | Media Resolution | メディア解像度 |
| mediaResolutionDescription | 图像/视频处理分辨率 | Image/video processing resolution. | 画像/動画処理解像度。 |
| mediaDefault | 默认 | Default | デフォルト |
| mediaLow | 低 (64 tokens) | Low (64 tokens) | 低（64トークン） |
| mediaMedium | 中 (256 tokens) | Medium (256 tokens) | 中（256トークン） |
| mediaHigh | 高 (缩放) | High (scaling) | 高（スケーリング） |
| sec | 秒 | sec | 秒 |

---

## 📝 使用说明

1. 翻译模型应按上述表格填充完整的翻译文件
2. 保持 JSON key 不变，只翻译 value
3. 带参数的文本（如 `{name}`, `{count}`）保持参数名不变
4. 专有名词（如 API Keys, Token, Gemini）可保持原样

---

*生成时间: 2026-01-20*
