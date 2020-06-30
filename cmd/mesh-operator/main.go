package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mesh-operator/cmd/mesh-operator/app"
	_ "github.com/mesh-operator/pkg/adapter/configcenter/zk"
	_ "github.com/mesh-operator/pkg/adapter/registrycenter/zk"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	rootCmd := app.GetRootCmd(os.Args[1:])

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}
}
