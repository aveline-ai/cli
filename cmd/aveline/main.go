// Command aveline is the Aveline CLI entrypoint.
package main

import (
	"github.com/aveline-ai/cli/internal/cmd"
)

// version is overridden at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	cmd.Execute(version)
}
