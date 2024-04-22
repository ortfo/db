# Getting started
## Installation
### Binaries

You can download the latest binaries for Linux, MacOS and Windows from the [releases page](https://github.com/ortfo/db/releases/latest).

Windows
: [64-bit (use this if you don't know)](https://github.com/ortfo/db/releases/latest/download/ortfodb_windows_amd64.exe)
: [32-bit (older computers)](https://github.com/ortfo/db/releases/latest/download/ortfodb_windows_386.exe)

Mac OS
: [64-bit](https://github.com/ortfo/db/releases/latest/download/ortfodb_darwin_amd64)
: [ARM64 (M1 Macs)](https://github.com/ortfo/db/releases/latest/download/ortfodb_darwin_arm64)

Linux
: [64-bit](https://github.com/ortfo/db/releases/latest/download/ortfodb_linux_amd64)
: [32-bit (older computers)](https://github.com/ortfo/db/releases/latest/download/ortfodb_linux_386)
: [ARM64 (Raspberry Pi, etc.)](https://github.com/ortfo/db/releases/latest/download/ortfodb_linux_arm64)


### Package managers

::: tip
Package files are also available to download from [Github Releases](https://github.com/ortfo/db/releases), in case the package manager's repositories are not up-to-date enough
:::

::: warning
This is my first time packaging a program to practically all package managers. I'm not familiar with most of them. If the installation does not work, please [open an issue](https://github.com/ortfo/db/issues/new).
:::

#### Linux

##### Distro-specific

::: code-group

```bash [Arch Linux (AUR)]
paru -S ortfodb-bin
```

```bash [Ubuntu, Debian]
echo "deb [trusted=yes] https://deb.ortfo.org/ /" | sudo tee /etc/apt/sources.list.d/ortfo.list
sudo apt update
sudo apt install ortfodb
```

```bash [Fedora]
# waiting on https://github.com/goreleaser/goreleaser/issues/3136 to add it to COPR
sudo dnf -y install dnf-plugins-core
sudo dnf config-manager --add-repo https://rpm.ortfo.org/
# rpm.ortfo.org is not signed yet, so we need to disable GPG checks
sudo dnf --nogpgcheck install ortfodb
```

```bash [Alpine Linux]
# not available yet
# look for a .apk file in the github releases
# apk add ortfodb
```

```bash [Termux]
# not available yet
# look for a .termux.deb file in the github releases
# pkg install ortfodb
# using Ubuntu's install instructions may work, idk
```

```bash [Nix]
# coming soon™
```

:::

##### Universal

::: code-group

```bash [Snap]
# coming soon™
```

```bash [Flatpak]
# coming soon™
```

```bash [AppImage]
# coming soon™
```

```bash [Homebrew]
# on its own tap for the moment
brew tap ortfo/brew-ortfodb/ortfodb
brew install ortfodb
```

:::

#### MacOS

::: code-group

```bash [Homebrew]
# on its own tap for the moment
brew tap ortfo/brew-ortfodb/ortfodb
brew install ortfodb
```

:::

#### Windows

::: code-group

```powershell [WinGet]
# To be submitted, not yet available
# winget install ortfo.db
```

```powershell [Scoop]
# not on official repos yet
scoop bucket add https://github.com/ortfo/scoop-ortfodb
scoop install ortfodb
```

```powershell [Chocolatey]
# not yet avaiable: needs to be on windows to build the package…
```

:::

### Using `go`

```bash
go install github.com/ortfo/db/cmd@latest
```

### Building from source

#### Requirements

- [Go](https://go.dev)
- [Just](https://just.systems): a modern alternative to Makefiles[^1]. See [Just's docs on installation](https://github.com/casey/just?tab=readme-ov-file#installation)

#### Steps

1. Clone the repository: `git clone https://github.com/ewen-lbh/portfoliodb`
2. `cd` into it: `cd portfoliodb`
3. Compile & install in `~/.local/bin/` `just install`... or simply build a binary to your working directory: `just build`.

[^1]: One big advantage of Just is that it works painlessly on Windows.
