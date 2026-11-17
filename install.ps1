# Install the xdlc command (Xora DLC for Claude Code and Kiro CLI) on Windows.
#   powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.ps1 | iex"
# Installs xdlc.py + xdlc.cmd to %LOCALAPPDATA%\Programs\xdlc and adds that folder to the user PATH.
# Options: XDLC_BIN_DIR (install folder), XDLC_REF (branch / tag), XDLC_PLUGIN_DIR (take xdlc from a local clone),
#          XDLC_NO_PATH=1 (leave PATH alone, for tests).
# ASCII only: Windows PowerShell 5.1 reads a .ps1 without BOM as ANSI, so non-ASCII text breaks the parser.
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

& {
    $Repo = 'TechX-Corp/xoradlc-plugin'
    $BinDir = if ($env:XDLC_BIN_DIR) { $env:XDLC_BIN_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\xdlc' }
    try { [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12 } catch {}

    # Python 3.9+: the py launcher first (stable across Python upgrades), then python / python3.
    # Skip the Microsoft Store stub (WindowsApps): it opens the Store instead of running Python.
    $python = $null
    foreach ($cand in @(@('py', '-3'), @('python'), @('python3'))) {
        $cmd = Get-Command $cand[0] -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
        if (-not $cmd -or $cmd.Source -like '*\WindowsApps\*') { continue }
        $extra = @($cand | Select-Object -Skip 1)
        & $cmd.Source @extra -c 'import sys; sys.exit(sys.version_info < (3, 9))' 2>$null
        if ($LASTEXITCODE -eq 0) { $python = @{ Exe = $cmd.Source; Args = $extra }; break }
    }
    if (-not $python) {
        throw "xdlc needs Python 3.9 or later. Install it: winget install Python.Python.3.12  (or python.org, tick 'Add python.exe to PATH'), open a new PowerShell and run the install command again."
    }

    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    $tmp = Join-Path $BinDir 'xdlc.download'
    try {
        if ($env:XDLC_PLUGIN_DIR) {
            Copy-Item (Join-Path $env:XDLC_PLUGIN_DIR 'xdlc') $tmp -Force
        } else {
            $ref = $env:XDLC_REF
            if (-not $ref) {  # latest release (tag vX.Y.Z), main when there is none
                $ref = 'main'
                try {
                    $tags = Invoke-RestMethod -UseBasicParsing "https://api.github.com/repos/$Repo/tags?per_page=100"
                    $best = $tags | Where-Object { $_.name -match '^v?\d+\.\d+\.\d+$' } |
                        Sort-Object { [version]($_.name.TrimStart('v')) } -Descending | Select-Object -First 1
                    if ($best) { $ref = $best.name }
                } catch {}
            }
            Invoke-WebRequest -UseBasicParsing "https://raw.githubusercontent.com/$Repo/$ref/xdlc" -OutFile $tmp
        }
        & $python.Exe @($python.Args) -m py_compile $tmp 2>$null
        if ($LASTEXITCODE -ne 0) { throw "xdlc: the downloaded file is not valid." }
        Move-Item $tmp (Join-Path $BinDir 'xdlc.py') -Force
    } finally {
        Remove-Item $tmp -Force -ErrorAction SilentlyContinue
        Remove-Item (Join-Path $BinDir '__pycache__') -Recurse -Force -ErrorAction SilentlyContinue
    }

    # xdlc.cmd: the xdlc command in PowerShell and cmd.
    $launch = (@("`"$($python.Exe)`"") + $python.Args) -join ' '
    Set-Content -Path (Join-Path $BinDir 'xdlc.cmd') -Encoding Ascii -Value "@echo off`r`n$launch `"%~dp0xdlc.py`" %*"

    $onPath = ($env:Path -split ';') -contains $BinDir
    if (-not $env:XDLC_NO_PATH) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if (-not (($userPath -split ';') -contains $BinDir)) {
            [Environment]::SetEnvironmentVariable('Path', ((@($userPath -split ';' | Where-Object { $_ }) + $BinDir) -join ';'), 'User')
        }
        if (-not $onPath) { $env:Path = "$env:Path;$BinDir" }  # usable in this window right away
    }

    $version = & (Join-Path $BinDir 'xdlc.cmd') --version | Select-Object -First 1
    Write-Host "Installed $version to $BinDir"
    Write-Host 'Next: xdlc setup (install for Kiro or Claude Code), then xdlc login.'
    if (-not $onPath -and -not $env:XDLC_NO_PATH) { Write-Host 'Other terminal windows must be reopened to see the xdlc command.' }
}
