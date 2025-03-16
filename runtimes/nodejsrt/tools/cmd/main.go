package main

import (
	"fmt"
	"os"

	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt/tools"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tools <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  prepare-direct  - Prepare code from tar file")
		fmt.Println("  prepare-test    - Prepare test environment")
		fmt.Println("  prepare-env     - Prepare runner environment")
		fmt.Println("  execute        - Execute code directly")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "prepare-direct":
		tools.PrepareDirect(os.Args[2:])
	case "prepare-test":
		tools.PrepareTest(os.Args[2:])
	case "prepare-env":
		tools.PrepareEnv(os.Args[2:])
	case "execute":
		tools.ExecuteDirect(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
