// +build linux

package trezor

import (
	"fmt"
	"log"

	"github.com/conejoninja/tesoro/transport"
	"github.com/zserge/hid"
)

var (
	ErrNoTrezor = fmt.Errorf("No Trezor devices found.")
)

func (trezor *trezor) Reconnect() error {
	success := false
	for !success {
		hid.UsbWalk(func(device hid.Device) {
			info := device.Info()
			if info.Vendor == 21324 && info.Product == 1 && info.Interface == 0 {
				var t transport.TransportHID
				t.SetDevice(device)
				trezor.Client.SetTransport(&t)
				trezor.Device = device
				success = true
				return
			}
		})
		if !success {
			log.Print("No Trezor devices found.")
			trezor.pinentry.SetPrompt("No Trezor devices found.")
			trezor.pinentry.SetDesc("Please check connection to your Trezor device.")
			trezor.pinentry.SetOK("Retry")
			trezor.pinentry.SetCancel("Unmount")
			shouldContinue := trezor.pinentry.Confirm()
			if !shouldContinue {
				log.Print("Cannot continue without Trezor devices.")
				return ErrNoTrezor
			}
		} else if !trezor.Ping() {
			log.Panic("An unexpected behaviour of the trezor device.")
		}
	}
	return nil
}

