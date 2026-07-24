package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmclink/flash-cli/cmd"
	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/parser"
)

func main() {
	if os.Getuid() == 0 {
		fmt.Fprintln(os.Stderr, "Error: do not run this application as root/sudo")
		os.Exit(1)
	}

	structuralRoot := cmd.NewRootCmd(nil)
	validCommands := make(map[string]bool)
	for _, c := range structuralRoot.Commands() {
		validCommands[c.Name()] = true
		for _, alias := range c.Aliases {
			validCommands[alias] = true
		}
	}

	parsedArgs, err := parser.ParseArgs(os.Args, validCommands)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: parsing and validating args\n\n%v", err)
		os.Exit(1)
	}
	// TODO: implement parsedArgs.CobraArgs() or something more readable and set os.Args to its return value
	// might need to change ParsedArgs struct to store the binary name
	os.Args = parsedArgs.Args(os.Args[0])

	a, err := app.NewApp(parsedArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: initializing app\n\n%v", err)
		os.Exit(1)
	}
	defer a.Close()

	rootCmd := cmd.NewRootCmd(a)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		rootCmd.Printf("Error %v\n", err)
		fmt.Println()
		rootCmd.Usage()
		os.Exit(1)
	}
}
