# Picoclaw Fork Workflow Guide

**Fork:** https://github.com/NikolayBobovnikov/picoclaw
**Upstream:** https://github.com/sipeed/picoclaw

---

## Remote Configuration

```
origin    https://github.com/sipeed/picoclaw.git (fetch)    # Upstream
fork      git@github.com:NikolayBobovnikov/picoclaw.git (fetch/push)  # Your fork
```

---

## Daily Workflow

### 1. Sync from Upstream (Pull Latest Changes)

```bash
# Fetch upstream changes
git fetch origin

# Merge upstream/main into your branch (or rebase)
git checkout dev/local
git merge origin/main

# OR rebase for cleaner history
git rebase origin/main
```

### 2. Push Changes to Your Fork

```bash
# Push your branch to fork
git push fork dev/local

# Force push if you rebased (careful!)
git push fork dev/local --force-with-lease
```

### 3. Create Pull Request to Upstream

Visit: https://github.com/sipeed/picoclaw/compare/main...NikolayBobovnikov:picoclaw:dev/local

---

## Branch Management

### Working on a Feature

```bash
# Create feature branch from dev/local
git checkout -b feature/your-feature-name

# Make changes and commit
git add .
git commit -m "feat: description"

# Push to fork
git push fork feature/your-feature-name
```

### Syncing Feature Branch with Upstream

```bash
git fetch origin
git rebase origin/main
git push fork feature/your-feature-name --force-with-lease
```

---

## Quick Reference Commands

| Action | Command |
|--------|---------|
| Pull upstream changes | `git fetch origin && git merge origin/main` |
| Push to your fork | `git push fork dev/local` |
| Sync fork with upstream | `git fetch origin && git push fork origin/main:main` |
| Create PR | Visit GitHub compare page |
| View remotes | `git remote -v` |

---

## Notes

- `origin` = upstream (sipeed/picoclaw) - fetch only
- `fork` = your fork (NikolayBobovnikov/picoclaw) - fetch and push
- Work on `dev/local` branch for development
- Use `--force-with-lease` instead of `--force` for safer rebase pushes
