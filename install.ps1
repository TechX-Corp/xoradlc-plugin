# Cài lệnh xdlc (Xora DLC cho Claude Code và Kiro CLI) trên Windows.
#   powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.ps1 | iex"
# Cài xdlc.py + xdlc.cmd vào %LOCALAPPDATA%\Programs\xdlc và thêm thư mục đó vào PATH của user.
# Tuỳ chọn: XDLC_BIN_DIR (thư mục cài), XDLC_REF (nhánh / tag), XDLC_PLUGIN_DIR (lấy xdlc từ bản clone có sẵn),
#           XDLC_NO_PATH=1 (không sửa PATH, dùng cho test).
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

& {
    $Repo = 'TechX-Corp/xoradlc-plugin'
    $BinDir = if ($env:XDLC_BIN_DIR) { $env:XDLC_BIN_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\xdlc' }
    try { [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12 } catch {}

    # Python 3.9+: py launcher trước (ổn định qua các lần nâng Python), rồi python / python3.
    # Bỏ qua bản giả của Microsoft Store (WindowsApps): nó mở Store thay vì chạy Python.
    $python = $null
    foreach ($cand in @(@('py', '-3'), @('python'), @('python3'))) {
        $cmd = Get-Command $cand[0] -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
        if (-not $cmd -or $cmd.Source -like '*\WindowsApps\*') { continue }
        $extra = @($cand | Select-Object -Skip 1)
        & $cmd.Source @extra -c 'import sys; sys.exit(sys.version_info < (3, 9))' 2>$null
        if ($LASTEXITCODE -eq 0) { $python = @{ Exe = $cmd.Source; Args = $extra }; break }
    }
    if (-not $python) {
        throw "xdlc cần Python 3.9 trở lên. Cài bằng: winget install Python.Python.3.12  (hoặc python.org, chọn 'Add python.exe to PATH'), mở PowerShell mới rồi chạy lại lệnh cài."
    }

    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    $tmp = Join-Path $BinDir 'xdlc.download'
    try {
        if ($env:XDLC_PLUGIN_DIR) {
            Copy-Item (Join-Path $env:XDLC_PLUGIN_DIR 'xdlc') $tmp -Force
        } else {
            $ref = $env:XDLC_REF
            if (-not $ref) {  # bản phát hành mới nhất (tag vX.Y.Z), chưa có thì main
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
        if ($LASTEXITCODE -ne 0) { throw "xdlc: file tải về không hợp lệ." }
        Move-Item $tmp (Join-Path $BinDir 'xdlc.py') -Force
    } finally {
        Remove-Item $tmp -Force -ErrorAction SilentlyContinue
        Remove-Item (Join-Path $BinDir '__pycache__') -Recurse -Force -ErrorAction SilentlyContinue
    }

    # xdlc.cmd: gõ được lệnh xdlc trong PowerShell và cmd.
    $launch = (@("`"$($python.Exe)`"") + $python.Args) -join ' '
    Set-Content -Path (Join-Path $BinDir 'xdlc.cmd') -Encoding Ascii -Value "@echo off`r`n$launch `"%~dp0xdlc.py`" %*"

    $onPath = ($env:Path -split ';') -contains $BinDir
    if (-not $env:XDLC_NO_PATH) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if (-not (($userPath -split ';') -contains $BinDir)) {
            [Environment]::SetEnvironmentVariable('Path', ((@($userPath -split ';' | Where-Object { $_ }) + $BinDir) -join ';'), 'User')
        }
        if (-not $onPath) { $env:Path = "$env:Path;$BinDir" }  # dùng được ngay trong cửa sổ này
    }

    $version = & (Join-Path $BinDir 'xdlc.cmd') --version | Select-Object -First 1
    Write-Host "Đã cài $version vào $BinDir"
    Write-Host 'Tiếp theo: xdlc setup (cài cho Kiro hoặc Claude Code), rồi xdlc login.'
    if (-not $onPath -and -not $env:XDLC_NO_PATH) { Write-Host 'Cửa sổ terminal khác cần mở lại để thấy lệnh xdlc.' }
}
