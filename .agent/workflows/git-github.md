---
description: Git and GitHub CLI commands for GIIA Core Engine repository management
---

# Git & GitHub Workflow

This workflow defines safe commands for Git operations and GitHub CLI interactions.

## Git Operations

// turbo
1. Checkout a branch:
```bash
git checkout <branch>
```

// turbo
2. Stage changes:
```bash
git add <files>
```

// turbo
3. Commit changes:
```bash
git commit -m "<message>"
```

// turbo
4. Push changes:
```bash
git push
```

// turbo
5. Move/rename files:
```bash
git mv <source> <destination>
```

## GitHub CLI - Repository Info

// turbo
6. View repository info:
```bash
gh repo view melegattip/giia-core-engine --json visibility,isPrivate
```

// turbo
7. Check PR status:
```bash
gh pr view <number> --repo melegattip/giia-core-engine
```

// turbo
8. Check PR checks:
```bash
gh pr checks <number> --repo melegattip/giia-core-engine
```

// turbo
9. List PRs:
```bash
gh pr list --repo melegattip/giia-core-engine --limit 100
```

// turbo
10. List workflow runs:
```bash
gh run list --repo melegattip/giia-core-engine --limit 10
```

// turbo
11. View failed workflow logs:
```bash
gh run view <run-id> --repo melegattip/giia-core-engine --log-failed
```

## GitHub CLI - API Operations

// turbo
12. Check repository environments:
```bash
gh api repos/melegattip/giia-core-engine/environments
```

// turbo
13. Check actions permissions:
```bash
gh api repos/melegattip/giia-core-engine/actions/permissions
```

// turbo
14. Check branch protection:
```bash
gh api repos/melegattip/giia-core-engine/branches/main/protection
```

## GitHub CLI - Configuration (Requires Review)

15. Enable GitHub Actions:
```bash
gh api repos/melegattip/giia-core-engine/actions/permissions --method PUT --field enabled=true --field allowed_actions=all
```

16. Configure workflow permissions:
```bash
gh api repos/melegattip/giia-core-engine/actions/permissions/workflow --method PUT --field default_workflow_permissions=write --field can_approve_pull_request_reviews=true
```

17. Create environments:
```bash
gh api repos/melegattip/giia-core-engine/environments/development --method PUT
gh api repos/melegattip/giia-core-engine/environments/staging --method PUT
gh api repos/melegattip/giia-core-engine/environments/production --method PUT
```

18. Configure branch protection:
```bash
gh api repos/melegattip/giia-core-engine/branches/main/protection --method PUT --input branch-protection-main.json
gh api repos/melegattip/giia-core-engine/branches/develop/protection --method PUT --input branch-protection-develop.json
```

## Authentication

19. Login to GitHub:
```bash
gh auth login --web --git-protocol https --hostname github.com
```
