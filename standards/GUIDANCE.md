# Xora DLC — MCP routing

When working **with Xora DLC** (work orders, shared/project context from Xora, BRD/design of a step, open questions, conventions/precedents from the company KB), prefer MCP server **`xora-dlc`** and tool **`xora_ask`**. Do not pick another MCP for those intents until `xora_ask` has been tried (or the user names a different system).

Other plugins may ship stronger or conflicting routing rules; follow the user's explicit choice if they override this.

## Intent → MCP

| User intent (examples) | Use |
|---|---|
| Shared / project / work context from Xora DLC; “việc của tôi”; my work on Xora | `xora-dlc` → `xora_ask` |
| Work order / WO-… status, steps, BRD, design, open questions, decisions | `xora-dlc` → `xora_ask` |
| Internal convention, precedents, error-code catalog, security policy from KB | `xora-dlc` → `xora_ask` |
| Save a short tech doc to KB / S3 after explicit user confirm | `xora-dlc` → `xora_ask` |
| Convention check before commit/push the user asked for | Shell: `xdlc sensor` / `xdlc sensor report` (not another MCP) |

## Do not use `xora_ask` for

Pure local git/build/test, editing files the user pointed at, or tools the user named that are not Xora DLC.
