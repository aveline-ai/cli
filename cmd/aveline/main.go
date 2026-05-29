// Command aveline is the CLI entrypoint.
package main

import (
	"os"

	"github.com/aveline-ai/cli/internal/commands"
)

func main() {
	os.Exit(commands.Execute())
}
