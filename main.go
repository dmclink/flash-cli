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

	parsedArgs, err := parser.ParseArgs(os.Args)
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
		// cobra logs execution errors to os.Stderr by default
		os.Exit(1)
	}
}
