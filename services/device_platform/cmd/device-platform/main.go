// Command device-platform starts the Clarkaitoy device management service.
package main

import (
	"os"

	"github.com/clarkaitoy/device_platform/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
