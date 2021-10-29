# go-pr-watcher

[![Go](https://github.com/olegelantsev/go-pr-watcher/actions/workflows/go.yml/badge.svg)](https://github.com/olegelantsev/go-pr-watcher/actions/workflows/go.yml)

---

Shows open pull requests on configured GitHub repositories.

![go-pr-watcher screenshot](./doc/screenshot.png)

## Getting started

1. Create GitHub personal token with read permissions
2. Create `config.yaml` with repositories map

```yaml
github:
  olegelantsev: # repo owner
    - gimmeemail # repo name
    - pr-watcher
```

## Features

[x] Show GitHub PRs
[ ] Show Bitbucket PRs
[ ] Show GitLab PRs
[x] Securely store token in the keyring
[x] Async load of PRs and table refresh
[ ] Smart load of many PRs repos with filtering
[ ] Improve test coverage
[ ] Refresh PRs list periodically
