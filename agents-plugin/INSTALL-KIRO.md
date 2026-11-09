# Cài plugin theo vai vào Kiro

Hướng dẫn cài một hoặc nhiều vai (`ba`, `techlead`, `dev`, `qa`, `reviewer`, `delivery`) vào Kiro. Mỗi vai gồm quy tắc làm việc (steering), skill, và MCP server `xora-<vai>` nối với agent AAP. Cài cho Claude Code: xem [README.md](README.md#cài-cho-claude-code).

## Cần có

- Kiro IDE có hỗ trợ Agent Skills.
- `python3` ≥ 3.9 trong `PATH`. MCP server và sensor chỉ dùng thư viện chuẩn, không cần `pip install`.
- macOS hoặc Linux. Windows: dùng WSL, hoặc cài tay theo [Cách C](#cách-c--cài-tay).
- Thông tin kết nối AAP: URL nền tảng, API key (`ak_…`), email công việc của bạn.

## Mỗi vai cài gì

| Vai | Skill | Steering thêm | MCP server |
|---|---|---|---|
| `ba` | `clarify-requirements`, `write-brd`, `write-user-stories`, `draw-process-flow`, `upload-kb` | `brd-documents.md` | `xora-ba` |
| `techlead` | `design`, `plan` | — | `xora-techlead` |
| `dev` | `implement`, `changeset`, `upload-tech-doc` | `changeset-documents.md` | `xora-dev` |
| `qa` | `test-plan`, `test-cases`, `test-results` | — | `xora-qa` |
| `reviewer` | `gather-evidence`, `decision` | — | `xora-reviewer` |
| `delivery` | `handover` | — | `xora-delivery` |

Vai nào cũng có thêm steering `xora-<vai>.md` (chính là `AGENTS.md` của vai, luôn được nạp). Riêng `dev` có thêm hook **Convention sensor**.

## Bước 1 — Lưu thông tin kết nối AAP

Làm **một lần** trên mỗi máy, dùng chung cho mọi vai. Kiro mở từ giao diện không thừa hưởng biến môi trường của terminal, nên thông tin được để trong một file mà MCP server tự đọc:

```bash
mkdir -p ~/.xora
cat > ~/.xora/env <<'ENV'
export XORA_BASE_URL=https://<platform-host>/api/v1
export XORA_API_KEY=ak_...
export XORA_USER_ID=<email công việc>
# export XORA_AGENT=<agent id>   # tuỳ chọn, mặc định: developer
ENV
chmod 600 ~/.xora/env
```

**Xora hiện tại (đã kiểm chứng):** API `/invoke` chưa có host công khai, nên dùng API `runs` của Xora — các giá trị lấy từ `agent-connection/xora.config.json` và `agent-connection/api-key.md` của dự án:

```bash
cat > ~/.xora/env <<'ENV'
export XORA_API=runs
export XORA_BASE_URL=https://d1xbzyews27i3w.cloudfront.net/v1
export XORA_AGENT=agent_mw3at9y00
export XORA_API_KEY=xap_...
ENV
chmod 600 ~/.xora/env
```

Với `XORA_API=runs`: không cần `XORA_USER_ID`; mỗi lượt hỏi độc lập (không giữ `session_id`); vai được gửi kèm ở đầu câu hỏi; `xora_get_run` dùng được.

- Không commit file này, không dán API key vào `mcp.json` hay vào repo dự án.
- Đổi key hoặc URL: sửa file này, rồi bấm **Retry** server trong Kiro.

## Bước 2 — Cài vai

Chọn **một** cách.

### Cách A — Script (khuyên dùng)

Chạy từ thư mục clone `xoradlc-plugin`, đích là **repo dự án** bạn mở trong Kiro:

```bash
sh agents-plugin/install-kiro.sh dev ~/code/core-banking
sh agents-plugin/install-kiro.sh ba  ~/code/core-banking     # cài nhiều vai vào cùng repo được
```

Script làm 3 việc:

| Việc | Đích |
|---|---|
| Chép vai thành bản cài cố định, chép file thật thay cho symlink `mcp/xora_mcp.py` | `~/.agent-plugins/xora-<vai>/` |
| Chép skill, `AGENTS.md` → `xora-<vai>.md`, steering của vai, hook (`dev`) | `<repo>/.kiro/skills/`, `<repo>/.kiro/steering/`, `<repo>/.kiro/hooks/` |
| Gộp server `xora-<vai>` vào cấu hình MCP, giữ nguyên server khác đã có | `<repo>/.kiro/settings/mcp.json` |

- **Cập nhật:** `git pull` repo `xoradlc-plugin` rồi chạy lại đúng lệnh trên. Chạy lại an toàn, không nhân đôi.
- **Cài cho mọi workspace:** thay đường dẫn repo bằng `--global` để cài vào `~/.kiro`. Hook của Kiro chỉ có ở workspace, nên muốn có sensor convention thì cài `dev` theo repo.
- Script **từ chối** cài vào chính thư mục `agents-plugin/`. Đích phải là repo dự án.
- Có nên commit `<repo>/.kiro/` không: tuỳ đội. Commit thì cả đội dùng chung; đừng commit nếu mỗi người tự cài.

### Cách B — Kiro Power

1. Chạy Cách A một lần (với `--global` hoặc bất kỳ repo nào) để có thư mục `~/.agent-plugins/xora-<vai>/`.
2. Trong Kiro: panel **Powers** → **Add power from local path** (hoặc *Import*) → chọn `~/.agent-plugins/xora-<vai>`.
3. Power tự kích hoạt khi bạn nhắc tới các từ khoá trong `POWER.md` của vai.

> Tên menu có thể khác giữa các bản Kiro. Không thấy mục thêm power cục bộ thì dùng Cách A.

### Cách C — Cài tay

1. Chép vai, **dereference symlink**: `cp -RL agents-plugin/<vai> ~/.agent-plugins/xora-<vai>`.
2. Chép vào repo dự án:

   | Từ `agents-plugin/<vai>/` | Tới `<repo>/.kiro/` |
   |---|---|
   | `skills/*` | `skills/` (mỗi skill là một thư mục chứa `SKILL.md`) |
   | `AGENTS.md` | `steering/xora-<vai>.md` |
   | `steering/*.md` (nếu có) | `steering/` |
   | `kiro-hooks/*.kiro.hook` (chỉ `dev`) | `hooks/` |
   | khối `mcpServers` trong `mcp.json` | gộp vào `settings/mcp.json` |

## Bước 3 — Kiểm tra

Mở lại Kiro, hoặc bấm **MCP Servers → Retry**, rồi kiểm:

1. **MCP:** panel MCP có `xora-<vai>` màu xanh với 2 tool `xora_ask`, `xora_get_run`.
2. **Steering:** panel **Agent Steering** có `xora-<vai>.md` (và `brd-documents.md` / `changeset-documents.md` với `ba` / `dev`).
3. **Skill:** gõ `/` trong chat, thấy skill của vai theo bảng trên.
4. **Kết nối AAP:** hỏi *"Hỏi AAP: bạn hỗ trợ những việc gì?"* → câu trả lời kết thúc bằng `[session_id: … | evidence: …]`.
5. **Sensor (`dev`):** panel **Agent Hooks** có *Convention sensor*. Chép `~/.agent-plugins/xora-dev/convention.sample.md` về gốc repo thành `convention.md`, nhờ agent viết một hàm có `System.out.println` hoặc `print()` → agent chạy sensor và báo `⚠ Convention: … [LOG-…]`.

Kiểm ngoài Kiro khi cần khoanh vùng lỗi:

```bash
# MCP server khởi động được và đọc đúng vai
printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' \
  | XORA_ROLE=dev sh -c '. ~/.xora/env; exec python3 ~/.agent-plugins/xora-dev/mcp/xora_mcp.py'

# Gọi AAP thật một lần
sh -c '. ~/.xora/env; XORA_ROLE=dev python3 ~/.agent-plugins/xora-dev/mcp/xora_mcp.py --check "hello"'

# Sensor convention trên các file đang đổi của repo
cd <repo> && python3 ~/.agent-plugins/xora-dev/sensor/convention_sensor.py
```

## Sự cố thường gặp

| Hiện tượng | Xử lý |
|---|---|
| Server đỏ, log `python3: not found` | Cài Python 3.9+, hoặc sửa `python3` trong `<repo>/.kiro/settings/mcp.json` thành đường dẫn tuyệt đối |
| Tool trả `Missing XORA_BASE_URL / XORA_API_KEY / XORA_USER_ID env vars` | Chưa có hoặc thiếu dòng trong `~/.xora/env` — làm lại Bước 1, rồi Retry |
| `HTTP 401` / `403` | API key sai / hết hạn, hoặc email chưa được cấp quyền trên AAP |
| `HTTP 422` | Nội dung gửi sai định dạng (ví dụ `text` quá 20.000 ký tự) — chia nhỏ nội dung |
| `HTTP 429` | Vượt giới hạn gọi; chờ đúng `Retry-After` giây rồi thử lại |
| `HTTP 503` | AAP tạm ngừng — đợi vài phút; kéo dài thì báo quản trị AAP |
| Không thấy skill khi gõ `/` | Mỗi skill phải là thư mục `.kiro/skills/<tên>/SKILL.md`; mở phiên chat mới sau khi cài |
| Agent làm theo quy tắc cũ sau khi cập nhật | Phiên chat cũ còn giữ skill / steering bản trước — mở phiên chat mới |
| Script báo *Workspace không được nằm trong agents-plugin/* | Đang cài vào chính thư mục plugin — chọn repo dự án |
| Sensor không chạy khi agent sửa code | Cài `dev` theo repo (không `--global`); kiểm hook *Convention sensor* đang bật; repo có `convention.md` |
| Sensor báo nhầm một dòng | Thêm `convention-ignore: <ID> <lý do>` trên dòng đó (cần Developer đồng ý) và ghi vào mục *Nợ kỹ thuật* của changeset; luật báo nhầm nhiều thì sửa `pattern` trong `convention.md` |

## Gỡ cài đặt

Trong `<repo>/.kiro/`, xoá:

- các thư mục skill của vai (bảng *Mỗi vai cài gì*);
- `steering/xora-<vai>.md` và steering riêng của vai;
- `hooks/convention-sensor.kiro.hook` (với `dev`);
- khối `xora-<vai>` trong `settings/mcp.json`.

Sau đó xoá `~/.agent-plugins/xora-<vai>`. Gỡ hết mọi vai thì xoá luôn `~/.xora/env`.
