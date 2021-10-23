# go-pr-watcher

[![Go](https://github.com/olegelantsev/go-pr-watcher/actions/workflows/go.yml/badge.svg)](https://github.com/olegelantsev/go-pr-watcher/actions/workflows/go.yml)

---

Shows open pull requests on configured GitHub repositories.

![go-pr-watcher screenshot](./doc/screenshot.png)

## Getting started

1. Create GitHub personal token with read permissions
2. Create `config.yaml` with repositories map

```yaml
repos:
  olegelantsev: # repo owner
    - gimmeemail # repo name
    - pr-watcher
```
