# Xora DLC cho Claude Code và Kiro CLI

Plugin `xora-dev` đưa Xora DLC vào agent bạn đang dùng để code (Kiro CLI hoặc Claude Code), cài và quản lý bằng lệnh `xdlc`:

- **Hỏi Xora DLC ngay trong agent**: việc của tôi, trạng thái work order, BRD / thiết kế của bước, open question, convention và tiền lệ nội bộ; câu trả lời có trích dẫn.
- **Quy chuẩn làm việc có sẵn**: skill `implement` (code theo BRD / thiết kế đã duyệt), `changeset` (hồ sơ PR), `upload-tech-doc` (lưu tài liệu vào KB), cùng rule cho agent.
- **Kiểm convention**: sau mỗi lần agent sửa code, sensor kiểm theo `convention.md` của repo và cảnh báo; khi `git push`, cảnh báo được gửi về màn hình Governance của Xora DLC.

## Cần có trước khi cài

| | Yêu cầu |
|---|---|
| Hệ điều hành | macOS, Linux, Windows 10/11 (hoặc WSL) |
| Python | 3.9 trở lên (`python3` trên macOS / Linux, `py` hoặc `python` trên Windows) |
| git | Bản bất kỳ còn hỗ trợ |
| Agent | **Kiro CLI** (`kiro-cli`) hoặc **Claude Code** (`claude`). Claude Code trên Windows cần Git for Windows. Kiro CLI trên Windows gốc chưa được kiểm: nếu chỉ chạy được trong WSL thì cài `xdlc` trong WSL |
| Tài khoản | Tài khoản Xora DLC (SSO hoặc tài khoản local, như trên web) có vai được hỏi agent, vd. DEV |
| Terminal | macOS / Linux: `curl`. Windows: PowerShell (có sẵn); nên dùng Windows Terminal |

## Cài đặt

Cài lệnh `xdlc` (một file). Gặp lỗi: xem [Lỗi thường gặp khi cài](#lỗi-thường-gặp-khi-cài).

**macOS / Linux / WSL** (cài vào `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh
```

**Windows** (PowerShell hoặc cmd; cài vào `%LOCALAPPDATA%\Programs\xdlc`, tự thêm vào `PATH`):

```powershell
powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.ps1 | iex"
```

Kiểm tra (mở terminal mới nếu vừa cài):

```bash
xdlc --version
```

## Bắt đầu nhanh

**1. Cài plugin cho agent.** Chọn một trong hai:

Cho project đang làm (cd vào thư mục project trước):

```bash
cd ~/your-project
xdlc setup --agent kiro --scope local
```

Cho mọi repo trên máy:

```bash
xdlc setup --agent kiro --scope global
```

| Flag | Giá trị | Nghĩa |
|---|---|---|
| `-a`, `--agent` | `kiro` | Kiro CLI |
| | `claude` | Claude Code |
| `-s`, `--scope` | `local` | Chỉ project hiện tại; git không theo dõi file của `xdlc`. Cài thêm hook `git push` để gửi cảnh báo convention về Xora DLC |
| | `global` | Mọi repo trên máy |

Bỏ flag nào thì `xdlc` hỏi trên terminal (↑/↓ chọn, Enter xác nhận). Nếu file của plugin đã được team commit vào repo, `xdlc` hỏi gỡ khỏi git hay giữ nguyên.

**2. Đăng nhập Xora DLC:**

```bash
xdlc login
```

**3. Mở phiên agent mới trong project và hỏi**, ví dụ:

> Việc của tôi trên Xora DLC là gì?

**4. (Tuỳ chọn) Gửi cảnh báo convention về Xora DLC khi `git push`:** map nhánh hiện tại với bước bạn đang làm:

```bash
xdlc link
```

Sensor kiểm convention ngay sau khi cài, không cần thao tác thêm: repo có `convention.md` (ở gốc hoặc `docs/`) thì theo luật của repo, không có thì theo bản mặc định đi kèm plugin (hiện chưa có luật nào, sẽ bổ sung sau).

## Các lệnh

Gõ `xdlc` để xem danh sách này:

```text
xdlc 0.2.0: Xora DLC cho Claude Code và Kiro CLI

Cách dùng: xdlc <lệnh> [tuỳ chọn]

Lệnh:
  setup      Cài plugin xora-dev cho Kiro CLI hoặc Claude Code
  remove     Gỡ một hay nhiều lần setup
  link       Chọn work order cho nhánh hiện tại (sensor gửi cảnh báo khi git push)
  login      Đăng nhập Xora DLC
  logout     Đăng xuất Xora DLC trên máy này
  status     Backend, tài khoản và vai đang đăng nhập, nơi đã setup
  use        Chọn backend: bản deploy (prod) hoặc local
  update     Cập nhật xdlc và plugin lên bản mới nhất
  uninstall  Gỡ toàn bộ Xora DLC khỏi máy

Tuỳ chọn:
  -v, --version  Version đang cài và bản mới nhất
  -h, --help     Danh sách lệnh này

Viết tắt (bỏ qua câu hỏi trên terminal):
  xdlc setup -a, --agent kiro|claude  -s, --scope global|local
  xdlc remove -a, --agent kiro|claude
  xdlc link [WO-…] [-t, --step BƯỚC] [-b, --branch NHÁNH] [-l, --list]
  xdlc use prod | xdlc use local [-u, --url URL]

Scope: global = mọi repo trên máy; local = project hiện tại, git không theo dõi file của xdlc.
Xem chi tiết một lệnh: xdlc <lệnh> --help
```

## Lưu ý

- Gặp lỗi (agent báo `forbidden`, không thấy MCP `xora-dev`…): chạy `xdlc status` để xem đang dùng tài khoản, vai nào và đã cài ở đâu.
- Đăng nhập nhầm tài khoản: đăng xuất trên web Xora DLC, rồi chạy `xdlc logout` và `xdlc login`.
- Sau khi chạy `xdlc use`: mở phiên agent mới.
- Bấm ↑/↓ mà menu không đổi: đặt `XDLC_PLAIN=1` để chọn bằng cách gõ số.

  ```bash
  XDLC_PLAIN=1 xdlc setup
  ```

  ```powershell
  $env:XDLC_PLAIN = "1"; xdlc setup
  ```

## Lỗi thường gặp khi cài

| Gặp | Xử lý |
|---|---|
| macOS / Linux: `xdlc: command not found` | Thêm `~/.local/bin` vào `PATH`: `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc` (bash: `~/.bashrc`), rồi mở terminal mới |
| Windows: `'xdlc' is not recognized` | Mở cửa sổ terminal mới. Vẫn lỗi: kiểm `PATH` của user có `%LOCALAPPDATA%\Programs\xdlc` |
| `cần python3` / `xdlc needs Python 3.9 or later` | Cài Python 3.9+ (Windows: chọn *Add python.exe to PATH*), mở terminal mới rồi chạy lại lệnh cài |
| Windows: `Python was not found; run without arguments to install from the Microsoft Store` | Đó là bản giả của Microsoft Store: cài Python thật, hoặc tắt *App execution aliases* cho `python.exe` trong Settings |
| Windows: PowerShell báo script bị chặn (ExecutionPolicy, AppLocker) | Dùng đúng lệnh cài ở trên (có `-ExecutionPolicy ByPass`). Máy công ty vẫn chặn thì nhờ IT |
| `không tải được xdlc` | Kiểm mạng tới `raw.githubusercontent.com` (proxy công ty có thể chặn) |
| Menu không đổi khi bấm ↑/↓ | Console cũ không hỗ trợ: dùng Windows Terminal, hoặc chọn bằng số (xem [Lưu ý](#lưu-ý)) |
