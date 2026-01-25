# Ubuntu 双系统测试指南

## 概述

由于 Desktop 版本依赖 CGO (WebView + GTK)，无法在 Windows 上交叉编译。
测试流程：**在 Ubuntu 系统上构建并测试**。

## 前置准备 (Windows)

### 1. 同步代码到 Ubuntu 可访问的位置

**方式 A: 共享分区**
如果您的双系统共享了分区（如 NTFS 数据盘），可以直接访问项目目录。

**方式 B: Git 同步**
```powershell
# 提交当前更改
cd c:\Users\muxueliunian\Desktop\gugugagaApi
git add .
git commit -m "Ubuntu test preparation"
git push origin main
```

**方式 C: U盘/网盘**
将整个项目目录复制到 U 盘或网盘。

---

## Ubuntu 系统测试步骤

### Phase 1: 环境准备

```bash
# 1. 更新系统
sudo apt-get update && sudo apt-get upgrade -y

# 2. 安装 Go (如果未安装)
# 方式 A: 使用 apt (可能版本较旧)
sudo apt-get install -y golang-go

# 方式 B: 安装最新版 Go (推荐)
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version  # 应显示 go1.22.0

# 3. 安装 Node.js (如果未安装)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
node --version  # 应显示 v20.x.x
npm --version

# 4. 安装 GTK3 和 WebKit2GTK 开发依赖
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev

# 5. 安装构建工具
sudo apt-get install -y build-essential git
```

### Phase 2: 获取代码

```bash
# 方式 A: 从 Git 克隆
git clone https://github.com/muxueliunian/muxueTools.git
cd muxueTools

# 方式 B: 从共享分区访问 (挂载 Windows 分区)
# 先挂载 Windows 分区
sudo mount /dev/sda3 /mnt/windows  # 根据实际分区调整
cd /mnt/windows/Users/muxueliunian/Desktop/gugugagaApi

# 方式 C: 从 U盘复制
cp -r /media/usb/gugugagaApi ~/muxueTools
cd ~/muxueTools
```

### Phase 3: 构建项目

```bash
# 给构建脚本添加执行权限
chmod +x scripts/build.sh
chmod +x scripts/installer/linux/build-deb.sh

# 查看帮助
./scripts/build.sh --help

# 构建所有目标 (Frontend + Server + Desktop)
./scripts/build.sh all

# 检查产物
ls -la bin/
# 应该看到:
# - muxuetools-server   (纯 Go 服务端)
# - muxuetools-desktop  (带 GUI 桌面版)
```

### Phase 4: 测试 Server 版本

```bash
# 运行 Server
./bin/muxuetools-server --version
./bin/muxuetools-server &

# 测试 API
curl http://localhost:8080/health
# 应返回: {"status":"ok",...}

curl http://localhost:8080/v1/models
# 应返回模型列表

# 检查数据目录
ls -la ~/.local/share/muxuetools/
# 应包含: muxuetools.db

ls -la ~/.config/muxuetools/
# 配置目录 (首次运行可能为空)

# 停止服务
pkill muxuetools-server
```

### Phase 5: 测试 Desktop 版本

```bash
# 运行 Desktop (需要桌面环境)
./bin/muxuetools-desktop &

# 验证:
# - [ ] GUI 窗口正常显示
# - [ ] 内置网页加载成功
# - [ ] 可以添加 API Key
# - [ ] 可以发送对话请求

# 关闭程序后检查数据
ls -la ~/.local/share/muxuetools/
```

### Phase 6: 构建并测试 .deb 包

```bash
# 构建 .deb 包
cd scripts/installer/linux
./build-deb.sh 0.3.1

# 查看产物
ls -la ../../../dist/
# 应该看到: muxuetools_0.3.1_amd64.deb

# 安装 .deb 包
sudo dpkg -i ../../../dist/muxuetools_0.3.1_amd64.deb

# 如果提示缺少依赖，执行:
sudo apt-get install -f

# 验证安装
which muxuetools
# 应返回: /usr/local/bin/muxuetools

muxuetools --version
# 应显示版本号

# 检查安装路径
ls -la /opt/muxuetools/
# 应包含:
# - muxuetools (主程序)
# - web/dist/  (前端资源)
# - config.example.yaml

# 从命令行启动
muxuetools

# 从应用程序菜单启动 (如果有桌面环境)
# 打开应用程序菜单，搜索 "MuxueTools"
```

### Phase 7: 测试卸载

```bash
# 卸载
sudo dpkg -r muxuetools

# 验证:
# - [ ] 程序被移除: ls /opt/muxuetools/ 应该失败
# - [ ] 用户数据保留: ls ~/.local/share/muxuetools/ 应该成功
# - [ ] 配置保留: ls ~/.config/muxuetools/ 应该成功 (如果之前创建过)

# 完全清理 (可选)
rm -rf ~/.local/share/muxuetools/
rm -rf ~/.config/muxuetools/
```

---

## 测试检查清单

### 环境准备
- [ ] Go 1.22+ 已安装
- [ ] Node.js 20+ 已安装
- [ ] GTK3 + WebKit2GTK 开发依赖已安装

### 构建测试
- [ ] `./scripts/build.sh all` 成功
- [ ] `bin/muxuetools-server` 存在
- [ ] `bin/muxuetools-desktop` 存在

### Server 版本测试
- [ ] `--version` 显示正确版本
- [ ] 程序启动成功
- [ ] `curl localhost:8080/health` 返回成功
- [ ] 数据写入 `~/.local/share/muxuetools/`

### Desktop 版本测试
- [ ] GUI 窗口正常显示
- [ ] 内置网页加载正常
- [ ] 添加 Key 功能正常
- [ ] 数据写入 `~/.local/share/muxuetools/`

### .deb 包测试
- [ ] `./build-deb.sh` 成功生成 .deb
- [ ] `sudo dpkg -i *.deb` 安装成功
- [ ] `muxuetools --version` 正常
- [ ] 程序可从 /usr/local/bin/muxuetools 启动
- [ ] 应用程序菜单中可见 (如有桌面环境)
- [ ] `sudo dpkg -r muxuetools` 卸载成功
- [ ] 卸载后用户数据保留

---

## 常见问题

### Q1: WebKit 依赖找不到
```bash
# Ubuntu 24.04 使用 4.1 版本
sudo apt-get install libwebkit2gtk-4.1-dev

# Ubuntu 22.04 可能需要 4.0 版本
sudo apt-get install libwebkit2gtk-4.0-dev
```

### Q2: 构建时 CGO 报错
```bash
# 确保设置了 CGO_ENABLED
export CGO_ENABLED=1
./scripts/build.sh desktop
```

### Q3: 程序启动后窗口不显示
```bash
# 确保在桌面环境中运行，不是纯命令行
# 检查 DISPLAY 环境变量
echo $DISPLAY
# 应该显示类似 :0 或 :1
```

### Q4: .deb 安装后找不到命令
```bash
# 检查 PATH
echo $PATH
# 应该包含 /usr/local/bin

# 手动运行
/opt/muxuetools/muxuetools
```

---

*文档创建时间: 2026-01-26*
