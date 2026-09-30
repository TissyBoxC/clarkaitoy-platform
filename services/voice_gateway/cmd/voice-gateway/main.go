// Command voice-gateway starts the 芽系列·初芽 realtime voice gateway.
package main

import (
	"fmt"
	"os"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
