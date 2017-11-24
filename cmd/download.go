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
	RootCmd.AddCommand(downloadCmd)
}
