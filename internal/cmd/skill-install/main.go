package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xoctopus/genx/pkg/agent"
)

func main() {
	if err := (&agent.Installer{}).Install(context.Background()); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}
}
