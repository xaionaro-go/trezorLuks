// +build linux

package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pborman/getopt/v2"
	"github.com/xaionaro-go/trezorLuks/trezor"
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

func getMasterkeyMetadata(devicePath string) ([]byte, []byte, error) {
	output, err := exec.Command("cryptsetup", "luksDump", devicePath).Output()
	if err != nil {
		return nil, nil, fmt.Errorf(`Cannot get luksDump for "%v": %v: %v`, devicePath, err, string(output))
	}
	lines := strings.Split(string(output), "\n")

	parseValue := func(lines []string) ([]string, string) {
		var value []string

		keysCount := 0
		for {
			line := lines[0]
			lineParts := strings.Split(line, ":")
			if len(lineParts) > 1 {
				keysCount++
				lineParts = lineParts[1:]
			}
			if keysCount > 1 {
				break
			}
			lines = lines[1:]
			valuePart := strings.Trim(strings.Join(lineParts, ":"), "\t ")

			value = append(value, valuePart)
		}

		return lines, strings.Join(value, "  ")
	}

	var rawDigest string
	var rawSalt string

	for len(lines) != 0 {
		line := lines[0]
		if strings.HasPrefix(line, "MK digest:") {
			lines, rawDigest = parseValue(lines)
		} else
		if strings.HasPrefix(line, "MK salt:") {
			lines, rawSalt = parseValue(lines)
		} else
		if len(lines) > 0 {
			lines = lines[1:]
		}
	}

	convertValue := func(rawValue string) ([]byte, error) {
		rawValue = strings.Replace(rawValue, " ", "", -1)
		return hex.DecodeString(rawValue)
	}

	digest, err := convertValue(rawDigest)
	if err != nil {
		return nil, nil, err
	}
	salt, err := convertValue(rawSalt)
	if err != nil {
		return nil, nil, err
	}

	return digest, salt, nil
}

func getInitialParameters(devicePath string) ([]byte, []byte, error) {
	mkDigest, mkSalt, err := getMasterkeyMetadata(devicePath)
	if err != nil {
		return nil, nil, err
	}

	initialKeyValue := mkSalt
	iv := mkDigest
	if len(iv) > 16 {
		iv = iv[:16]
	}
	return initialKeyValue, iv, nil
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
	var luksCmdIdx int
	for idx, arg := range args {
		if !strings.HasPrefix(arg, "luks") {
			continue
		}
		luksCmd = arg
		luksCmdIdx = idx
		break
	}

	var err error
	var decryptedKey []byte
	var stdin io.Reader
	stdin = os.Stdin
	switch luksCmd {
	case "":
		os.Exit(usage())

	case "luksOpen", "luksFormat", "luksResume", "luksAddKey", "luksChangeKey":

		// We need to create the master key, first
		if luksCmd == "luksFormat" {
			fmt.Println("Initializing with a temporary password (to generate a master key)")
			err = run(bytes.NewReader([]byte{}), "cryptsetup", append([]string{"--key-file", "/proc/cmdline"}, args...)...)
			checkError(err)
		}

		// Search for an argument with the encrypted device path
		fmt.Println("Getting the master key metadata")
		var devicePath string
		for idx, arg := range args[luksCmdIdx+1:] {
			if arg == "--" {
				fmt.Println(len(args), ",", idx, ",", len(args[luksCmdIdx:]))
			}
			if strings.HasPrefix(arg, "-") { // the first non-option (not "-*") after the command argument ("luks*")
				continue
			}
			devicePath = arg
			break
		}

		// Getting initialKeyValue and IV from UUID of the device
		fmt.Println("Generating an initial key and an IV")
		initialKeyValue, iv, err := getInitialParameters(devicePath)
		checkError(err)

		//                     (256b,            32b, 256b,   ?,         ?,       ?,   ?)          -> 256b
		// Getting a real key: (initialKeyValue, IV,  Trezor, BIP32Path, keyName, PIN, Passphrase) -> key
		trezorInstance := trezor.New()
		fmt.Println("Sent a request to the Trezor device (please confirm the operation if required)")
		decryptedKey, err = trezorInstance.DecryptKey(initialKeyValue, iv, *keyNameParameter)
		checkError(err)

		// Using this key to encrypt/decrypt the master key (in `cryptsetup`)
		args = append([]string{"--key-file", "-"}, args...)
		stdin = bytes.NewReader(decryptedKey)

		// Changing the password to the permonent one
		if luksCmd == "luksFormat" {
			fmt.Println("Adding the secure key")
			err = run(stdin, "cryptsetup", "--key-file", "/proc/cmdline", "luksAddKey", devicePath, "-")
			fmt.Println("Removing the temporary key")
			err = run(stdin, "cryptsetup", "--key-file", "/proc/cmdline", "luksRemoveKey", devicePath)
			fmt.Println("Done")
			return
		}
	}

	err = run(stdin, "cryptsetup", args...)
	checkError(err)
	fmt.Println("Done")
}
