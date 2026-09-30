// Command voice-gateway starts the Clarkaitoy realtime voice gateway.
package main

import (
	"fmt"
	"os"

	"github.com/clarkaitoy/voice_gateway/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
