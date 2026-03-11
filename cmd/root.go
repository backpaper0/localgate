package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"localgate/internal/logger"
)

var debugFlag bool

var rootCmd = &cobra.Command{
	Use:   "localgate",
	Short: "ローカルサービスのための動的リバースプロキシ",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.SetDebug(debugFlag)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "デバッグログを有効にする (LOCALGATE_DEBUG=1 でも有効)")
}
