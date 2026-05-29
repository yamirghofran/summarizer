---
description: write a Runbook and save it in Linear
agent: build
model: opencode/deepseek-v4-flash-free #opencode-go/deepseek-v4-flash #opencode-go/mimo-v2.5
---

Based on the conversation and action history (you can also spawn parallel `explore` subagents to explore and understand the codebase), create a Runbook that includes enough context and step-by-step instructions for someone else to be able to understand, manage, and operate the given process/workflow. Create the Runbook as a Linear document and attach it to the related project. If there is no project related, attach it under the "Documentation" project. If you don't have access to Linear, save it as a markdown file and inform the user.

# Template:
**What:** A checklist for when a specific alert fires or a system breaks. It is operational, not theoretical.

**Why:** Incidents at 3 AM are not the time to reason from first principles. A runbook turns outages into repeatable procedures.

# Runbook: Checkout Service — High Error Rate

**Alert:** `checkout: error_rate > 5%`

## Immediate Checks (0–3 min)
- [ ] Check [Grafana dashboard](link)
- [ ] Check [recent deploys](link) — did someone ship in last 30 min?

## If a bad deploy is suspected (3–5 min)
```bash
kubectl rollout undo deployment/checkout
```
- [ ] Confirm error rate drops in dashboard.

## If database-related (connection pool exhaustion)
- [ ] Check RDS "Connections" metric.
- [ ] If > 80%, consider restarting pods (temporary) and page DBA.

## Escalation
If not resolved in 15 minutes, escalate to CTO.
