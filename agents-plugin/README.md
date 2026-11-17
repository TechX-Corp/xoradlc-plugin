# Plugin theo vai (Claude Code / Kiro ↔ Xora DLC ↔ agent AAP)

Hiện chỉ dùng plugin `xora-dev` (vai Developer): skill `implement`, `changeset`, `upload-tech-doc`, sensor convention, và MCP server `xora-dev` (lệnh `xdlc mcp` trên máy, chuyển tới `<backend>/mcp/user`). Tool `xora_ask` gửi câu hỏi tới **backend Xora DLC** như chat trên web: backend xác định người hỏi qua token đăng nhập mà `xdlc` giữ, gắn vai và phiên MCP rồi chuyển cho **agent AAP** của vai đó. AAP tự chọn function MCP của Xora DLC cần gọi (việc của tôi, work order, tài liệu, open question) và tra KB, S3, kiểm quyền, ghi audit. Chi tiết: [dev/README.md](dev/README.md).

Các vai khác (`ba`, `techlead`, `qa`, `reviewer`, `delivery`) đã chuyển vào [archived/](archived/), không có trong marketplace và không được kiểm. Muốn dùng lại: chuyển thư mục về cạnh `dev/`, đổi `.mcp.json` và `mcp.json` sang `xdlc mcp` như `dev/` (các file này còn trỏ tới `xora_mcp.py` đã bỏ), thêm vào `kiro_files` của `xdlc`, thêm vào `marketplace.json` và `ROLES` trong `test_roles.py`.

## Cài đặt

Cần Python 3.9+ và git. Dùng Claude Code thì cần lệnh `claude`; Kiro thì `kiro-cli`. Ba bước: cài lệnh `xdlc`, cài plugin cho agent, đăng nhập.

### macOS / Linux / WSL

```bash
# 1. Cài xdlc vào ~/.local/bin (macOS có sẵn python3)
curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh

# 2. Cài plugin: cd vào project nếu chọn scope local
cd ~/code/core-banking
xdlc setup                              # hỏi agent, rồi scope
xdlc setup --agent kiro --scope local   # hoặc viết tắt

# 3. Đăng nhập (mở trình duyệt; SSO hoặc tài khoản local như trên web)
xdlc login
```

Nếu được nhắc, thêm `~/.local/bin` vào `PATH` (vd. `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc`) rồi mở terminal mới. Dùng WSL thì Kiro / Claude Code cũng phải chạy trong WSL.

### Windows

Chạy trong **PowerShell** hoặc **cmd** (nên dùng Windows Terminal: menu phím mũi tên hiển thị đúng).

**Bước 0: cài Python và git** (bỏ qua nếu máy đã có; kiểm bằng `py --version` và `git --version`):

```powershell
winget install Python.Python.3.12
winget install Git.Git
```

Cài từ python.org thì chọn *Add python.exe to PATH*. Cài xong, **mở cửa sổ PowerShell mới**.

**Bước 1: cài `xdlc`** vào `%LOCALAPPDATA%\Programs\xdlc` (tự thêm vào `PATH` của user):

```powershell
powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.ps1 | iex"
```

Kiểm: `xdlc --version`. Cửa sổ terminal mở trước khi cài phải mở lại mới thấy lệnh `xdlc`.

**Bước 2, 3: cài plugin và đăng nhập**, giống hệt macOS:

```powershell
cd C:\code\core-banking
xdlc setup --agent claude --scope local
xdlc login
```

Trên Windows, `xdlc` lưu dữ liệu ở `%USERPROFILE%\.xdlc` (token mã hoá bằng DPAPI, chỉ tài khoản Windows của bạn giải mã được). Lệnh `xdlc` là `xdlc.cmd` gọi `py -3 xdlc.py`.

| Gặp | Xử lý |
|---|---|
| `xdlc is not recognized` / `'xdlc' không được nhận dạng` | Mở cửa sổ terminal mới. Vẫn lỗi: kiểm `PATH` của user có `%LOCALAPPDATA%\Programs\xdlc` |
| `Python was not found; run without arguments to install from the Microsoft Store` | Đó là bản giả của Store. Cài Python ở Bước 0, hoặc tắt *App execution aliases* cho `python.exe` trong Settings |
| `xdlc cần Python 3.9 trở lên` | Cài Python ở Bước 0, mở PowerShell mới, chạy lại lệnh cài |
| PowerShell báo script bị chặn (ExecutionPolicy, AppLocker) | Dùng đúng lệnh có `-ExecutionPolicy ByPass` ở trên; máy công ty vẫn chặn thì nhờ IT |
| Menu không đổi khi bấm ↑/↓ (console cũ) | Dùng Windows Terminal, hoặc đặt `set XDLC_PLAIN=1` để gõ số |

Claude Code trên Windows cần Git for Windows (Bước 0). Kiro CLI trên Windows gốc chưa được kiểm: nếu Kiro chỉ chạy trong WSL thì cài `xdlc` trong WSL theo mục trên.

## Các lệnh

Gõ `xdlc` để xem danh sách này trên terminal:

```text
xdlc 0.2.0: Xora DLC cho Claude Code và Kiro CLI

Cách dùng: xdlc <lệnh> [tuỳ chọn]

Lệnh:
  setup      Cài plugin xora-dev cho Kiro CLI hoặc Claude Code
  remove     Gỡ một hay nhiều lần setup
  link       Chọn work order cho nhánh hiện tại (sensor gửi cảnh báo khi git push)
  login      Đăng nhập Xora DLC (mở trình duyệt)
  logout     Đăng xuất Xora DLC trên máy này
  status     Backend, tài khoản và vai đang đăng nhập, nơi đã setup
  use        Chọn backend: bản deploy (prod) hoặc local
  update     Cập nhật xdlc và plugin lên bản mới nhất
  version    Version đang cài và bản mới nhất
  uninstall  Gỡ toàn bộ Xora DLC khỏi máy

Viết tắt (bỏ qua câu hỏi trên terminal):
  xdlc setup --agent kiro|claude --scope global|local
  xdlc remove --agent kiro|claude
  xdlc link [WO-… [--step …]] [--branch …] [--list]
  xdlc use prod | xdlc use local [--url URL]

Scope: global = mọi repo trên máy; local = project hiện tại, git không theo dõi file của xdlc.
Xem chi tiết một lệnh: xdlc <lệnh> --help
```

Thêm về từng lệnh:

- Câu hỏi chọn trên terminal: ↑/↓ di chuyển, Enter chọn, Esc huỷ; chọn nhiều thì Space đánh dấu, `a` chọn tất cả. Không phải terminal tương tác (pipe, CI) hoặc `XDLC_PLAIN=1` thì gõ số. Câu hỏi có / không thì gõ `y`, để trống là không.
- `xdlc remove`: liệt kê các lần setup (đánh dấu thư mục hiện tại), chọn một hay nhiều mục (không có terminal: gõ `1,3`, `1-2`, `all`), xem lại danh sách sẽ gỡ rồi xác nhận một lần. Chỉ xoá đúng file `xdlc` đã cài (ghi trong `~/.xdlc/installs.json`); file bạn đã sửa thì hỏi riêng; trong `mcp.json` chỉ xoá khối `xora-dev`.
- `xdlc login`: trình duyệt đang đăng nhập web Xora DLC thì dùng luôn tài khoản đó. `xdlc` in ra tài khoản và vai để bạn kiểm; sai tài khoản thì đăng xuất web, `xdlc logout`, rồi `xdlc login`.
- `xdlc use local`: dùng backend chạy trên máy (mặc định `http://localhost:8000`, đổi bằng `--url`). Token tách riêng theo backend. Đổi xong, mở phiên agent mới.
- `xdlc update`: tải bản phát hành mới nhất, thay `xdlc` và plugin, cài lại vào mọi chỗ đã setup (giữ lựa chọn cũ, không hỏi lại).
- `xdlc uninstall`: gỡ mọi lần setup, đăng xuất, xoá `~/.xdlc` và chính lệnh `xdlc`.

`xdlc` giữ token trong `~/.xdlc/auth.json` (macOS / Linux: quyền `600`; Windows: mã hoá DPAPI, chỉ tài khoản Windows đó giải mã được) và tự làm mới. Agent chạy `xdlc mcp` (MCP server cục bộ, stdio); chưa đăng nhập thì `xora_ask` trả lời "chạy `xdlc login`". Backend không giữ danh sách token: muốn chặn hẳn một người thì khoá tài khoản.

**Phát hành bản mới** (người bảo trì): tăng `VERSION` trong `xdlc` và `version` trong `.claude-plugin/marketplace.json` (phải khớp, test kiểm), merge vào `main`, tạo tag `vX.Y.Z`. Thử bản chưa phát hành: `XDLC_REF=<nhánh> xdlc update`, hoặc `XDLC_PLUGIN_DIR=<bản clone>` cho `install.sh` / `xdlc setup` / `xdlc update`.

## Sensor convention (`dev`)

Sau mỗi lần agent ghi / sửa code, sensor kiểm các **dòng vừa đổi** theo `convention.md` ở gốc repo và **chỉ cảnh báo**, không chặn. Luồng chi tiết: [dev/README.md](dev/README.md#sensor-chạy-thế-nào).

- Tạo `convention.md`: copy `convention.sample.md` của plugin về gốc repo rồi sửa cho khớp dự án. Chưa có file này thì sensor không kiểm gì.
  - macOS / Linux: `cp ~/.xdlc/plugin/agents-plugin/dev/convention.sample.md convention.md`
  - Windows: `Copy-Item $HOME\.xdlc\plugin\agents-plugin\dev\convention.sample.md .\convention.md`
- Luật `forbid` chỉ báo trên dòng đã đổi so với HEAD; `--full` để rà cả code cũ. Mỗi luật có thể thêm `"severity": "red"` (mặc định `yellow`): mức của alert trên Xora DLC.
- Chạy tay / pre-commit: `xdlc sensor [file …] [--staged] [--full]` (`--staged` cho git pre-commit). CI không có `xdlc`: `python3 agents-plugin/dev/sensor/convention_sensor.py …` từ bản clone.
- Bỏ qua một dòng: `convention-ignore: <ID> <lý do>`.

**Gửi cảnh báo về Xora DLC khi `git push`** (màn hình Governance, vai Dev):

1. Hook git `pre-push` của `xdlc`: setup **local** cài sẵn; setup **global** thì chạy `xdlc link` trong repo (lệnh này cài hook cho repo đó). Repo dùng `core.hooksPath` (husky, …) hoặc đã có hook `pre-push` riêng: `xdlc` không ghi đè, chỉ in dòng lệnh để bạn thêm vào hook của repo.
2. Map nhánh với work order: `xdlc link` (menu các bước bạn đang làm, lấy từ Việc của tôi), hoặc `xdlc link WO-2026-001 [--step impl]`. Lần push đầu của một nhánh chưa map, hook hỏi ngay trên terminal; push từ nơi không có terminal (git GUI) thì bỏ qua lần gửi đó. Map theo URL `origin` + tên nhánh, dùng chung cho mọi bản clone; xem bằng `xdlc status` hoặc `xdlc link --list`.
3. Khi push, sensor kiểm các file thay đổi so với commit đã có trên remote, in cảnh báo, rồi gửi về Xora DLC bằng token của `xdlc login`. Không bao giờ chặn push: chưa đăng nhập, mất mạng, không phải Dev của work order… đều chỉ in lý do. `git push --no-verify` bỏ qua hook.

## Kiểm tra trước khi mở PR

`python3 agents-plugin/test_roles.py`

## Còn để ngỏ

- Câu hỏi qua Xora DLC tối đa 3.800 ký tự (giới hạn `question` của backend), nên `upload-tech-doc` chỉ lưu được tài liệu ngắn hoặc phải tách phần.
- Mỗi lần push gửi lại mọi cảnh báo còn lại: backend chưa gộp alert trùng giữa các lần gửi, cũng chưa tự đóng alert khi lỗi đã sửa.
- Pre-push kiểm bản trong thư mục làm việc của các file thay đổi (thay đổi chưa commit cũng tính), không phải đúng bản trong commit được push.
