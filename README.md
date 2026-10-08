# MonkeForge CLI
`mforge` source code, a cross-platform CLI tool for interacting with [monkeforge](https://monkeforge.org) on your Gorilla Tag installs

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

### `select`
Auto-finds your game path on Steam or lets you find it yourself
```sh
mforge select
# or to skip auto-detection
mforge select -m
```