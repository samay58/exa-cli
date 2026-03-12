package main

import (
	"context"
	"os"

	"github.com/samaydhawan/exa-cli/internal/app"
	"github.com/samaydhawan/exa-cli/internal/version"
)

func main() {
	info := version.Info{
		Version: version.Version,
		Commit:  version.Commit,
		Date:    version.Date,
	}

	os.Exit(app.Run(
		context.Background(),
		os.Args[1:],
		os.Stdin,
		os.Stdout,
		os.Stderr,
		os.Environ(),
		info,
	))
}
