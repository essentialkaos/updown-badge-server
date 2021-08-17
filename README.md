<p align="center"><a href="#readme"><img src="https://gh.kaos.st/updown-badge.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/updown-badge-server/ci"><img src="https://kaos.sh/w/updown-badge-server/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/r/updown-badge-server"><img src="https://kaos.sh/r/updown-badge-server.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/b/updown-badge-server"><img src="https://kaos.sh/b/e17e7c90-8b26-4af8-8737-22a86cea9b45.svg" alt="Codebeat badge" /></a>
  <a href="https://kaos.sh/w/updown-badge-server/codeql"><img src="https://kaos.sh/w/updown-badge-server/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`updown-badge-server` is a service for generating badges for [updown.io](https://updown.io) checks.

### Installation

#### From [ESSENTIAL KAOS Public Repository](https://yum.kaos.st)

```bash
sudo yum install -y https://yum.kaos.st/get/$(uname -r).rpm
sudo yum install updown-badge-server
```

### Badges examples

| Endpoint              | Badges |
|-----------------------|--------|
| `/{token}/status.svg` | ![status-up](.github/images/status_up.svg) ![status-down](.github/images/status_down.svg) |
| `/{token}/uptime.svg` ||
| `/{token}/apdex.svg`  ||

### Build Status

| Branch | Status |
|--------|----------|
| `master` | [![CI](https://kaos.sh/w/updown-badge-server/ci.svg?branch=master)](https://kaos.sh/w/updown-badge-server/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/updown-badge-server/ci.svg?branch=develop)](https://kaos.sh/w/updown-badge-server/ci?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
