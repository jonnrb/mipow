package mipow

import (
	"fmt"

	"github.com/go-ble/ble"
)

var (
	ServiceUUID = ble.MustParse("ff0d")

	EffectUUID = ble.MustParse("fffb")
	ColorUUID  = ble.MustParse("fffc")
)

type Bulb struct {
	ble.Client
	color, effect *ble.Characteristic
}

func NewBulb(p ble.Client) (*Bulb, error) {
	// TODO(jonnrb): This still necessary bub?
	p.ExchangeMTU(500)

	// Discover services.
	ss, err := p.DiscoverServices([]ble.UUID{ServiceUUID})
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %v", err)
	}

	var color, effect *ble.Characteristic
	for _, s := range ss {
		cs, err := p.DiscoverCharacteristics([]ble.UUID{ColorUUID, EffectUUID}, s)
		if err != nil {
			return nil, fmt.Errorf("failed to discover characteristics (service %v): %v",
				s.UUID, err)
		}
		for _, c := range cs {
			if c.UUID.Equal(ColorUUID) {
				color = c
			} else if c.UUID.Equal(EffectUUID) {
				effect = c
			}
		}

		// Return from here.
		if color != nil && effect != nil {
			return &Bulb{
				Client: p,
				color:  color,
				effect: effect,
			}, nil
		}
	}

	return nil, fmt.Errorf("color and effect characteristics not found on peripheral")
}

func (p *Bulb) Close() error {
	return p.CancelConnection()
}

// Eschews a nice natural color for some (definitely less than) 24-bit color.
// The more you tend toward the zero vector, the dimmer the light gets.
//
func (p *Bulb) SetColor(r, g, b byte) {
	p.WriteCharacteristic(p.color, []byte{0, r, g, b}, true)
}

// Sets the color to the (what I'll dub) "natural" setting, so not per se
// "white". Setting this to zero will, as you'll probably guess, turn off the
// bulb and setting it to 255 will max its brightness.
//
func (p *Bulb) SetWhiteBrightness(brightness byte) {
	p.WriteCharacteristic(p.color, []byte{brightness, 0, 0, 0}, true)
}

func (p *Bulb) setEffect(command []byte) {
	p.WriteCharacteristic(p.effect, command, true)
}

func (p *Bulb) SetRainbowPulse(speed byte) {
	p.setEffect([]byte{
		0x00,
		0x00, 0x00, 0x00,
		0x02,
		0x00,
		speed,
		0x00,
	})
}

func (p *Bulb) SetRainbowFade(speed byte) {
	p.setEffect([]byte{
		0x00,
		0x00, 0x00, 0x00,
		0x03,
		0x00,
		speed,
		0x00,
	})
}

func (p *Bulb) SetPulse(r, g, b, speed byte) {
	p.setEffect([]byte{
		0x00,
		r, g, b,
		0x00,
		0x00,
		speed,
		0x00,
	})
}
