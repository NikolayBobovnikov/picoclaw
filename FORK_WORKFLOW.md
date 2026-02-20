# Picoclaw Fork Workflow Guide

**Fork (origin):** https://github.com/NikolayBobovnikov/picoclaw
**Upstream:** https://github.com/sipeed/picoclaw

---

## Remote Configuration

```
origin    git@github.com:NikolayBobovnikov/picoclaw.git (fetch/push)  # Your fork (main)
upstream  https://github.com/sipeed/picoclaw.git (fetch)              # Upstream
```

---

## Daily Workflow

### 1. Sync from Upstream (Pull Latest Changes)

```bash
# Fetch upstream changes
git fetch upstream

# Merge upstream/main into your branch (or rebase)
git checkout dev/local
git merge upstream/main

# OR rebase for cleaner history
git rebase upstream/main
```

### 2. Push Changes to Your Fork

```bash
# Push your branch to origin (your fork)
git push origin dev/local

# Force push if you rebased (careful!)
git push origin dev/local --force-with-lease
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
git push origin feature/your-feature-name
```

### Syncing Feature Branch with Upstream

```bash
git fetch upstream
git rebase upstream/main
git push origin feature/your-feature-name --force-with-lease
```

---

## Quick Reference Commands

| Action | Command |
|--------|---------|
| Pull upstream changes | `git fetch upstream && git merge upstream/main` |
| Push to your fork | `git push origin dev/local` |
| Sync fork with upstream | `git fetch upstream && git push origin upstream/main:main` |
| Create PR | Visit GitHub compare page |
| View remotes | `git remote -v` |

---

## Notes

- `origin` = your fork (NikolayBobovnikov/picoclaw) - fetch and push
- `upstream` = upstream (sipeed/picoclaw) - fetch only
- Work on `dev/local` branch for development
- Use `--force-with-lease` instead of `--force` for safer rebase pushes
