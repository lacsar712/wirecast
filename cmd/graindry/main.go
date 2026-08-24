package main

import (
	"os"

	"github.com/lacsar712/wirecast/internal/app"
)

func main() {
	os.Exit(app.RunCLI())
}
