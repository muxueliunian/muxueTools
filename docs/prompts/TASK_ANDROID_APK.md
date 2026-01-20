# 任务：Android APK 适配 (Android 12+)

## 角色
Developer (senior-golang + gomobile + kotlin-android)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md` - Go 代码规范
- `.agent/skills/gomobile/SKILL.md` - gomobile 编译与 API 设计
- `.agent/skills/kotlin-android/SKILL.md` - Kotlin Android 开发

## 背景

MuxueTools 当前支持 Windows/Linux/macOS 桌面平台。为了扩大用户群体，需要开发 Android 版本。

**技术方案：**
使用 `gomobile` 将 Go 后端编译为 Android 库 (`.aar`)，在 Android Studio 中构建原生外壳，使用 WebView 展示前端界面。

**已完成的依赖模块：**（参见 `docs/DEVELOPMENT.md`）
- Go 后端完整功能
- Vue3 前端完整功能
- 桌面应用 WebView 封装经验

**目标平台：**
- 最低版本: Android 12 (API 31)
- 目标版本: Android 15 (API 35)
- 架构: arm64-v8a, armeabi-v7a

## 目标

| Phase | 目标 | 难度 |
|-------|------|------|
| **Phase 1** | Go 后端 Android 适配 | ⭐⭐ 中 |
| **Phase 2** | Android 原生项目搭建 | ⭐⭐ 中 |
| **Phase 3** | WebView 集成与调试 | ⭐ 低 |
| **Phase 4** | 打包与测试 | ⭐ 低 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范（必须全部阅读）**
   - `.agent/skills/senior-golang/SKILL.md` - Go 代码规范与模式
   - `.agent/skills/gomobile/SKILL.md` - **gomobile 环境配置、API 设计、编译命令**
   - `.agent/skills/kotlin-android/SKILL.md` - **Kotlin Android 开发、WebView、前台服务**

2. **项目文档**
   - `docs/ARCHITECTURE.md` - 系统架构
   - `docs/DEVELOPMENT.md` - 开发工作流
   - `docs/IMPLEMENTATION_PLAN.md` - 阶段六 Android 章节

3. **相关代码**
   - `cmd/server/main.go` - 当前入口
   - `cmd/desktop/main.go` - 桌面应用入口（参考）
   - `embed.go` - 前端资源嵌入

---

### Phase 1: Go 后端 Android 适配

#### 1.1 环境准备

```bash
# 安装 gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest

# 初始化 gomobile (需要 Android SDK/NDK)
gomobile init
```

**环境要求：**
- Go 1.22+
- Android SDK (API 31+)
- Android NDK r25+ (通过 Android Studio 安装)
- 设置环境变量: `ANDROID_HOME`, `ANDROID_NDK_HOME`

#### 1.2 创建 Mobile 入口包

创建 `mobile/` 目录，封装供 Android 调用的接口：

```go
// mobile/mobile.go
package mobile

// StartServer 启动 HTTP 服务器
// bindAddr: 绑定地址，如 "127.0.0.1:8080"
// dataDir: 数据目录路径 (Android 内部存储)
// 返回: 错误信息，空字符串表示成功
func StartServer(bindAddr, dataDir string) string

// StopServer 停止服务器
func StopServer()

// IsRunning 检查服务器状态
func IsRunning() bool

// GetVersion 获取版本号
func GetVersion() string
```

#### 1.3 适配文件路径

修改配置文件和数据库路径逻辑，支持 Android 内部存储：

```go
// 桌面系统: 使用程序目录
// Android: 使用传入的 dataDir 参数
func GetDataDir() string {
    if runtime.GOOS == "android" {
        return androidDataDir // 由 StartServer 传入
    }
    return filepath.Dir(os.Args[0])
}
```

#### 1.4 编译 AAR

```bash
# 编译 Android 库
gomobile bind -target=android/arm64,android/arm -androidapi 31 -o muxuetools.aar ./mobile

# 输出文件:
# - muxuetools.aar (Android Archive)
# - muxuetools-sources.jar (源码)
```

---

### Phase 2: Android 原生项目搭建

> ⚠️ **重要**: 不创建独立项目，而是在**现有项目根目录**下添加 `android/` 子目录，保持单一仓库结构。

#### 2.1 在现有项目中创建 Android 模块

使用 Android Studio：
1. 打开 Android Studio → **File → New → New Project**
2. 选择 **Empty Activity**
3. 配置：
   - Name: `MuxueTools`
   - Package name: `com.muxue.tools`
   - Save location: **`<项目根目录>/android`** (重要！)
   - Language: Kotlin
   - Minimum SDK: API 31 (Android 12)
4. 创建完成后，`android/` 目录会出现在项目根目录

**最终项目结构：**
```
MuxueTools/                    # 现有项目根目录
├── android/                   # 新增: Android 原生项目
│   ├── app/
│   │   ├── src/main/
│   │   │   ├── java/com/muxue/tools/
│   │   │   │   ├── MainActivity.kt
│   │   │   │   ├── ServerService.kt
│   │   │   │   └── MuxueApplication.kt
│   │   │   ├── res/
│   │   │   └── AndroidManifest.xml
│   │   ├── libs/
│   │   │   └── muxuetools.aar    # Go 编译产物
│   │   └── build.gradle.kts
│   ├── build.gradle.kts
│   ├── settings.gradle.kts
│   └── .gitignore                # Android 专用 gitignore
├── mobile/                    # 新增: Go mobile 入口包
│   ├── mobile.go
│   └── server.go
├── cmd/                       # 现有
├── internal/                  # 现有
├── web/                       # 现有
├── scripts/
│   ├── build.ps1             # 现有
│   └── build-android.ps1     # 新增
├── go.mod                     # 现有
└── embed.go                   # 现有
```

#### 2.2 更新根目录 .gitignore

在项目根目录的 `.gitignore` 中添加：
```gitignore
# Android
android/.gradle/
android/.idea/
android/app/build/
android/build/
android/local.properties
*.aar
```

#### 2.3 配置 build.gradle

```kotlin
android {
    namespace = "com.muxue.tools"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.muxue.tools"
        minSdk = 31
        targetSdk = 35
        versionCode = 1
        versionName = "1.0.0"
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(...)
        }
    }
}

dependencies {
    implementation(files("libs/muxuetools.aar"))
}
```

#### 2.4 AndroidManifest.xml

```xml
<manifest xmlns:android="http://schemas.android.com/apk/res/android">
    <!-- 权限 -->
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
    <uses-permission android:name="android.permission.FOREGROUND_SERVICE_SPECIAL_USE" />
    <uses-permission android:name="android.permission.POST_NOTIFICATIONS" />

    <application
        android:name=".MuxueApplication"
        android:icon="@mipmap/ic_launcher"
        android:label="@string/app_name"
        android:usesCleartextTraffic="true">
        
        <activity
            android:name=".MainActivity"
            android:exported="true"
            android:launchMode="singleTask">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>

        <service
            android:name=".ServerService"
            android:foregroundServiceType="specialUse"
            android:exported="false" />
    </application>
</manifest>
```

---

### Phase 3: WebView 集成

#### 3.1 MainActivity.kt

```kotlin
class MainActivity : AppCompatActivity() {
    private lateinit var webView: WebView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // 启动后台服务
        startServerService()
        
        // 配置 WebView
        webView = WebView(this).apply {
            settings.javaScriptEnabled = true
            settings.domStorageEnabled = true
            settings.allowFileAccess = true
        }
        setContentView(webView)
        
        // 等待服务器启动后加载
        loadWebUI()
    }

    private fun loadWebUI() {
        // 加载本地服务器页面
        webView.loadUrl("http://127.0.0.1:8080")
    }
}
```

#### 3.2 ServerService.kt (前台服务)

```kotlin
class ServerService : Service() {
    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        // 创建通知渠道 (Android 13+)
        createNotificationChannel()
        
        // 启动前台服务
        val notification = buildNotification()
        startForeground(NOTIFICATION_ID, notification)
        
        // 启动 Go 服务器
        val dataDir = filesDir.absolutePath
        Mobile.startServer("127.0.0.1:8080", dataDir)
        
        return START_STICKY
    }
}
```

---

### Phase 4: 打包与测试

#### 4.1 构建脚本

创建 `scripts/build-android.ps1`:

```powershell
# 1. 编译 Go AAR
gomobile bind -target=android/arm64,android/arm -androidapi 31 -o android/app/libs/muxuetools.aar ./mobile

# 2. 构建前端
cd web
npm run build
cd ..

# 3. 构建 APK (需要 Android Studio 命令行工具)
cd android
./gradlew assembleRelease
```

#### 4.2 APK 签名

```bash
# 生成签名密钥
keytool -genkey -v -keystore muxuetools.keystore -alias muxuetools -keyalg RSA -keysize 2048 -validity 10000

# 签名 APK
jarsigner -verbose -sigalg SHA256withRSA -digestalg SHA-256 -keystore muxuetools.keystore app-release-unsigned.apk muxuetools
```

#### 4.3 测试清单

| 测试项 | 说明 |
|--------|------|
| 安装测试 | APK 可正常安装 |
| 服务器启动 | 后台服务正常启动 |
| WebView 加载 | 界面正常显示 |
| 功能测试 | Chat/Key/Stats 功能正常 |
| 后台保活 | 切换到后台后服务继续运行 |
| 多架构 | arm64 和 arm 设备都能运行 |

---

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `mobile/mobile.go` | **NEW** | Go Android 入口 |
| `mobile/server.go` | **NEW** | 服务器封装 |
| `android/` | **NEW** | Android 项目目录 |
| `android/app/src/main/java/.../MainActivity.kt` | **NEW** | 主界面 |
| `android/app/src/main/java/.../ServerService.kt` | **NEW** | 后台服务 |
| `android/app/src/main/AndroidManifest.xml` | **NEW** | 清单文件 |
| `scripts/build-android.ps1` | **NEW** | 构建脚本 |
| `internal/config/loader.go` | **MODIFY** | 适配 Android 路径 |

---

## 约束

### 技术约束
- Go 1.22+
- gomobile (golang.org/x/mobile)
- Android SDK API 31+
- Android NDK r25+
- Kotlin 1.9+

### 质量约束
- 遵循 `.agent/skills/senior-golang/SKILL.md` 代码规范
- Android 代码遵循 Kotlin 官方规范
- APK 体积控制在 30MB 以内

### 兼容性约束
- 支持 Android 12-15 (API 31-35)
- 支持 arm64-v8a 和 armeabi-v7a 架构
- 保持与桌面版功能一致

---

## 验收标准

- [ ] `gomobile bind` 成功编译 `.aar` 文件
- [ ] Android Studio 项目可正常打开和编译
- [ ] APK 可在 Android 12+ 设备上安装
- [ ] 应用启动后显示 WebView 界面
- [ ] 后台服务正常运行
- [ ] Chat 功能正常 (流式响应)
- [ ] Key 管理功能正常
- [ ] 设置保存正常 (使用 Android 内部存储)
- [ ] 应用切换到后台后服务不被杀死
- [ ] `./gradlew assembleRelease` 生成签名 APK

---

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/IMPLEMENTATION_PLAN.md` | 标记阶段六为已完成 |
| `docs/FRONTEND_TASKS.md` | 标记任务 9 为已完成 |
| `docs/README.md` | 新增 Android 构建说明 |
| `README.md` | 新增 Android 使用说明 |

---

## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的标准开发流程。

---

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| gomobile 编译失败 | 阻塞 | 确保 NDK 版本正确，清理缓存重试 |
| AAR 体积过大 | 用户体验 | 使用 `-ldflags="-s -w"` 压缩 |
| 后台服务被杀 | 功能异常 | 使用前台服务 + 通知 |
| WebView 兼容性 | 界面异常 | 使用系统 WebView，无需自带 |

---

## 参考资料

| 资源 | 链接 |
|------|------|
| gomobile 官方文档 | https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile |
| Android 前台服务 | https://developer.android.com/develop/background-work/services/foreground-services |
| WebView 最佳实践 | https://developer.android.com/develop/ui/views/layout/webapps/webview |
| Kotlin 协程 | https://kotlinlang.org/docs/coroutines-overview.html |

---

*任务创建时间: 2026-01-20*
