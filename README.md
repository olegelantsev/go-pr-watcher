# go-pr-watcher

## About

Shows open pull requests on configured GitHub repositories.

![go-pr-watcher screenshot](./doc/screenshot.png)

## Getting started

1. Create GitHub personal token with read permissions
2. Create `config.yaml` with the token and repositories

```
repos:
  olegelantsev:
    - gimmeemail
    - pr-watcher
token: <GitHub personal token>
```
