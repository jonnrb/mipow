// +build darwin

// TODO(jonnrb): Linux support duh.
package helper

import (
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/darwin"
)

func InitDevice() error {
	opt := darwin.OptCentralRole() // Lets us connect to devices.
	d, err := darwin.NewDevice(opt)
	if err != nil {
		return err
	}

	ble.SetDefaultDevice(d)
	return nil
}
