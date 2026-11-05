---
name: techx-dlc-onboard
plugin: techx-dlc
depth: Standard
keywords:
  - onboard
  - onboarding
  - baseline
description: Chạy một lần cho mỗi hệ thống ngân hàng - lập đường cơ sở hiểu biết và hạ tầng
skeleton: off
existing_code: true
runner: true
review_cap: advisory
guard_policy: strict
sensors: on
learnings: on
summary_confirmation: on
plan_approval: on
collaborators: on
---

# techx-dlc-onboard scope

Chạy **một lần cho mỗi hệ thống**, không phải mỗi feature. Mục đích là lập đường
cơ sở mà mọi feature sau đó dùng lại: hiểu biết về codebase có sẵn, thực hành của
đội, và hạ tầng CI/CD/quan sát.

Kết quả đổ vào `aidlc/spaces/<space>/codekb/` và `memory/team.md` ở cấp space,
nên tồn tại xuyên suốt mọi intent sau này. Đó là lý do `techx-dlc-feature` không cần
chạy lại `reverse-engineering` hay `practices-discovery`.

Giữ `guard_policy: strict`, `learnings: on` và `summary_confirmation: on`: đây là
lần duy nhất đội ngồi xuống chốt chuẩn cho cả hệ thống, đáng để làm kỹ.
`collaborators: on` để practices-discovery kéo đủ quality, developer và devsecops vào.

## Why these stages, why skip those

Không có yêu cầu nghiệp vụ nào để phân tích ở đây, nên toàn bộ ideation và inception
phía sau reverse-engineering/practices-discovery đều SKIP, construction cũng vậy.
Chỉ chạy phần dựng nền: quét code, chốt thực hành, và bảy bước hạ tầng vận hành.

## Membership

Initialization, reverse-engineering, practices-discovery, nfr-design, infrastructure-design, ci-pipeline,
deployment-pipeline, environment-provisioning, observability-setup và
incident-response execute; phần còn lại SKIP.
