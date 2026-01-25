# Ubuntu 测试清单

> 待所有 Phase 完成后统一执行

---

## 前置条件

```bash
# 安装编译依赖
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev

# 安装 Go 1.22+
# 安装 Node.js 18+
```

---

## Phase 2: Linux 构建系统测试

```bash
# 赋予执行权限
chmod +x scripts/build.sh

# 构建所有目标
./scripts/build.sh all

# 验证产物
ls -la bin/
# 期望: muxuetools-server, muxuetools-desktop

# 运行测试
./bin/muxuetools-server --version
./bin/muxuetools-desktop --version
```

**验收标准**:
- [ ] `build.sh all` 无错误执行
- [ ] 生成 `bin/muxuetools-server`
- [ ] 生成 `bin/muxuetools-desktop`
- [ ] `--version` 显示正确版本

---

## Phase 4: Linux .deb 安装包测试

```bash
# 构建 .deb 包
cd scripts/installer/linux
chmod +x build-deb.sh
./build-deb.sh 0.3.1

# 安装
sudo dpkg -i ../../../dist/muxuetools_0.3.1_amd64.deb

# 安装缺失依赖 (如有)
sudo apt-get install -f

# 验证安装
muxuetools --version
which muxuetools
ls -la /opt/muxuetools/

# 验证桌面集成
ls /usr/share/applications/muxuetools.desktop
ls /usr/share/icons/hicolor/256x256/apps/muxuetools.png

# 启动程序
muxuetools

# 验证数据目录创建
ls ~/.local/share/muxuetools/
ls ~/.config/muxuetools/

# 卸载测试
sudo dpkg -r muxuetools

# 验证卸载后用户数据保留
ls ~/.local/share/muxuetools/
```

**验收标准**:
- [ ] `build-deb.sh` 成功生成 `dist/muxuetools_*.deb`
- [ ] `dpkg -i` 安装成功
- [ ] `muxuetools --version` 显示正确版本
- [ ] 程序可正常启动并显示 GUI
- [ ] 应用程序菜单中可见 MuxueTools 图标
- [ ] `dpkg -r` 卸载成功
- [ ] 卸载后用户数据保留

---

## Phase 5: CI/CD 测试 (GitHub Actions)

在本地无法测试，推送 tag 后由 GitHub Actions 自动执行。

---

*创建时间: 2026-01-26*
