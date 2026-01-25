# 任务：CI/CD 集成安装程序构建（Phase 5）

## 角色
Developer (senior-golang skill)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`

## 背景

MuxueTools 已完成以下跨平台适配工作：
- **Phase 1**: 数据路径重构 ✅ 
- **Phase 2**: Linux 构建系统 ✅ 
- **Phase 3**: Windows 安装程序 ✅ (`scripts/installer/windows/setup.iss`)
- **Phase 4**: Linux 安装包 ✅ (`scripts/installer/linux/build-deb.sh`)

当前 CI/CD 工作流仅构建 Windows 便携版 ZIP 包，需要扩展以支持：
1. Windows 安装程序 (.exe Setup)
2. Linux 二进制与 .deb 安装包

**现有工作流**：`.github/workflows/release.yml`
- 触发条件：推送 `v*` 标签
- 当前产物：`muxueTools-windows-amd64.zip`
- 部署方式：FTP + GitHub Release

## 目标

| 任务 | 目标 | 优先级 |
|------|------|--------|
| **5.1** | 添加 Windows 安装程序构建步骤 | ⭐⭐⭐ 高 |
| **5.2** | 添加 Linux 构建 Job | ⭐⭐⭐ 高 |
| **5.3** | 添加 .deb 包构建步骤 | ⭐⭐ 中 |
| **5.4** | 更新 GitHub Release 产物列表 | ⭐⭐ 中 |
| **5.5** | 更新 `latest.json` 下载链接 | ⭐ 低 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **项目文档**
   - `docs/UBUNTU_ADAPTATION_PLAN.md` - 完整适配计划（Section 4 CI/CD）
   - `.github/workflows/release.yml` - 现有工作流
   - `scripts/build.ps1` - Windows 构建脚本 (installer 目标)
   - `scripts/installer/linux/build-deb.sh` - Linux 打包脚本

2. **外部参考**
   - [GitHub Actions 文档](https://docs.github.com/en/actions)
   - [action-gh-release](https://github.com/softprops/action-gh-release)

### 阶段 1：Windows 安装程序构建

在 `build-windows` job 中添加 Inno Setup 构建步骤：

```yaml
      - name: Install Inno Setup
        run: choco install innosetup -y

      - name: Download Chinese Language File
        shell: pwsh
        run: |
          $langPath = "C:\Program Files (x86)\Inno Setup 6\Languages"
          Invoke-WebRequest -Uri "https://raw.githubusercontent.com/jrsoftware/issrc/main/Files/Languages/Unofficial/ChineseSimplified.isl" -OutFile "$langPath\ChineseSimplified.isl"

      - name: Build Windows Installer
        shell: pwsh
        run: |
          $env:VERSION = "${{ steps.version.outputs.VERSION }}"
          & "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DVersion=${{ steps.version.outputs.VERSION }} scripts\installer\windows\setup.iss
```

### 阶段 2：添加 Linux 构建 Job

新增 `build-linux` job：

```yaml
  build-linux:
    runs-on: ubuntu-24.04
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
      
      - name: Get version from tag
        id: version
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "VERSION=$VERSION" >> $GITHUB_OUTPUT
          echo "TAG=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
      
      - name: Install GTK/WebKit dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
      
      - name: Build Frontend
        run: |
          cd web
          npm ci
          npm run build
      
      - name: Build Linux Desktop
        run: |
          chmod +x scripts/build.sh
          VERSION=${{ steps.version.outputs.VERSION }} ./scripts/build.sh desktop
      
      - name: Build .deb Package
        run: |
          chmod +x scripts/installer/linux/build-deb.sh
          cd scripts/installer/linux
          ./build-deb.sh ${{ steps.version.outputs.VERSION }}
      
      - name: Create tarball
        run: |
          mkdir -p dist
          tar -czvf dist/muxueTools-linux-amd64.tar.gz -C bin muxuetools-desktop
      
      - name: Upload Linux artifacts
        uses: actions/upload-artifact@v4
        with:
          name: linux-packages
          path: |
            dist/*.tar.gz
            dist/*.deb
```

### 阶段 3：合并 Release Job

创建独立的 `release` job，收集所有平台产物：

```yaml
  release:
    needs: [build-windows, build-linux]
    runs-on: ubuntu-latest
    
    steps:
      - name: Download Windows artifacts
        uses: actions/download-artifact@v4
        with:
          name: windows-packages
          path: dist/
      
      - name: Download Linux artifacts
        uses: actions/download-artifact@v4
        with:
          name: linux-packages
          path: dist/
      
      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            dist/MuxueTools-Setup-*.exe
            dist/muxueTools-windows-amd64.zip
            dist/muxueTools-linux-amd64.tar.gz
            dist/muxuetools_*.deb
          generate_release_notes: true
```

### 阶段 4：更新 latest.json

更新自动更新元数据，包含所有平台下载链接：

```json
{
  "version": "$VERSION",
  "release_date": "$DATE",
  "downloads": {
    "windows-amd64": "...",
    "windows-amd64-installer": "...",
    "linux-amd64": "...",
    "linux-amd64-deb": "..."
  }
}
```

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `.github/workflows/release.yml` | **MODIFY** | 添加安装程序构建和 Linux 支持 |

## 约束

### 技术约束
- GitHub Actions runners: `windows-latest`, `ubuntu-24.04`
- Go 1.22+, Node.js 20
- Inno Setup 6 (Windows)
- dpkg-deb (Ubuntu)

### 质量约束
- 所有构建步骤需要有明确的错误处理
- 使用 `actions/upload-artifact@v4` 传递跨 job 产物
- Release 产物命名遵循统一规范

### 兼容性约束
- 保持现有 FTP 部署逻辑（如需要）
- 确保 latest.json 向后兼容

## 验收标准

- [ ] 推送 `v*.*.*` 标签时工作流正常触发
- [ ] Windows 构建 job 成功生成 `.exe` 安装程序
- [ ] Linux 构建 job 成功生成 `.deb` 包和 `.tar.gz`
- [ ] GitHub Release 包含所有 4 种产物
- [ ] latest.json 包含所有下载链接
- [ ] 工作流总耗时 < 15 分钟

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/UBUNTU_ADAPTATION_PLAN.md` | 更新 Phase 5 状态为已完成 |
| `README.md` | 添加 CI/CD 构建说明 (可选) |

## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的开发流程。

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Inno Setup 安装失败 | Windows 构建失败 | 使用 `choco install` 并检查退出码 |
| 中文语言文件缺失 | 安装程序无中文 | 从 GitHub 下载并放入 Languages 目录 |
| GTK/WebKit 版本问题 | Linux 构建失败 | 使用 `ubuntu-24.04` runner |
| 跨 job 产物传递 | Release job 无法获取文件 | 使用 artifacts 并正确设置路径 |
| FTP 密钥泄露 | 安全风险 | 使用 GitHub Secrets |

---

## 参考：完整工作流示例

```yaml
name: Build and Release

on:
  push:
    tags:
      - 'v*'

env:
  GO_VERSION: '1.22'
  NODE_VERSION: '20'

permissions:
  contents: write

jobs:
  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
      
      - name: Get version
        id: version
        shell: bash
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "VERSION=$VERSION" >> $GITHUB_OUTPUT
      
      - name: Build Frontend
        run: |
          cd web
          npm ci
          npm run build
      
      - name: Build Desktop
        shell: cmd
        run: |
          set CGO_ENABLED=1
          set GOOS=windows
          set GOARCH=amd64
          go build -ldflags="-X main.Version=${{ steps.version.outputs.VERSION }}" -o bin/muxueTools.exe ./cmd/desktop
      
      - name: Install Inno Setup
        run: choco install innosetup -y
      
      - name: Download Chinese Language
        shell: pwsh
        run: |
          Invoke-WebRequest -Uri "https://raw.githubusercontent.com/jrsoftware/issrc/main/Files/Languages/Unofficial/ChineseSimplified.isl" -OutFile "C:\Program Files (x86)\Inno Setup 6\Languages\ChineseSimplified.isl"
      
      - name: Build Installer
        run: |
          & "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DVersion=${{ steps.version.outputs.VERSION }} scripts\installer\windows\setup.iss
      
      - name: Create ZIP
        shell: pwsh
        run: |
          Compress-Archive -Path bin/muxueTools.exe -DestinationPath dist/muxueTools-windows-amd64.zip
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: windows-packages
          path: dist/

  build-linux:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
      
      - name: Get version
        id: version
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "VERSION=$VERSION" >> $GITHUB_OUTPUT
      
      - name: Install dependencies
        run: sudo apt-get update && sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
      
      - name: Build Frontend
        run: |
          cd web
          npm ci
          npm run build
      
      - name: Build Desktop
        run: |
          chmod +x scripts/build.sh
          VERSION=${{ steps.version.outputs.VERSION }} ./scripts/build.sh desktop
      
      - name: Build deb package
        run: |
          chmod +x scripts/installer/linux/build-deb.sh
          cd scripts/installer/linux
          ./build-deb.sh ${{ steps.version.outputs.VERSION }}
      
      - name: Create tarball
        run: |
          mkdir -p dist
          tar -czvf dist/muxueTools-linux-amd64.tar.gz -C bin muxuetools-desktop
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: linux-packages
          path: dist/

  release:
    needs: [build-windows, build-linux]
    runs-on: ubuntu-latest
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: dist/
          merge-multiple: true
      
      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
          generate_release_notes: true
```

---

*任务创建时间: 2026-01-26*
