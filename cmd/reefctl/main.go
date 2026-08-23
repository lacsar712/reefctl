package main

import (
	"os"

	"github.com/lacsar712/reefctl/internal/app"
)

func main() {
	os.Exit(app.RunCLI())
}
