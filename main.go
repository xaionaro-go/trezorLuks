package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pborman/getopt/v2"
	"github.com/xaionaro-go/trezorLuks/trezor"
)

var (
	// Just some random values (from /dev/random)
	initialKeyValye = []byte{
		0xea, 0x30, 0xe0, 0xc7, 0x11, 0x4a, 0x64, 0x8b, 0x4a, 0xb3, 0x8f, 0xb9, 0xf1, 0x8a, 0x8d, 0xa1,
		0x56, 0x03, 0xbe, 0xd2, 0xa3, 0xba, 0x63, 0x18, 0xf0, 0xd2, 0xda, 0x47, 0x2a, 0x97, 0xfa, 0x48,
	}
	iv = []byte{
		0xf9, 0xa1, 0x99, 0xec, 0xa6, 0x81, 0x78, 0x19, 0xcc, 0x67, 0x55, 0x61, 0x6e, 0xc3, 0x1e, 0xd8,
	}
)

var tmpDir string

func checkError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "Got error:", err)
	os.Exit(-1)
}

func run(stdin io.Reader, cmdName string, params ...string) error {
	fmt.Println("Running:", cmdName, params)
	cmd := exec.Command(cmdName, params...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin
	return cmd.Run()
}

func usage() int {
	getopt.Usage()
	err := run(os.Stdin, "cryptsetup", "--help")
	checkError(err)
	return int(syscall.EINVAL)
}

func main() {
	helpFlag := getopt.BoolLong("help", 'h', "print help message")
	keyNameParameter := getopt.StringLong("trezor-key-name", 0, "luks", "sets the name of a key to be received from the Trezor")
	getopt.Parse()
	args := getopt.Args()

	if *helpFlag {
		os.Exit(usage())
	}

	var luksCmd string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "luks") {
			continue
		}
		luksCmd = arg
		break
	}

	var err error
	var decryptedKey []byte
	switch luksCmd {
	case "":
		os.Exit(usage())

	case "luksOpen", "luksFormat", "luksDump", "luksResume", "luksAddKey", "luksChangeKey":
		fmt.Println("Sent the request to the Trezor device (please confirm the operation if required)")
		trezorInstance := trezor.New()
		decryptedKey, err = trezorInstance.DecryptKey(initialKeyValye, iv, *keyNameParameter)
		checkError(err)
		args = append([]string{"--key-file", "-"}, args...)
	}

	err = run(bytes.NewReader(decryptedKey), "cryptsetup", args...)
	checkError(err)
}
