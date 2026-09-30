// Command device-platform starts the 芽系列·初芽 device management service.
package main

import (
	"fmt"
	"os"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
