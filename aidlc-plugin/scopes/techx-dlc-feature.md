---
name: techx-dlc-feature
plugin: techx-dlc
depth: Minimal
keywords:
  - brd
  - techx-dlc-feature
description: Một BRD đã phê duyệt trên hệ thống đã onboard - đường ngắn nhất từ yêu cầu tới bằng chứng
skeleton: off
existing_code: true
runner: true
review_cap: advisory
guard_policy: strict
sensors: on
learnings: off
summary_confirmation: off
plan_approval: on
collaborators: off
---

# techx-dlc-feature scope

Dành cho một BRD đã được review và phê duyệt **bên ngoài** AI-DLC, chạy trên hệ
thống đã qua `techx-dlc-onboard`. Ý tưởng đã chốt, nên toàn bộ ideation là SKIP.

Depth `Minimal` (2-4 câu mỗi bước thay vì 5-8) và `summary_confirmation: off`
cắt phần lớn thời gian mà **không bỏ một cổng duyệt nào** - cả 11 bước nghiệp vụ
vẫn có gate của con người và vẫn ghi audit event. Với ngân hàng, giảm nghi thức
được, giảm bằng chứng thì không.

`guard_policy: strict` giữ nguyên: đây là tiền của khách hàng.
`learnings: off` vì chuẩn đội đã chốt ở `techx-dlc-onboard`.
`collaborators: off` để mỗi bước chạy với lead agent, đổi lại tốc độ.

## Why these stages, why skip those

Bốn bước chịu lực đều giữ: `requirements-analysis` (chủ sở hữu ID FR/NFR mà sensor
traceability dựa vào), `units-generation` (sinh `unit-of-work` điều khiển vòng lặp
construction), `user-stories` (bước duy nhất BA và dev cùng vào một artifact),
`nfr-requirements` (độ trễ, lưu trữ, phân quyền - BRD ngân hàng luôn có).

`reverse-engineering` và `practices-discovery` SKIP vì `techx-dlc-onboard` đã chạy và
kết quả nằm ở cấp space. Mockups SKIP vì backend thuần. Toàn bộ hạ tầng vận hành
SKIP vì đã dựng một lần. `deployment-execution` giữ lại để đưa được thay đổi ra.

## Membership

Initialization, requirements-analysis, user-stories, domain-design,
units-generation, contract-design, delivery-planning, functional-design,
nfr-requirements, code-generation, build-and-test và deployment-execution
execute; phần còn lại SKIP.
