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
Initializing with a temporary password (to generate a master key)
Running: cryptsetup [--key-file /proc/cmdline luksFormat /dev/loop0]
Getting the master key metadata
Generating an initial key and an IV
Sent a request to the Trezor device (please confirm the operation if required)
Adding the secure key
Running: cryptsetup [--key-file /proc/cmdline luksAddKey /dev/loop0 -]
Removing the temporary key
Running: cryptsetup [--key-file /proc/cmdline luksRemoveKey /dev/loop0]
Done

$ sudo ./trezorLuks luksOpen /dev/loop0 test
Getting the master key metadata
Generating an initial key and an IV
Sent a request to the Trezor device (please confirm the operation if required)
Running: cryptsetup [--key-file - luksOpen /dev/loop0 sdf]
Done

$ ls -ld /dev/mapper/test
lrwxrwxrwx 1 root root 7 Jun  9 17:25 /dev/mapper/test -> ../dm-0

$ sudo ./trezorLuks luksClose test
Running: cryptsetup [luksClose test]
Done

$ ls -ld /dev/mapper/test
ls: cannot access '/dev/mapper/test': No such file or directory
```

Other projects (to encrypt FS using Trezor):
* [gocryptfs](https://github.com/rfjakob/gocryptfs/pull/243)

Documentation:
* [LUKS On-Disk Format Specification Version 1.0](http://clemens.endorphin.org/LUKS-on-disk-format.pdf)
* [SLIP-0011 : Symmetric encryption of key-value pairs using deterministic hierarchy](https://github.com/satoshilabs/slips/blob/master/slip-0011.md)
