package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/micromdm/micromdm/platform/device"
)

func (cmd *removeCommand) removeDevices(args []string) error {
	flagset := flag.NewFlagSet("remove-devices", flag.ExitOnError)
	var (
		flIdentifier = flagset.String("udid", "", "device UDID, optionally comma separated")
		flSerial     = flagset.String("serial", "", "device serial, optionally comma separated")
	)
	flagset.Usage = usageFor(flagset, "mdmctl remove devices [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	if *flIdentifier == "" && *flSerial == "" {
		return errors.New("bad input: device UDID or Serial must be provided")
	}

	opts := device.RemoveDevicesOptions{}
	if *flIdentifier != "" {
		opts.UDIDs = strings.Split(*flIdentifier, ",")
	}
	if *flSerial != "" {
		opts.Serials = strings.Split(*flSerial, ",")
	}

	ctx := context.Background()
	err := cmd.devicesvc.RemoveDevices(ctx, opts)
	if err != nil {
		return err
	}

	fmt.Printf("removed devices(s): %s\n", strings.Join(append(opts.UDIDs, opts.Serials...), ", "))

	return nil
}
