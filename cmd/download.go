package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/airworthy/pkg/airworthy"
)

var flags = &airworthy.Flags{}

var downloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) >= 1 {
			flags.URL = args[0]
		}
		a := airworthy.New(newLogger(cmd))
		a.Must(a.Run(flags))
	},
}

func init() {
	downloadCmd.Flags().StringVarP(
		&flags.Output,
		"output",
		"o",
		"",
		"specify an output path",
	)

	downloadCmd.Flags().StringVarP(
		&flags.SignatureArmored,
		"signature-armored",
		"s",
		"",
		"specify an URL for the armored signature file",
	)

	downloadCmd.Flags().StringVarP(
		&flags.SignatureBinary,
		"signature-binary",
		"S",
		"",
		"specify an URL for the armored signature file",
	)

	downloadCmd.Flags().StringVar(
		&flags.SHA256Sums,
		"sha256sums",
		"",
		"specify an URL for the sha256sum file",
	)

	RootCmd.AddCommand(downloadCmd)
}
