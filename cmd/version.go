package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"localgate/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "バージョン情報を表示する",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version.FormatOutput())
		},
	}
}

func init() {
	rootCmd.AddCommand(newVersionCmd())
}
