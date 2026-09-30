// Command voice-gateway starts the Clarkaitoy realtime voice gateway.
package main

import (
	"os"

	"github.com/clarkaitoy/voice_gateway/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
