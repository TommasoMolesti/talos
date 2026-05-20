---
name: Bug report
about: Report incorrect Talos behavior
title: ""
labels: bug
assignees: ""
---

## What Happened

Describe the unexpected behavior.

## Expected Behavior

Describe what you expected Talos to do.

## Reproduction

Include the smallest workflow and command that reproduce the issue.

```yaml
tasks:
  test:
    command: "echo replace-me"
```

```bash
talos run --dry-run
```

## Environment

- Talos version:
- Operating system:
- Shell, if relevant:
- Go version, if building from source:

## Extra Context

Add logs, screenshots, or notes about whether this affects `run`, `validate`, `visualize`, or installation.
