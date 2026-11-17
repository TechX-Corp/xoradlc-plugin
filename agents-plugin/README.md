# Plugin theo vai (Claude Code / Kiro ↔ Xora DLC ↔ agent AAP)

Hiện chỉ dùng plugin `xora-dev` (vai Developer): skill `implement`, `changeset`, `upload-tech-doc`, sensor convention, và MCP server `xora-dev` (lệnh `xdlc mcp` trên máy, chuyển tới `<backend>/mcp/user`). Tool `xora_ask` gửi câu hỏi tới **backend Xora DLC** như chat trên web: backend xác định người hỏi qua token đăng nhập mà `xdlc` giữ, gắn vai và phiên MCP rồi chuyển cho **agent AAP** của vai đó. AAP tự chọn function MCP của Xora DLC cần gọi (việc của tôi, work order, tài liệu, open question) và tra KB, S3, kiểm quyền, ghi audit. Chi tiết: [dev/README.md](dev/README.md).

Các vai khác (`ba`, `techlead`, `qa`, `reviewer`, `delivery`) đã chuyển vào [archived/](archived/), không có trong marketplace và không được kiểm. Muốn dùng lại: chuyển thư mục về cạnh `dev/`, đổi `.mcp.json` và `mcp.json` sang `xdlc mcp` như `dev/` (các file này còn trỏ tới `xora_mcp.py` đã bỏ), thêm vào `kiro_files` của `xdlc`, thêm vào `marketplace.json` và `ROLES` trong `test_roles.py`.

## Cài đặt

Cần `python3` 3.9+ (macOS có sẵn). Dùng Claude Code thì cần thêm lệnh `claude`; Kiro thì `kiro-cli`.

**1. Cài lệnh `xdlc`** (một file, vào `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh
```

Nếu được nhắc, thêm `~/.local/bin` vào `PATH` rồi mở terminal mới.

**2. Cài plugin cho agent.** Muốn cài cho một project thì `cd` vào thư mục project trước:

```bash
cd ~/code/core-banking
xdlc setup                                  # hỏi agent, rồi hỏi scope
xdlc setup --agent kiro --scope local       # hoặc viết tắt, không hỏi
```

- **global**: mọi repo trên máy.
- **local**: project hiện tại; git không theo dõi file của `xdlc` (ghi vào `.git/info/exclude`, chỉ máy bạn). Nếu các file đó đã được git theo dõi từ trước (team đã commit), `xdlc` liệt kê và hỏi: **gỡ khỏi git** (`git rm --cached`: file vẫn còn trên máy, nhưng khi commit thì bị gỡ khỏi repo của cả team) hay **giữ như hiện tại** (thay đổi của `xdlc` sẽ hiện trong `git status`).

Setup local cho nhiều project được; setup lại đúng chỗ cũ là cài lại.

**3. Đăng nhập** (SSO hoặc tài khoản local, như trên web):

```bash
xdlc login
```

Mở phiên Kiro / Claude Code mới và hỏi, ví dụ *"Việc của tôi trên Xora DLC là gì?"*.

## Các lệnh

Gõ `xdlc` để xem danh sách này trên terminal:

```text
xdlc 0.2.0: Xora DLC cho Claude Code và Kiro CLI

Cách dùng: xdlc <lệnh> [tuỳ chọn]

Lệnh:
  setup      Cài plugin xora-dev cho Kiro CLI hoặc Claude Code
  remove     Gỡ một hay nhiều lần setup
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

`xdlc` giữ token trong `~/.xdlc/auth.json` (quyền `600`) và tự làm mới. Agent chạy `xdlc mcp` (MCP server cục bộ, stdio); chưa đăng nhập thì `xora_ask` trả lời "chạy `xdlc login`". Backend không giữ danh sách token: muốn chặn hẳn một người thì khoá tài khoản.

**Phát hành bản mới** (người bảo trì): tăng `VERSION` trong `xdlc` và `version` trong `.claude-plugin/marketplace.json` (phải khớp, test kiểm), merge vào `main`, tạo tag `vX.Y.Z`. Thử bản chưa phát hành: `XDLC_REF=<nhánh> xdlc update`, hoặc `XDLC_PLUGIN_DIR=<bản clone>` cho `install.sh` / `xdlc setup` / `xdlc update`.

## Sensor convention (`dev`)

Sau mỗi lần agent ghi / sửa code, sensor kiểm theo `convention.md` ở gốc repo và **chỉ cảnh báo**.

- Tạo `convention.md`: copy `~/.xdlc/plugin/agents-plugin/dev/convention.sample.md` về gốc repo, sửa cho khớp dự án.
- Luật `forbid` chỉ báo trên dòng đã đổi so với HEAD; `--full` để rà cả code cũ.
- Chạy tay / CI: `python3 ~/.xdlc/plugin/agents-plugin/dev/sensor/convention_sensor.py [file …] [--staged] [--full]` (`--staged` cho git pre-commit).
- Plugin dev làm được gì: [dev/README.md](dev/README.md).
- Bỏ qua một dòng: `convention-ignore: <ID> <lý do>`.

## Kiểm tra trước khi mở PR

`python3 agents-plugin/test_roles.py`

## Còn để ngỏ

- Câu hỏi qua Xora DLC tối đa 3.800 ký tự (giới hạn `question` của backend), nên `upload-tech-doc` chỉ lưu được tài liệu ngắn hoặc phải tách phần.
- Sensor chưa gửi cảnh báo về Xora DLC (xem [dev/README.md](dev/README.md)).
