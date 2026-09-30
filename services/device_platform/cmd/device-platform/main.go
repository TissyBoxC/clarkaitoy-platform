// Command device-platform starts the Clarkaitoy device management service.
package main

import (
	"fmt"
	"os"

	"github.com/clarkaitoy/device_platform/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
