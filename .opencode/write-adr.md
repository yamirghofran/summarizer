---
description: write an ADR (Architecture Decision Record) document and save in Linear or markdown file.
agent: build
model: opencode/deepseek-v4-flash-free #opencode-go/deepseek-v4-flash #opencode-go/mimo-v2.5
---

Based on the conversation and action history (you can also spawn parallel `explore` subagents to explore and understand the codebase), summarize the important decisions made and work done and create an ADR based on the Linear ADR Template (also attached below). If the ADR corresponds directly to a Linear issue, attach the ADR document to that issue. If not, attach it under the "Documentation" project in Linear. If you can't connect to Linear, save the ADR as a markdown file and inform the user.

# Template:
Status

Accepted

## Context

The Rails monolith handles 10k RPM. CPU is pegged at 90% during peak. We need better concurrency for the next business inflection point.

## Decision

We will rewrite the external API gateway in Go (Chi framework). The monolith will remain for internal admin tools.

## Consequences

- Positive: Lower latency, cheaper infra, easier horizontal scaling.

- Negative: Team must learn Go; initial deploy tooling must support dual stacks.

- Risks: Database contention if Go services blast the same Postgres pool.
