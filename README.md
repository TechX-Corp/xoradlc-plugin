# Plugin `techx-dlc` cho AI-DLC

Hai scope rút gọn dành cho đội phát triển phần mềm trong ngân hàng, nơi yêu cầu
đến dưới dạng **BRD đã được review và phê duyệt bên ngoài**, và code thì **đã chạy
sẵn trong sản xuất**.

- **Phiên bản**: 0.1.0
- **Yêu cầu**: `aidlc` 2.11.2 trở lên
- **Harness**: Claude Code (bản dựng cho harness khác sinh bằng `aidlc engine plugin build <harness>`)
- **Đóng góp**: 2 scope + 23 overlay. Không thêm stage, agent hay sensor nào.

## Giải quyết vấn đề gì

AI-DLC bản gốc có 30 bước. Thang scope có sẵn thì lưỡng cực: `poc` 5 bước,
`express` 7, `infra` 10, `classic` 15, `mvp` 20 — rồi nhảy thẳng lên `feature` và
`enterprise` đều 30/30. Một feature ngân hàng cỡ vừa không có nấc nào vừa.

Thêm vào đó, ba giả định của bản gốc không đúng với ngân hàng:

| Giả định của AI-DLC | Thực tế ngân hàng |
|---|---|
| Ý tưởng được khám phá trong ideation | BRD đã chốt và ký trước khi chạm AI-DLC |
| Mỗi lần chạy đều cần hiểu lại codebase | Một hệ thống chỉ cần quét một lần |
| Hạ tầng CI/CD dựng cùng feature | Dựng một lần, dùng cho mọi feature sau |

Plugin này tách hai nhịp đó ra: **onboarding một lần cho mỗi hệ thống**, rồi
**một vòng ngắn cho mỗi BRD**.

## Hai scope

### `techx-dlc-onboard` — 9/30 bước

Chạy **một lần cho mỗi hệ thống**, không phải mỗi feature.

```
initialization   workspace-scaffold, workspace-detection, state-init
inception        reverse-engineering, practices-discovery
construction     nfr-design, infrastructure-design, ci-pipeline
operation        deployment-pipeline, environment-provisioning,
                 observability-setup, incident-response
```

Kết quả đổ vào `aidlc/spaces/<space>/codekb/` và `memory/team.md` ở **cấp space**,
nên tồn tại xuyên suốt mọi intent sau này. Đó là lý do `techx-dlc-feature` không cần
chạy lại `reverse-engineering` hay `practices-discovery`.

Cấu hình: `depth: Standard`, `learnings: on`, `summary_confirmation: on`,
`collaborators: on`. Đây là lần duy nhất đội ngồi xuống chốt chuẩn cho cả hệ
thống — đáng để làm kỹ.

### `techx-dlc-feature` — 11/30 bước

Chạy **mỗi BRD đã phê duyệt**.

```
initialization   workspace-scaffold, workspace-detection, state-init
inception        requirements-analysis, user-stories, domain-design,
                 units-generation, contract-design, delivery-planning
construction     functional-design, nfr-requirements, code-generation,
                 build-and-test
operation        deployment-execution
```

Cấu hình: `depth: Minimal` (2–4 câu hỏi mỗi bước thay vì 5–8),
`summary_confirmation: off`, `learnings: off`, `collaborators: off`,
nhưng **`guard_policy: strict`** và **`sensors: on`**.

> **Giảm nghi thức, không giảm bằng chứng.** Cả 11 bước vẫn có cổng duyệt của con
> người và vẫn ghi audit event. Thứ bị cắt là số câu hỏi và bước xác nhận tóm tắt,
> không phải cổng duyệt.

### Bốn bước chịu lực được giữ nguyên

| Bước | Vì sao không bỏ được |
|---|---|
| `requirements-analysis` | Chủ sở hữu ID FR/NFR trong `traceability.json`; 8 stage chạy sensor traceability dựa vào nó |
| `units-generation` | Sinh `unit-of-work` — thứ điều khiển `for_each` của 5 stage construction |
| `user-stories` | Stage duy nhất `mode: mob`; chỗ BA và dev cùng viết vào một artifact |
| `nfr-requirements` | Độ trễ, lưu trữ, phân quyền — BRD ngân hàng luôn có mục này |

`reverse-engineering` và `practices-discovery` cũng chịu lực, nhưng đã nằm ở
`techx-dlc-onboard`.

## Chạy thế nào

Lần đầu trên một hệ thống:

```
/aidlc --scope techx-dlc-onboard --project-type brownfield "lập đường cơ sở cho hệ thống <tên>"
```

Mỗi BRD sau đó:

```
/aidlc-knowledge onboard docs/<BRD>.md
/aidlc --scope techx-dlc-feature "triển khai <mã BRD>: <tiêu đề>"
```

Ở `units-generation` của feature đầu tiên, ghìm xuống **1–2 Unit**. Năm bước
construction lặp theo từng Unit, nên số lần dừng nhân với số Unit.

## Cấu trúc

```
plugins/techx-dlc/
├── .aidlc-plugin/plugin.json   manifest: name, version, aidlc.contributes
├── scopes/
│   ├── techx-dlc-onboard.md         tên file phải là <plugin>-<name>.md
│   └── techx-dlc-feature.md         và name: trong frontmatter phải khớp stem
├── contributions/              23 overlay, một file mỗi stage lõi
│   ├── initialization/   3
│   ├── inception/        8
│   ├── construction/     7
│   └── operation/        5
└── hooks/                      bản sao nguyên văn template compose của engine
```

Membership của scope **không** nằm trong file scope mà ở trường `scopes:` của
từng stage. Plugin gắn vào bằng overlay thay vì sửa stage lõi:

```yaml
---
target: requirements-analysis
plugin: techx-dlc
adds:
  scopes:
    - techx-dlc-feature
---
```

Sửa thẳng 23 stage file lõi cũng chạy, nhưng `aidlc update` ghi đè chúng và hai
scope chết lặng. Overlay nằm ngoài, sống sót qua mọi lần nâng cấp, và copy sang
ngân hàng khác bằng một thư mục.


## Cài đặt theo từng harness

Nguồn plugin — 2 scope + 23 overlay — **trung lập với harness**. Không có dòng nào
nói về một công cụ cụ thể: scope là frontmatter YAML, overlay là `target` +
`adds.scopes`, cả hai do engine `aidlc` đọc. Nhưng **cách mỗi host nhận plugin thì
khác nhau thật sự**, và tài liệu chính thức không phủ hết.

| Harness | Cách cài | Tài liệu chính thức |
|---|---|---|
| Claude Code | catalogue (marketplace) | ✅ có, lệnh cụ thể |
| Codex CLI | catalogue (marketplace) | ✅ có, lệnh cụ thể |
| Kiro CLI V2 | chép thư mục + sync thủ công | ✅ có, lệnh cụ thể |
| Kiro IDE / CLI V3 | chép thư mục, hook tự compose | ✅ có, lệnh cụ thể |
| Cursor | chép thư mục, hook tự compose | ⚠️ chỉ có cơ chế hook, **không có bước cài** |
| opencode | chép thư mục + sync thủ công | ⚠️ chỉ có lệnh sync, **không có bước cài** |
| GitHub Copilot | chưa rõ | ❌ **không có tài liệu nào** |

Ba dòng cuối: bản dựng vẫn ra đúng (`plugin build` trả `COMPLETE` cho cả bảy),
nhưng bước đăng ký với host là phần tự suy từ cơ chế, chưa được tài liệu xác nhận.
Dùng ba harness đó thì coi như đang đi đường chưa có biển báo.

### Bước 0 — Dựng bản cài

```
aidlc engine plugin build <harness> dist/techx-dlc-<harness> --plugin-root plugins/techx-dlc
```

Dạng tài liệu chính thức ghi là positional, không có `--plugin-root`:

```
bun <thư-mục-tools>/aidlc-plugin-build.ts plugins/techx-dlc <harness> dist/techx-dlc-<harness>
```

Cả hai đều chạy trên `aidlc` 2.11.2. Build tự validate trước và **từ chối nếu có
lỗi** (thoát mã 1); harness lạ thoát mã 2. Không truyền outDir thì mặc định ra
`plugins/techx-dlc/dist/<harness>/`.

### 1. Claude Code

Đẩy bản dựng `dist/techx-dlc-claude/` lên một repo git, rồi trong Claude Code:

```
/plugin marketplace add <tổ-chức>/<repo-plugin>
/plugin install techx-dlc@techx-dlc
```

Không chép gì vào project — bạn đăng ký catalogue. Hook SessionStart có sẵn sẽ
compose ở phiên kế tiếp. Cài từ thư mục local cũng được:
`/plugin marketplace add <đường-dẫn>/dist/techx-dlc-claude`.

Cập nhật về sau: `/plugin marketplace update techx-dlc`.

### 2. Codex CLI

```
codex plugin marketplace add <tổ-chức>/<repo-plugin>
codex plugin add techx-dlc@techx-dlc
```

Chạy lệnh `marketplace add` **từ bên trong một git repo**. Codex chỉ đọc
`.agents/plugins/marketplace.json` — bản dựng đã có sẵn file đó đúng chỗ.

Codex bắt **duyệt tin cậy hook một lần cho mỗi plugin**, ghim theo hash nội dung.
Duyệt xong thì hook SessionStart tự compose ở lần tương tác đầu.

### 3. Kiro CLI V2 — bản dựng `kiro`

Không có catalogue. Chép thẳng cây `.kiro/` vào project:

```
cp -r dist/techx-dlc-kiro/.kiro <project>/
```

Plugin nằm ở `<project>/.kiro/aidlc-plugins/techx-dlc/`. Kiro CLI V2 **không chạy
compose hook**, nên phải sync tay:

```
AIDLC_PLUGIN_ROOT="<project>/.kiro/aidlc-plugins/techx-dlc" \
AIDLC_PROJECT_DIR="<project>" \
AIDLC_HARNESS_DIR=.kiro \
aidlc engine plugin sync
```

Nếu lệnh trên không chạy được, dùng bản dự phòng:

```
AIDLC_PLUGIN_ROOT="<project>/.kiro/aidlc-plugins/techx-dlc" \
AIDLC_PROJECT_DIR="<project>" \
AIDLC_HARNESS_DIR=.kiro \
bun "<project>/.kiro/aidlc-plugins/techx-dlc/hooks/compose.ts"
```

Rồi mở phiên:

```
kiro-cli --v2
/aidlc
```

Cần Kiro CLI **2.21.4 trở lên** cho cờ `--v2`.

### 4. Kiro IDE và Kiro CLI V3 — bản dựng `kiro-ide`

```
cp -r dist/techx-dlc-kiro-ide/.kiro <project>/
```

Bản này kèm `.kiro/hooks/aidlc-techx-dlc-compose.json` — một đăng ký SessionStart
timeout 1800 giây, tự chạy compose. Không cần sync tay.

Sau khi chép, trong **Kiro IDE**:

1. Mở project, bấm **Trust Folder & Continue** (hoặc Trust trong Workspace Trust)
2. Chạy **Developer: Reload Window**
3. Chọn agent **aidlc** trong bộ chọn agent của khung chat
4. Chạy `/aidlc --doctor`

Với **Kiro CLI V3**: `kiro-cli --v3`, chọn "Trust, continue" nếu được hỏi.

### 5. Cursor — chưa có tài liệu cài

Bản dựng có `hooks/hooks.json` dùng cú pháp camelCase của Cursor
(`hooks.sessionStart[].command`) với hai mục, nên cơ chế compose là tự động.
Nhưng **tài liệu không nói chép bản dựng đi đâu hay đăng ký thế nào**.

Hướng khả dĩ: chép nội dung `dist/techx-dlc-cursor/` vào gốc project rồi mở phiên
Cursor mới và kiểm bằng `aidlc engine orchestrate help | grep techx-dlc`. Nếu scope
không hiện, chạy sync tay như opencode bên dưới. Đội nào làm đầu tiên nên ghi lại
cách đã chạy được và cập nhật mục này.

### 6. opencode — chưa có tài liệu cài

opencode **không chạy compose hook**. Sau khi đưa bản dựng vào project, sync tay:

```
aidlc engine plugin sync --plugin-root <đường-dẫn-tới-thư-mục-plugin>
```

Tài liệu có lệnh sync này nhưng **không có bước chép hay đăng ký**. Cùng tình
trạng như Cursor.

### 7. GitHub Copilot — không có tài liệu

Bản dựng ra được (`.plugin/` làm manifest, engine ở `.aidlc/`), nhưng tài liệu
chính thức **không nói gì** về cài, sync hay nối hook plugin cho Copilot. Chưa
khuyến nghị dùng đường này cho ngân hàng cho tới khi upstream bổ sung.

### Sau khi cài — giống nhau ở mọi harness

```
aidlc engine plugin select techx-dlc
aidlc engine plugin sync
aidlc engine plugin list
```

`select` chỉ nhận plugin đã cài. Nếu project chưa từng đặt lựa chọn nào thì plugin
vừa compose đã hoạt động ngay; nhưng khi đã có một lựa chọn, compose **không** tự
bật plugin mới — phải `select`.

Xác nhận:

```
aidlc engine orchestrate help | grep techx-dlc
```

Phải ra hai dòng `techx-dlc-feature` và `techx-dlc-onboard`.

**Khởi động lại?** Tài liệu không yêu cầu restart sau khi cài plugin. Compose chạy
ở `SessionStart`, nên **mở một phiên mới** là đủ trên Claude Code, Codex, Kiro IDE
và Cursor. Mục xử lý sự cố của tài liệu chỉ nói: *"Refreshed skills do not take
effect → Start a new harness session."*

### Khác biệt về tin cậy — đọc kỹ nếu triển khai cho ngân hàng

| Harness | Cổng tin cậy |
|---|---|
| Claude Code | theo `strictKnownMarketplaces` do tổ chức quản lý |
| Codex CLI | duyệt hook một lần cho mỗi plugin, ghim theo hash nội dung |
| Kiro | **không có cổng nào** — chép thư mục vào *là* quyết định tin cậy |

Với Kiro, hãy đối xử với plugin như `git clone && chạy`: rà soát nội dung trước khi
chép, và chỉ lấy từ nguồn nội bộ của tổ chức.

### Điều kiện chung

AI-DLC phải đã cài trên harness đó trước. Plugin đóng góp scope **vào** AI-DLC,
nó không tự chạy một mình.

Plugin này **không đóng góp agent nào**, nên không vướng yêu cầu của Kiro về việc
phải viết tay `.kiro/agents/<agent>.json` kèm mục `trustedAgents` cho agent do
plugin cung cấp.

## Phát triển

```
aidlc engine plugin validate plugins/techx-dlc      # phải trả VALID
aidlc engine plugin build claude dist/techx-dlc-claude --plugin-root plugins/techx-dlc
aidlc engine graph validate-scope techx-dlc-feature # kiểm phụ thuộc artifact
```

Bốn cảnh báo `[advisory]` khi kiểm scope là **đúng thiết kế**, không phải lỗi:
`ci-pipeline` cần kết quả build từ một lần chạy feature; `deployment-execution`
trong `techx-dlc-feature` cần `cd-config` và `environment-inventory` từ
`techx-dlc-onboard`. Engine nói đúng — *"Ensure existing artifact is current"*.

## Chưa có trong v0.1.0

| Việc | Làm được bằng plugin? |
|---|---|
| Stage `brd-ingestion` — BRD thành requirement có thẻ nguồn | Có, chưa làm |
| Stage `external-approval-attest` — ghi ai duyệt BRD, lúc nào, hash file | Có, chưa làm |
| Stage `release-evidence-export` — gói bằng chứng cho release gate | Có, chưa làm |
| Sensor `brd-coverage` — mọi mục BRD phải được phủ bởi ≥1 requirement | Có, chưa làm |
| Thẻ nguồn `[doc:...]` | **Không** — sensor `claim-sources` chỉ nhận `[desc]` `[scope]` `[Q<n>]` `[memory:<id>]` `[assumption]`; phải sửa lõi |
| Audit event cho phê duyệt ngoài | **Không** — phải sửa engine |
| Maker-checker (người duyệt khác người làm) | **Không** — gate chỉ cần một người bấm Approve, và audit shard định danh bằng `.aidlc-clone-id` chứ không phải con người |

Ba dòng cuối là lý do cần quyết sớm: fork AI-DLC hay đẩy PR ngược lên upstream.
Riêng maker-checker sẽ là câu hỏi đầu tiên bộ phận tuân thủ của ngân hàng đặt ra.
