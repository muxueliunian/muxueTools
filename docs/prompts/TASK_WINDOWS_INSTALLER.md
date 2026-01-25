# 任务：实现 Windows 安装程序（Phase 3）

## 角色
Developer (senior-golang skill)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`

## 背景

MuxueTools 当前采用 ZIP 压缩包分发，用户需手动解压运行。为提升用户体验，需要创建 Windows 安装程序。

**目标：**
- 使用 Inno Setup 创建 Windows 安装程序
- 支持自定义安装路径
- 创建开始菜单和桌面快捷方式
- 支持卸载程序

**依赖完成：**
- Phase 1: 数据路径重构 ✅
- Phase 2: Linux 构建系统 ✅

## 目标

| 任务 | 目标 | 优先级 |
|------|------|--------|
| **3.1** | 新增 `scripts/installer/windows/setup.iss` | ⭐⭐⭐ 高 |
| **3.2** | 更新 `scripts/build.ps1` 添加安装程序构建 | ⭐⭐ 中 |
| **3.3** | 验证安装程序功能 | ⭐⭐ 中 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **项目文档**
   - `docs/UBUNTU_ADAPTATION_PLAN.md` - 完整适配计划（Section 3.1）
   - `scripts/build.ps1` - 现有构建脚本

2. **外部参考**
   - [Inno Setup 官方文档](https://jrsoftware.org/ishelp/)

### 阶段 1：创建 Inno Setup 脚本

**新增目录**: `scripts/installer/windows/`

**新增文件**: `scripts/installer/windows/setup.iss`

```iss
; MuxueTools Windows Installer Script
; Inno Setup 6.x

#define MyAppName "MuxueTools"
#define MyAppVersion GetEnv('VERSION')
#if MyAppVersion == ""
  #define MyAppVersion "0.3.1"
#endif
#define MyAppPublisher "muxueliunian"
#define MyAppURL "https://github.com/muxueliunian/muxueTools"
#define MyAppExeName "muxueTools.exe"

[Setup]
; 生成 GUID: PowerShell [guid]::NewGuid()
AppId={{8F42A5E3-XXXX-XXXX-XXXX-XXXXXXXXXXXX}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
LicenseFile=..\..\..\LICENSE
OutputDir=..\..\..\dist
OutputBaseFilename=MuxueTools-Setup-{#MyAppVersion}
SetupIconFile=..\..\..\assets\icon.ico
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

; 权限设置 (不需要管理员权限)
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog

[Languages]
Name: "chinesesimplified"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "japanese"; MessagesFile: "compiler:Languages\Japanese.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
; 主程序
Source: "..\..\..\bin\muxueTools.exe"; DestDir: "{app}"; Flags: ignoreversion

; 前端资源
Source: "..\..\..\web\dist\*"; DestDir: "{app}\web\dist"; Flags: ignoreversion recursesubdirs createallsubdirs

; 配置模板
Source: "..\..\..\configs\config.example.yaml"; DestDir: "{app}"; DestName: "config.example.yaml"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Code]
// 安装完成后打开浏览器（可选）
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    // 可以添加后处理逻辑
  end;
end;
```

### 阶段 2：更新构建脚本

**修改文件**: `scripts/build.ps1`

添加新的构建目标 `installer`：

```powershell
function Build-Installer {
    Write-Header "Building Windows Installer (Inno Setup)"
    
    # Check if Inno Setup is installed
    $isccPath = Get-Command iscc -ErrorAction SilentlyContinue
    if (-not $isccPath) {
        # Try common installation paths
        $commonPaths = @(
            "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
            "C:\Program Files\Inno Setup 6\ISCC.exe"
        )
        foreach ($path in $commonPaths) {
            if (Test-Path $path) {
                $isccPath = $path
                break
            }
        }
    }
    
    if (-not $isccPath) {
        Write-Host "Error: Inno Setup not found. Please install from https://jrsoftware.org/isdl.php" -ForegroundColor Red
        return
    }
    
    # Build installer
    $setupScript = Join-Path $ProjectRoot "scripts\installer\windows\setup.iss"
    
    if (-not (Test-Path $setupScript)) {
        Write-Host "Error: setup.iss not found at $setupScript" -ForegroundColor Red
        return
    }
    
    Write-Host "Running Inno Setup compiler..." -ForegroundColor Cyan
    & $isccPath $setupScript /DVersion=$Version
    
    Write-Host "Installer built successfully!" -ForegroundColor Green
}
```

更新 switch 语句：

```powershell
switch ($Target) {
    # ... existing cases ...
    "installer" {
        # Build desktop first, then create installer
        $frontendOk = Build-Frontend
        if ($frontendOk) {
            Build-Desktop -Arch "amd64"
            Build-Installer
        }
    }
}
```

### 阶段 3：验证安装程序

```powershell
# 1. 安装 Inno Setup
# 下载: https://jrsoftware.org/isdl.php

# 2. 构建安装程序
.\scripts\build.ps1 installer

# 3. 运行安装程序
.\dist\MuxueTools-Setup-0.3.1.exe

# 4. 验证安装
# - 检查 C:\Program Files\MuxueTools\ 目录
# - 检查开始菜单快捷方式
# - 运行程序验证功能
# - 测试卸载程序
```

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `scripts/installer/windows/setup.iss` | **NEW** | Inno Setup 脚本 |
| `scripts/build.ps1` | **MODIFY** | 添加 `installer` 构建目标 |

## 约束

### 技术约束
- Inno Setup 6.x
- Windows 10/11 x64

### 质量约束
- 支持中文、英文、日文界面
- 不需要管理员权限安装（使用 PrivilegesRequired=lowest）
- 安装包大小 < 50MB

### 兼容性约束
- 与便携版共存（安装版数据在 %APPDATA%，便携版在当前目录）

## 验收标准

- [x] `scripts/installer/windows/setup.iss` 语法正确
- [x] `.\scripts\build.ps1 installer` 成功生成 `dist/MuxueTools-Setup-*.exe`
- [x] 安装程序可正常运行 (支持中/日/英三语)
- [x] 安装后程序可正常启动
- [x] 开始菜单快捷方式创建成功
- [x] 卸载程序可正常运行
- [x] 卸载时可选择删除用户数据
- [x] 数据正确写入 `%APPDATA%\MuxueTools\`

## 交付文档

| 文档 | 更新内容 | 状态 |
|------|----------|------|
| `docs/UBUNTU_ADAPTATION_PLAN.md` | 更新 Phase 3 状态为已完成 | ✅ |
| `README.md` | 添加安装程序构建说明 | ✅ |

## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的开发流程。

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Inno Setup 未安装 | 无法构建安装程序 | 脚本检测并提示下载链接 |
| 签名缺失 | Windows SmartScreen 警告 | 提醒用户信任或后续添加代码签名 |
| 路径冲突 | 旧版本残留 | 安装前提示卸载旧版本 |

---

## ✅ 任务完成总结

**完成日期**: 2026-01-26

### 实现的功能

| 功能 | 说明 |
|------|------|
| **安装程序** | 使用 Inno Setup 6 创建现代风格安装向导 |
| **多语言支持** | 中文(默认)、英文、日文界面 |
| **自定义安装路径** | 用户可选择安装位置 |
| **快捷方式** | 开始菜单 + 桌面(可选) |
| **无需管理员权限** | 使用 PrivilegesRequired=lowest |
| **旧版本检测** | 安装时检测并提示卸载旧版本 |
| **进程检测** | 安装/卸载前检测程序是否运行，自动关闭 |
| **用户数据管理** | 卸载时询问是否删除用户数据(API密钥、对话历史) |
| **删除失败处理** | 删除失败时提示结束进程并重试 |

### 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `scripts/installer/windows/setup.iss` | **NEW** | Inno Setup 脚本 (约 370 行) |
| `scripts/build.ps1` | **MODIFY** | 添加 `Build-Installer` 函数和 `installer` 目标 |
| `README.md` | **MODIFY** | 添加安装程序构建说明 |
| `docs/UBUNTU_ADAPTATION_PLAN.md` | **MODIFY** | 更新 Phase 3 状态 |

### 构建命令

```powershell
# 构建 Windows 安装程序 (需要安装 Inno Setup 6)
.\scripts\build.ps1 installer
```

### 前置条件

1. 安装 [Inno Setup 6](https://jrsoftware.org/isdl.php)
2. 下载简体中文语言包:
   ```powershell
   Start-Process powershell -Verb RunAs -ArgumentList "-Command `"Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/jrsoftware/issrc/main/Files/Languages/Unofficial/ChineseSimplified.isl' -OutFile 'C:\Program Files (x86)\Inno Setup 6\Languages\ChineseSimplified.isl'`""
   ```

---

*任务创建时间: 2026-01-25*
*任务完成时间: 2026-01-26*

