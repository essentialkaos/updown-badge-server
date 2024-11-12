<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/updown-badge-server/ci"><img src="https://kaos.sh/w/updown-badge-server/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/updown-badge-server/codeql"><img src="https://kaos.sh/w/updown-badge-server/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#badges">Badges</a> • <a href="#build-status">Build Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`updown-badge-server` is a service for generating badges for [updown.io](https://updown.io) checks.

### Installation

#### From [ESSENTIAL KAOS Public Repository](https://kaos.sh/kaos-repo)

```bash
sudo dnf install -y https://pkgs.kaos.st/kaos-repo-latest.el$(grep 'CPE_NAME' /etc/os-release | tr -d '"' | cut -d':' -f5).noarch.rpm
sudo dnf install updown-badge-server
```

### Badges

| Endpoint              | Badges |
|-----------------------|--------|
| `/{token}/status.svg` | ![status-up](.github/images/status_up.svg) ![status-down](.github/images/status_down.svg) |
| `/{token}/uptime.svg` | ![uptime-1](.github/images/uptime_1.svg) ![uptime-2](.github/images/uptime_2.svg) ![uptime-3](.github/images/uptime_3.svg) ![uptime-4](.github/images/uptime_4.svg) |
| `/{token}/apdex.svg`  | ![apdex-1](.github/images/apdex_1.svg) ![apdex-2](.github/images/apdex_2.svg) ![apdex-3](.github/images/apdex_3.svg) ![apdex-4](.github/images/apdex_4.svg) |

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
