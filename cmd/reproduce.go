package cmd

import (
	"os"

	"github.com/hupe1980/zipbomb/pkg/reproduce"
	"github.com/spf13/cobra"
)

func newSelfReproduceCmd(rootOpts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "reproduce",
		Short:         "Create recursive self-reproducing zipbomb",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			archive, err := os.Create(rootOpts.output)
			if err != nil {
				return err
			}

			defer archive.Close()

			return reproduce.Make(archive)
		},
	}

	return cmd
}
