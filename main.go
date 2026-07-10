package main

import (
	"fmt"
	"os"

	"github.com/dmclink/flash-cli/cmd"
	"github.com/dmclink/flash-cli/internal/database"
)

// TODO: consider logging errors or otherwise handling differently
// since this prints the whole chain to the user
func main() {
	if os.Getuid() == 0 {
		fmt.Fprintln(os.Stderr, "Error: do not run this application as root/sudo")
		os.Exit(1)
	}

	db, err := database.Open()
	if err != nil {
		// TODO: consider logging the error trace and printing a simple statement to user
		// on this and others
		fmt.Fprintf(os.Stderr, "Error: finding path and opening database | %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	err = database.Init(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: initializing database tables | %v\n", err)
		os.Exit(1)
	}

	err = cmd.Execute(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: executing command | %v\n", err)
	}
}
