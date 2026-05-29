---
description: write a Postmortem / Incident Report and save it in Linear
agent: build
model: opencode/deepseek-v4-flash-free #opencode-go/deepseek-v4-flash #opencode-go/mimo-v2.5
---

Based on the conversation and action history (you can also spawn parallel `explore` subagents to explore and understand the codebase), create a blameless Postmortem / Incident Report that focuses on system failure, not people failure. Create the report as a Linear document and attach it to the related project. If there is no project related, attach it under the "Documentation" project. If you don't have access to Linear, save it as a markdown file and inform the user.

# Template:
**What:** A blameless write-up after an incident. It focuses on system failure, not people failure.

**Why:** Turns pain into permanent process improvement. Another strong O1 evidence item ("I improved organizational reliability").

# Postmortem: [Incident Name] — [YYYY-MM-DD]

## Summary
[Duration] outage. [Impact metric]. Root cause: [one-line root cause].

## Impact
- Revenue: [financial impact].
- Users: [user impact, e.g., failed transactions, degraded experience].
- [Any other relevant impact metrics].

## Timeline (all times UTC)
- `HH:MM` — [Event description].
- `HH:MM` — [Event description].
- `HH:MM` — [Detection / escalation].
- `HH:MM` — [Resolution / recovery].

## Root Cause
[Detailed explanation of the technical or process root cause.]

## Lessons Learned
- [Key takeaway 1].
- [Key takeaway 2].
- [Key takeaway 3].

## Action Items
| Task | Owner | Due |
| ---- | ----- | --- |
| [Action item 1] | @owner | +X days |
| [Action item 2] | @owner | +X days |
| [Action item 3] | @owner | +X days |
