Install:
```sh
go get github.com/xaionaro-go/trezorLuks
go install github.com/xaionaro-go/trezorLuks
```

Example:
```sh
`go env GOPATH`/bin/trezorLuks luksFormat /dev/loop0
`go env GOPATH`/bin/trezorLuks luksOpen /dev/loop0 mySecureStorage
```

With a custom key (default key name is "luks"):
```sh
`go env GOPATH`/bin/trezorLuks --trezor-key-name myKey luksFormat /dev/loop0
`go env GOPATH`/bin/trezorLuks luksOpen /dev/loop0 mySecureStorage
```

Passing an option to cryptsetup:
```sh
`go env GOPATH`/bin/trezorLuks -- --verbose luksOpen /dev/loop0 mySecureStorage
```

Session example:
```sh
$ sudo ./trezorLuks luksFormat /dev/loop0
Sent the request to the Trezor device (please confirm the operation if required)
Running: cryptsetup [--key-file - luksFormat /dev/loop0]

$ sudo ./trezorLuks luksOpen /dev/loop0 test
Sent the request to the Trezor device (please confirm the operation if required)
Running: cryptsetup [--key-file - luksOpen /dev/loop0 test]

$ ls -ld /dev/mapper/test
lrwxrwxrwx 1 root root 7 Jun  9 17:25 /dev/mapper/test -> ../dm-0

$ sudo ./trezorLuks luksClose test
Running: cryptsetup [luksClose test]

$ ls -ld /dev/mapper/test
ls: cannot access '/dev/mapper/test': No such file or directory
```

Other projects (to encrypt FS using Trezor):
* [gocryptfs](https://github.com/rfjakob/gocryptfs/pull/243)

Documentation:
* [LUKS On-Disk Format Specification Version 1.0](http://clemens.endorphin.org/LUKS-on-disk-format.pdf)

