package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute(version string) {
	printLogo()

	rootCmd := newRootCmd(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type rootOptions struct {
	output string
}

func newRootCmd(version string) *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:           "zipbomb",
		Version:       version,
		Short:         "Tool that creates different types of zip bombs",
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVarP(&opts.output, "output", "o", "bomb.zip", "output filename")

	cmd.AddCommand(
		newNoOverlapCmd(opts),
		newOverlapCmd(opts),
	)

	return cmd
}

func printInfof(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[i] %s\n", fmt.Sprintf(format, a...))
}

func emptyLine() {
	fmt.Fprintf(os.Stderr, "\n")
}

func printLogo() {
	fmt.Fprint(os.Stderr, ` _____ _     _____           _   
|__   |_|___| __  |___ _____| |_ 
|   __| | . | __ -| . |     | . |
|_____|_|  _|_____|___|_|_|_|___|
	|_|                      `, "\n\n")
}
