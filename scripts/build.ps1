# MuxueTools Build Script
# This script builds MuxueTools for different platforms and targets.

param(
    [Parameter(Position = 0)]
    [ValidateSet("server", "desktop", "desktop-x86", "all", "clean")]
    [string]$Target = "all"
)

# Configuration
$ErrorActionPreference = "Stop"
$BinDir = Join-Path $PSScriptRoot "..\bin"
$ProjectRoot = Join-Path $PSScriptRoot ".."

# Version info (can be overridden by CI)
$Version = if ($env:VERSION) { $env:VERSION } else { "dev" }
$BuildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$GitCommit = try { git rev-parse --short HEAD 2>$null } catch { "unknown" }

# LDFlags for version injection
$LDFlags = "-X 'main.Version=$Version' -X 'main.BuildTime=$BuildTime' -X 'main.GitCommit=$GitCommit'"

function Write-Header {
    param([string]$Message)
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host " $Message" -ForegroundColor Cyan
    Write-Host "========================================`n" -ForegroundColor Cyan
}

function Ensure-BinDir {
    if (-not (Test-Path $BinDir)) {
        New-Item -ItemType Directory -Path $BinDir | Out-Null
        Write-Host "Created bin directory: $BinDir" -ForegroundColor Green
    }
}

function Build-Frontend {
    Write-Header "Building Frontend (Vite + Vue)"
    
    $WebDir = Join-Path $ProjectRoot "web"
    
    if (-not (Test-Path $WebDir)) {
        Write-Host "Error: Web directory not found at $WebDir" -ForegroundColor Red
        return $false
    }
    
    Push-Location $WebDir
    try {
        # Check if node_modules exists
        if (-not (Test-Path "node_modules")) {
            Write-Host "Installing dependencies..." -ForegroundColor Yellow
            npm ci
            if ($LASTEXITCODE -ne 0) {
                Write-Host "npm ci failed!" -ForegroundColor Red
                return $false
            }
        }
        
        # Build frontend
        Write-Host "Building frontend..." -ForegroundColor Cyan
        npm run build
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Frontend build failed!" -ForegroundColor Red
            return $false
        }
        
        Write-Host "Frontend built successfully: web/dist" -ForegroundColor Green
        return $true
    }
    finally {
        Pop-Location
    }
}

function Build-Server {
    Write-Header "Building Server (Pure Go)"
    
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    
    $OutputPath = Join-Path $BinDir "muxueTools-server.exe"
    
    Push-Location $ProjectRoot
    try {
        go build -ldflags $LDFlags -o $OutputPath ./cmd/server
        Write-Host "Built: $OutputPath" -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

function Ensure-IconResource {
    param(
        [string]$Arch = "amd64"
    )
    
    $RcPath = Join-Path $ProjectRoot "cmd\desktop\app.rc"
    $IconPath = Join-Path $ProjectRoot "assets\icon.ico"
    $SysoPath = Join-Path $ProjectRoot "cmd\desktop\rsrc_windows_$Arch.syso"
    
    if (-not (Test-Path $IconPath)) {
        Write-Host "Warning: Icon file not found at $IconPath" -ForegroundColor Yellow
        return
    }
    
    # 检测 windres (MinGW)
    $windresPath = $null
    $mingwPaths = @(
        "C:\msys64\mingw64\bin\windres.exe",
        "C:\msys64\ucrt64\bin\windres.exe",
        "C:\mingw64\bin\windres.exe"
    )
    
    foreach ($path in $mingwPaths) {
        if (Test-Path $path) {
            $windresPath = $path
            break
        }
    }
    
    # 优先使用 windres（如果可用且 app.rc 存在）
    if ($windresPath -and (Test-Path $RcPath)) {
        Write-Host "Using windres to compile icon resource..." -ForegroundColor Cyan
        Push-Location (Join-Path $ProjectRoot "cmd\desktop")
        try {
            & $windresPath -i app.rc -o $SysoPath
            if ($LASTEXITCODE -eq 0) {
                Write-Host "Generated (windres): $SysoPath" -ForegroundColor Green
                return
            }
            else {
                Write-Host "windres failed, falling back to rsrc..." -ForegroundColor Yellow
            }
        }
        finally { Pop-Location }
    }
    
    # 回退到 rsrc
    $rsrcPath = (Get-Command rsrc -ErrorAction SilentlyContinue).Source
    if (-not $rsrcPath) {
        Write-Host "Installing rsrc tool..." -ForegroundColor Yellow
        go install github.com/akavel/rsrc@latest
    }
    
    Write-Host "Generating Windows resource file for icon ($Arch) using rsrc..." -ForegroundColor Cyan
    Push-Location $ProjectRoot
    try {
        rsrc -arch $Arch -ico $IconPath -o $SysoPath
        Write-Host "Generated (rsrc): $SysoPath" -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

function Build-Desktop {
    param(
        [string]$Arch = "amd64",
        [string]$Suffix = ""
    )
    
    Write-Header "Building Desktop (CGO + WebView) - $Arch"
    
    # Generate icon resource file
    Ensure-IconResource -Arch $Arch
    
    # Desktop build requires CGO for WebView
    $env:CGO_ENABLED = "1"
    $env:GOOS = "windows"
    $env:GOARCH = $Arch
    
    $OutputName = if ($Suffix) { "muxueTools$Suffix.exe" } else { "muxueTools.exe" }
    $OutputPath = Join-Path $BinDir $OutputName
    
    # -H windowsgui hides the console window
    $DesktopLDFlags = "$LDFlags -H windowsgui"
    
    Push-Location $ProjectRoot
    try {
        go build -ldflags $DesktopLDFlags -o $OutputPath ./cmd/desktop
        Write-Host "Built: $OutputPath" -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

function Clean-Build {
    Write-Header "Cleaning build artifacts"
    
    if (Test-Path $BinDir) {
        Remove-Item -Recurse -Force $BinDir
        Write-Host "Removed: $BinDir" -ForegroundColor Yellow
    }
    
    Write-Host "Clean complete" -ForegroundColor Green
}

# Main execution
Write-Host "MuxueTools Build Script" -ForegroundColor Magenta
Write-Host "Version: $Version | Commit: $GitCommit" -ForegroundColor DarkGray

Ensure-BinDir

switch ($Target) {
    "server" {
        Build-Server
    }
    "desktop" {
        # Build frontend first, then desktop
        $frontendOk = Build-Frontend
        if ($frontendOk) {
            Build-Desktop -Arch "amd64"
        }
    }
    "desktop-x86" {
        # Build frontend first, then desktop x86
        $frontendOk = Build-Frontend
        if ($frontendOk) {
            Build-Desktop -Arch "386" -Suffix "-x86"
        }
    }
    "all" {
        # Build frontend first
        $frontendOk = Build-Frontend
        if ($frontendOk) {
            Build-Server
            Build-Desktop -Arch "amd64"
        }
    }
    "clean" {
        Clean-Build
    }
}

Write-Host "`nBuild complete!" -ForegroundColor Green

