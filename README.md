# MonkeForge CLI
`mforge` source code, a cross-platform CLI tool for interacting with [monkeforge](https://monkeforge.org) on your Gorilla Tag installs

## Installation
The CLI has install commands for both Windows 10/11 and Linux as shown on the [MonkeForge app](https://monkeforge.org/get-app) page. Copy the command for your system and the script will handle the rest:

Windows (requires batch shell, run `cmd` if in PowerShell):
```bat
curl -sL "https://monkeforge.org/cli/setup.bat" | cmd
```

Linux:
```sh
curl -sL "https://monkeforge.org/cli/setup.sh" | bash
```

## Basics
### `install`
Install a package (MonkeFrames for example)
```sh
# you can use either a package name (com.monkeframes.monkeframes) or basic ID (monkeframes)
# get this from the address bar on monkeforge.org or scroll down and copy the GUID field
mforge install monkeframes

# disable checksum verification
mforge install monkeframes -N

# select a specific release channel
mforge install monkeframes -r dev
```

### `uninstall`
Uninstall a package (MonkeFrames for example)
```sh
mforge uninstall monkeframes
```

### `upgrade`
Upgrade all packages on the install
```sh
mforge upgrade
# accepts same arguments as `mforge install`
```

### `login`
Log in to MonkeForge for downloading private mods and release channels:
```sh
mforge login
# opens a browser for login, just hit Accept in Discord and that's all you have to do
```

### `logout`
If logged in, you can logout. This is very helpful documentation.
```
mforge logout
# you are now logged out yay
```

### `select`
Auto-finds your game path on Steam or lets you find it yourself
```sh
mforge select
# or to skip auto-detection
mforge select -m
```
