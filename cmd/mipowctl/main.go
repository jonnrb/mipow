package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-ble/ble"
	"github.com/jonnrb/mipow"
	"github.com/jonnrb/mipow/helper"
)

func main() {
	err := helper.InitDevice()
	if err != nil {
		fmt.Printf("Error initializing device: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) != 2 {
		fmt.Println("Expected a verb (either \"allon\" or \"alloff\").")
		os.Exit(1)
	}

	switch verb := os.Args[1]; verb {
	case "list":
		listBulbs()
	case "allon", "alloff":
		fmt.Println("Connecting to all nearby MiPOW bulbs.")

		var wg sync.WaitGroup
		connectToAll(func(b *mipow.Bulb) {
			wg.Add(1)
			go func() {
				switch verb {
				case "allon":
					turnOn(b)
				case "alloff":
					turnOff(b)
				default:
					panic("condition should not be possible")
				}
				wg.Done()
			}()
		})
		wg.Wait()
	default:
		fmt.Println("Expected verb to be either \"list\", \"allon\", or \"alloff\".")
		os.Exit(1)
	}

}

func turnOn(b *mipow.Bulb) {
	fmt.Printf("Turning on %v.\n", b.Addr())
	b.SetWhiteBrightness(255)

	time.Sleep(time.Second)
	if err := b.CancelConnection(); err != nil {
		fmt.Printf("Error closing connection to %q: %v", b.Addr(), err)
	}
}

func turnOff(b *mipow.Bulb) {
	fmt.Printf("Turning off %v.\n", b.Addr())
	b.SetWhiteBrightness(0)

	time.Sleep(time.Second)
	if err := b.CancelConnection(); err != nil {
		fmt.Printf("Error closing connection to %q: %v", b.Addr(), err)
	}
}

const longEnough = 15 * time.Second

func listBulbs() {
	ctx, cancel := timedInterruptableContext(longEnough)
	defer cancel()

	err := helper.Scan(ctx, func(a ble.Advertisement) {
		fmt.Printf("%q (%v)\n", a.LocalName(), a.Addr())
	})
	if err != context.DeadlineExceeded {
		panic(err)
	}
}

func connectToAll(each func(b *mipow.Bulb)) {
	ctx, cancel := timedInterruptableContext(longEnough)
	defer cancel()

	var (
		ads   = make(chan ble.Advertisement)
		bulbs = make(chan *mipow.Bulb)

		scanErr error
	)

	go func() {
		scanErr = helper.Scan(ctx, func(a ble.Advertisement) {
			ads <- a
		})
	}()

	go func() {
		defer close(bulbs)
		for a := range ads {
			c, err := ble.Dial(context.TODO(), a.Addr())
			if err != nil {
				fmt.Printf("Error dialing %q: %v\n", a.LocalName(), err)
				continue
			}

			b, err := mipow.NewBulb(c)
			if err != nil {
				fmt.Printf("Error creating bulb from %q: %v\n", a.LocalName(), err)
				continue
			}

			bulbs <- b
		}
	}()

	func() {
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-bulbs:
				each(b)
			}
		}
	}()

	if scanErr != nil && scanErr != context.DeadlineExceeded {
		fmt.Printf("Scan error: %v\n", scanErr)
	}
}

func timedInterruptableContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	go func() {
		c := make(chan os.Signal, 1)
		defer close(c)

		signal.Notify(c, os.Interrupt)
		defer signal.Stop(c)

		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
