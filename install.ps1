# Install the xdlc binary (Xora DLC for Claude Code and Kiro CLI) on Windows.
#   powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.ps1 | iex"
# Installs xdlc.exe to %LOCALAPPDATA%\Programs\xdlc and adds that folder to the user PATH.
# Options: XDLC_BIN_DIR, XDLC_REF (tag vX.Y.Z), XDLC_PLUGIN_DIR (local clone with built xdlc.exe),
#          XDLC_NO_PATH=1 (leave PATH alone, for tests).
# ASCII only: Windows PowerShell 5.1 reads a .ps1 without BOM as ANSI.
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

& {
    $Repo = 'TechX-Corp/xoradlc-plugin'
    $BinDir = if ($env:XDLC_BIN_DIR) { $env:XDLC_BIN_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\xdlc' }
    try { [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12 } catch {}

    $arch = $env:PROCESSOR_ARCHITECTURE
    if ($arch -eq 'AMD64' -or $arch -eq 'x86') { $goarch = 'amd64' }
    elseif ($arch -eq 'ARM64') { $goarch = 'arm64' }
    else { throw "xdlc: unsupported architecture: $arch" }
    $Asset = "xdlc-windows-$goarch.exe"

    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    $dest = Join-Path $BinDir 'xdlc.exe'
    $tmp = Join-Path $BinDir 'xdlc.download'
    try {
        if ($env:XDLC_PLUGIN_DIR) {
            $src = Join-Path $env:XDLC_PLUGIN_DIR 'xdlc.exe'
            if (-not (Test-Path $src)) { $src = Join-Path $env:XDLC_PLUGIN_DIR 'xdlc' }
            if (-not (Test-Path $src)) { throw "xdlc: XDLC_PLUGIN_DIR has no xdlc.exe (go build -o xdlc.exe ./cmd/xdlc)" }
            Copy-Item $src $tmp -Force
        } else {
            $ref = $env:XDLC_REF
            if (-not $ref) {
                try {
                    $rel = Invoke-RestMethod -UseBasicParsing "https://api.github.com/repos/$Repo/releases/latest"
                    $ref = $rel.tag_name
                } catch {}
            }
            if (-not $ref) { throw "xdlc: no GitHub release yet (tag vX.Y.Z). Build from source or set XDLC_PLUGIN_DIR." }
            $url = "https://github.com/$Repo/releases/download/$ref/$Asset"
            Invoke-WebRequest -UseBasicParsing $url -OutFile $tmp
        }
        Move-Item $tmp $dest -Force
    } finally {
        Remove-Item $tmp -Force -ErrorAction SilentlyContinue
    }

    $onPath = ($env:Path -split ';') -contains $BinDir
    if (-not $env:XDLC_NO_PATH) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if (-not (($userPath -split ';') -contains $BinDir)) {
            [Environment]::SetEnvironmentVariable('Path', ((@($userPath -split ';' | Where-Object { $_ }) + $BinDir) -join ';'), 'User')
        }
        if (-not $onPath) { $env:Path = "$env:Path;$BinDir" }
    }

    $version = & $dest --version | Select-Object -First 1
    Write-Host "Installed $version to $BinDir"
    Write-Host 'Next: xdlc setup (install for Kiro or Claude Code), then xdlc login.'
    if (-not $onPath -and -not $env:XDLC_NO_PATH) { Write-Host 'Other terminal windows must be reopened to see the xdlc command.' }
}
