package helper

import (
	"context"

	"github.com/go-ble/ble"
	"github.com/jonnrb/mipow"
)

// Wraps ble.Scan but filters to MiPOW bulbs.
//
func Scan(ctx context.Context, handler func(a ble.Advertisement)) error {
	seen := make(map[string]bool) // TODO(jonnrb): I don't think the "seen" flag works upstream.
	filter := func(a ble.Advertisement) bool {
		if seen[a.Addr().String()] {
			return false
		}
		services := a.Services()
		for _, s := range services {
			if s.Equal(mipow.ServiceUUID) {
				seen[a.Addr().String()] = true
				return true
			}
		}
		return false
	}
	return ble.Scan(ctx, false, handler, filter)
}
