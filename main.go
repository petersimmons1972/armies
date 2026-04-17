package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/petersimmons1972/armies/cmd"
)

//go:embed examples/generals
var GeneralsFS embed.FS

func main() {
	cmd.RegisterSeedCommand(GeneralsFS)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
