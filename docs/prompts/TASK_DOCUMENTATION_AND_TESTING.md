# 任务：文档更新与测试验证（Phase 6）

## 角色
Developer (senior-golang skill)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`

## 背景

MuxueTools 跨平台适配工作已完成所有核心实施阶段：

- **Phase 1**: 数据路径重构 ✅ 
- **Phase 2**: Linux 构建系统 ✅ 
- **Phase 3**: Windows 安装程序 ✅
- **Phase 4**: Linux 安装包 ✅
- **Phase 5**: CI/CD 集成 ✅

现在需要完成最后的文档更新和全面测试验证，确保项目交付质量。

## 目标

| 任务 | 目标 | 优先级 |
|------|------|--------|
| **6.1** | 更新 `docs/ARCHITECTURE.md` 添加安装程序架构 | ⭐⭐ 中 |
| **6.2** | 更新 `UBUNTU_ADAPTATION_PLAN.md` 完成状态 | ⭐ 低 |
| **6.3** | Windows 安装程序本地测试 | ⭐⭐⭐ 高 |
| **6.4** | CI/CD 流程验证（可选，需推送 tag） | ⭐⭐ 中 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **项目文档**
   - `docs/ARCHITECTURE.md` - 现有架构文档
   - `docs/UBUNTU_ADAPTATION_PLAN.md` - 适配计划
   - `README.md` - 项目说明

2. **相关代码**
   - `scripts/installer/` - 安装程序脚本目录

---

### 阶段 1：更新 ARCHITECTURE.md

#### 6.1.1 更新目录结构

在 Section 2 目录结构中添加 `scripts/installer/` 目录：

```
├── scripts/
│   ├── build.ps1                       # Windows 构建脚本
│   ├── build.sh                        # Linux 构建脚本
│   ├── convert-icon.py                 # PNG → ICO 图标转换脚本
│   └── installer/                      # 安装程序脚本
│       ├── windows/
│       │   └── setup.iss               # Inno Setup 脚本
│       └── linux/
│           ├── DEBIAN/
│           │   ├── control             # 包元信息
│           │   ├── postinst            # 安装后脚本
│           │   └── prerm               # 卸载前脚本
│           ├── build-deb.sh            # .deb 构建脚本
│           └── muxuetools.desktop      # 桌面菜单项
```

#### 6.1.2 添加新章节：安装程序架构

在文档适当位置（Section 9 之后）添加：

```markdown
## 10. 安装程序架构

### 10.1 Windows 安装程序 (Inno Setup)

**工具**: Inno Setup 6.x

**特性**:
- 现代化安装向导界面
- 多语言支持 (中文/英文/日文)
- 自定义安装路径
- 开始菜单 + 桌面快捷方式
- 无需管理员权限 (PrivilegesRequired=lowest)
- 旧版本检测与卸载提示
- 卸载时可选删除用户数据

**数据路径**:
- 安装目录: `%ProgramFiles%\MuxueTools\` (可自定义)
- 用户数据: `%APPDATA%\MuxueTools\`
- 配置文件: `%APPDATA%\MuxueTools\config.yaml`
- 数据库: `%APPDATA%\MuxueTools\data\muxuetools.db`

**构建命令**:
```powershell
.\scripts\build.ps1 installer
```

### 10.2 Linux 安装包 (.deb)

**工具**: dpkg-deb (Ubuntu 原生)

**目标平台**: Ubuntu 24.04 LTS (amd64)

**依赖项**:
- `libgtk-3-0` - GTK3 运行时
- `libwebkit2gtk-4.1-0` - WebKit2GTK 运行时

**安装路径**:
- 程序目录: `/opt/muxuetools/`
- 启动脚本: `/usr/local/bin/muxuetools`
- 桌面文件: `/usr/share/applications/muxuetools.desktop`

**数据路径** (XDG 规范):
- 用户数据: `~/.local/share/muxuetools/`
- 配置文件: `~/.config/muxuetools/config.yaml`
- 数据库: `~/.local/share/muxuetools/muxuetools.db`

**构建命令**:
```bash
cd scripts/installer/linux
./build-deb.sh 0.3.1
```

### 10.3 CI/CD 自动构建

GitHub Actions 工作流在推送 `v*` 标签时自动构建：

| 产物 | 平台 | 描述 |
|------|------|------|
| `MuxueTools-Setup-x.x.x.exe` | Windows | 安装程序 |
| `muxueTools-windows-amd64.zip` | Windows | 便携版 |
| `muxueTools-linux-amd64.tar.gz` | Linux | 二进制 |
| `muxuetools_x.x.x_amd64.deb` | Linux | Debian 包 |

**工作流架构**:
```
┌─────────────────┐     ┌─────────────────┐
│  build-windows  │     │   build-linux   │
│ (windows-latest)│     │  (ubuntu-24.04) │
└────────┬────────┘     └────────┬────────┘
         │                       │
         └───────────┬───────────┘
                     │
              ┌──────▼──────┐
              │   release   │
              │(ubuntu-latest)│
              └──────┬──────┘
                     │
              GitHub Release
```
```

---

### 阶段 2：测试验证

#### 6.3 Windows 安装程序本地测试

**前置条件**:
1. 安装 Inno Setup 6
2. 下载中文语言包

**测试步骤**:

```powershell
# Step 1: 构建安装程序
.\scripts\build.ps1 installer

# Step 2: 验证产物
Test-Path dist\MuxueTools-Setup-*.exe

# Step 3: 运行安装程序
# - 选择安装路径
# - 完成安装

# Step 4: 验证安装
# - 检查开始菜单快捷方式
# - 检查桌面快捷方式 (如果选择创建)
# - 运行程序，验证启动成功
# - 检查 %APPDATA%\MuxueTools\ 目录创建

# Step 5: 测试卸载
# - 通过开始菜单或控制面板卸载
# - 验证是否询问删除用户数据
# - 验证卸载完成
```

**验证清单**:
- [ ] 安装程序正常启动
- [ ] 支持中文界面
- [ ] 可自定义安装路径
- [ ] 程序可正常运行
- [ ] 用户数据写入 `%APPDATA%\MuxueTools\`
- [ ] 卸载程序正常工作
- [ ] 卸载时询问是否删除用户数据

#### 6.4 CI/CD 流程验证 (可选)

如果需要完整验证 CI/CD 流程，可以推送测试标签：

```bash
# 创建测试标签
git tag -a v0.3.1-test -m "Test release"
git push origin v0.3.1-test

# 检查 GitHub Actions 运行状态
# 验证 Release 页面产物

# 清理测试标签
git tag -d v0.3.1-test
git push origin :refs/tags/v0.3.1-test
```

---

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `docs/ARCHITECTURE.md` | **MODIFY** | 添加安装程序架构章节 |
| `docs/UBUNTU_ADAPTATION_PLAN.md` | **MODIFY** | 更新 Phase 6 状态 |

## 约束

### 质量约束
- 文档更新需保持现有格式风格
- 代码块使用正确的语言标识符
- 目录结构树与实际一致

### 兼容性约束
- 不修改任何代码文件
- 仅文档更新

## 验收标准

### 文档更新
- [x] `ARCHITECTURE.md` 目录结构包含 `scripts/installer/`
- [x] `ARCHITECTURE.md` 包含安装程序架构章节
- [x] `UBUNTU_ADAPTATION_PLAN.md` Phase 6 状态更新为已完成

### 测试验证
- [x] Windows 安装程序可正常构建
- [x] Windows 安装程序可正常安装/运行/卸载
- [x] 用户数据路径正确 (`%APPDATA%\MuxueTools\`)

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/ARCHITECTURE.md` | 新增安装程序架构章节 |
| `docs/UBUNTU_ADAPTATION_PLAN.md` | 更新状态为 100% 完成 |

## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的开发流程。

---

*任务创建时间: 2026-01-26*
*任务完成时间: 2026-01-26*

