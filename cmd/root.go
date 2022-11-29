package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/hupe1980/zipbomb/pkg/zipbomb"
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
		newSelfReproduceCmd(opts),
		newZipSlipCmd(opts),
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

func printStats(name string, duration time.Duration, zbomb *zipbomb.ZipBomb) error {
	finfo, err := os.Stat(name)
	if err != nil {
		return err
	}

	emptyLine()
	printInfof("Archive: %s", name)
	printInfof("Zip64: %t", zbomb.IsZip64())
	printInfof("Comcompressed size: %d %s", finfo.Size()/1024, "KB")
	printInfof("Uncomcompressed size: %d %s", zbomb.UncompressedSize()/(1024*1024), "MB")
	printInfof("Ratio: %.2f", float64(zbomb.UncompressedSize())/float64(finfo.Size()))
	printInfof("Creating time elapsed: %s\n", duration)

	return nil
}
