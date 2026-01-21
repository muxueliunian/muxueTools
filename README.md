# MuxueTools

<p align="center">
  <img src="web/public/icon.png" alt="MuxueTools Logo" width="128" height="128">
</p>

<p align="center">
  <strong>OpenAI Compatible Gemini API Proxy</strong>
</p>

<p align="center">
  <a href="https://github.com/muxueliunian/muxueTools/releases">
    <img src="https://img.shields.io/github/v/release/muxueliunian/muxueTools?style=flat-square" alt="Release">
  </a>
  <a href="https://github.com/muxueliunian/muxueTools/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License">
  </a>
</p>

---

## Features

- **OpenAI Compatible API** - Seamless integration with existing OpenAI applications
- **Multi-Key Rotation** - Smart load balancing and automatic failover
- **Built-in Chat UI** - Beautiful Claude-style interface
- **Statistics Dashboard** - Real-time API usage monitoring
- **Multi-language Support** - Chinese, English, Japanese
- **Stream/Non-stream Output** - Configurable response mode
- **Session Persistence** - Auto-save chat history
- **Auto Update Detection** - Dual source update (mxln server + GitHub)

---

## Quick Start

### Download

Get the latest version from [Releases](https://github.com/muxueliunian/muxueTools/releases)

### Run

**Windows:**
```bash
.\muxueTools.exe
```

**Linux/macOS:**
```bash
chmod +x muxueTools
./muxueTools
```

### Access

Open browser: `http://localhost:8080`

---

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `POST /v1/chat/completions` | OpenAI compatible chat API |
| `GET /v1/models` | List available models |
| `GET /health` | Health check |
| `GET /api/keys` | Manage API Keys |
| `GET /api/config` | Configuration management |

### Quick Test

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk-mxln-proxy-local"  # No key needed for local proxy
)

response = client.chat.completions.create(
    model="gemini-2.0-flash",
    messages=[{"role": "user", "content": "Hello!"}],
    stream=True
)

for chunk in response:
    print(chunk.choices[0].delta.content, end="")
```

---

## Configuration

Config file: `config.yaml` in the program directory.

```yaml
server:
  port: 8080
  host: "0.0.0.0"

pool:
  strategy: "round_robin"  # round_robin, random, least_used, weighted
  cooldown_seconds: 60
  max_retries: 3

logging:
  level: "info"  # debug, info, warn, error

update:
  enabled: true
  source: "mxln"  # mxln or github

model_settings:
  stream_output: true  # Enable streaming output
  temperature: 1.0
```

### Desktop Version: Fixed Port

By default, the Desktop version uses port **8080**. To change:

1. **Via Settings UI**: Go to **Settings → Security → Server Port**, modify and save, then restart the app
2. **Via config.yaml**: Set `server.port` to your desired port (e.g., `8888`)

> **Note**: If the configured port is in use, the app will automatically fall back to a random available port.

Example for Cursor/third-party integration:
- **API Key**: Your proxy key from Dashboard  
- **Base URL**: `http://127.0.0.1:8080/v1`

---

## Development

### Requirements

- Go 1.22+
- Node.js 18+
- npm or pnpm

### Local Development

```bash
# 1. Clone repository
git clone https://github.com/muxueliunian/muxueTools.git
cd muxueTools

# 2. Install frontend dependencies
cd web
npm install

# 3. Start frontend dev server
npm run dev

# 4. Start backend (new terminal)
cd ..
go run ./cmd/server
```

### Build

```bash
# Frontend build
cd web
npm run build

# Backend build (Windows)
go build -ldflags="-s -w" -o build/muxueTools.exe ./cmd/server

# Desktop build
go build -ldflags="-s -w -H windowsgui" -o build/muxueTools-desktop.exe ./cmd/desktop
```

---

## Release Process

### Version Numbering

Follow [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`

- **MAJOR**: Incompatible API changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Steps

#### Step 1: Update Version Numbers

Update the following files:

| File | Location |
|------|----------|
| `web/package.json` | `"version": "x.x.x"` |
| `cmd/server/main.go` | `Version = "x.x.x"` |
| `cmd/desktop/main.go` | `Version = "x.x.x"` |

#### Step 2: Update CHANGELOG

Add new version changes to `CHANGELOG.md`.

#### Step 3: Verify Build

```bash
# Frontend build
cd web && npm run build

# Backend build
go build ./...

# Test version
go build -o test.exe ./cmd/server && .\test.exe -version
```

#### Step 4: Commit Changes

```bash
git add .
git commit -m "chore: bump version to vX.X.X"
git push origin main
```

#### Step 5: Create Tag

```bash
# Create annotated tag
git tag -a vX.X.X -m "Release vX.X.X - Brief description"

# Push tag
git push origin vX.X.X
```

#### Step 6: Automatic Deployment

After pushing the tag, GitHub Actions will automatically:
1. Build frontend and backend
2. Package into ZIP archive
3. Generate `latest.json` version info
4. Upload to FTP server
5. Create GitHub Release

#### Step 7: Verify Release

- Check [Releases](https://github.com/muxueliunian/muxueTools/releases) page
- Verify download links work
- Test auto-update functionality

---

## CI/CD Configuration

### Automation Flow

```
Push v* Tag -> Build -> Package -> FTP Upload -> Create Release
```

### Required Secrets

Configure in repo Settings -> Secrets -> Actions:

| Secret | Description |
|--------|-------------|
| `FTP_SERVER` | FTP server address |
| `FTP_USERNAME_TOOLS` | FTP username |
| `FTP_PASSWORD_TOOLS` | FTP password |

### Update Service

The app supports dual-source update checking:

| Source | URL |
|--------|-----|
| mxln Server | `https://mxlnuma.space/muxueTools/update/latest.json` |
| GitHub | GitHub Releases API |

---

## Project Structure

```
muxueTools/
├── cmd/
│   ├── server/      # Server entry point
│   └── desktop/     # Desktop app entry point
├── internal/
│   ├── api/         # HTTP handlers
│   ├── config/      # Configuration management
│   ├── gemini/      # Gemini client
│   ├── keypool/     # Key pool management
│   ├── storage/     # Data persistence
│   └── types/       # Type definitions
├── web/             # Vue3 frontend
│   ├── src/
│   │   ├── api/         # API client
│   │   ├── components/  # UI components
│   │   ├── views/       # Pages
│   │   ├── stores/      # Pinia stores
│   │   └── i18n/        # Internationalization
│   └── dist/        # Build output
├── docs/            # Documentation
├── scripts/         # Build scripts
└── .github/
    └── workflows/   # CI/CD config
```

---

## Documentation

See [docs/](./docs/) directory:

- [API Documentation](./docs/API.md)
- [Architecture](./docs/ARCHITECTURE.md)
- [Development Guide](./docs/DEVELOPMENT.md)
- [Task Planning](./docs/README.md)

---

## License

[MIT License](./LICENSE)

---

## Acknowledgments

- [Google Gemini](https://ai.google.dev/) - AI model provider
- [Vue.js](https://vuejs.org/) - Frontend framework
- [Naive UI](https://www.naiveui.com/) - UI component library
- [Gin](https://gin-gonic.com/) - Go web framework

---

<p align="center">
  Made with love by <a href="https://github.com/muxueliunian">muxueliunian</a>
</p>
