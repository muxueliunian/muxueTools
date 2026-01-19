# MuxueTools (muxueTools)

MuxueTools 是一�?OpenAI 兼容�?Gemini API 代理，支持多 Key 轮询、会话管理和内置聊天界面�?

## 快速开�?

下载最新版本：[Releases](https://github.com/muxueliunian/muxueTools/releases)

## CI/CD 自动化部�?

本项目使�?GitHub Actions 实现自动化构建和发布�?

### 触发条件

当推送以 `v` 开头的 tag 时自动触发：

```bash
git tag v1.0.0
git push origin v1.0.0
```

### 自动化流�?

```
推�?v* Tag �?构建 �?打包 �?FTP 上传 �?创建 Release
```

| 步骤 | 描述 |
|------|------|
| **构建前端** | `npm ci && npm run build` |
| **构建后端** | Windows AMD64 可执行文�?|
| **打包** | 生成 ZIP 压缩�?|
| **生成 latest.json** | 自动生成版本信息文件 |
| **FTP 上传** | 上传�?mxlnuma.space 服务�?|
| **GitHub Release** | 创建 Release 并上传构建产�?|

### 更新服务

应用支持双源更新检查：

| 更新�?| URL |
|--------|-----|
| mxln 服务�?| `https://mxlnuma.space/muxueTools/update/latest.json` |
| GitHub Releases | GitHub API |

### 所需 Secrets

在仓�?Settings �?Secrets �?Actions 中配置：

| Secret | 描述 |
|--------|------|
| `FTP_SERVER` | FTP 服务器地址 |
| `FTP_USERNAME_TOOLS` | FTP 用户�?|
| `FTP_PASSWORD_TOOLS` | FTP 密码 |

### 发布新版�?

```bash
# 1. 提交代码
git add .
git commit -m "Release v1.0.0"
git push origin main

# 2. 创建并推�?tag
git tag v1.0.0
git push origin v1.0.0

# 3. 等待 Actions 完成，检�?Releases 页面
```

## 功能特�?

- �?OpenAI 兼容 API 代理
- �?�?Key 轮询管理  
- �?内置聊天界面
- �?统计数据看板
- �?配置持久�?
- �?自动更新检�?

## 许可�?

MIT License
