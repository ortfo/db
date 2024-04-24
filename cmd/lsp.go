package main

import (
	"os"
	"path/filepath"

	"github.com/ortfo/languageserver"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start a Language Server Protocol server for ortfo",
	Run: func(cmd *cobra.Command, args []string) {
		logger, _ := zap.NewProductionConfig().Build()
		languageserver.StartServer(logger, flags.Config, filepath.Join(os.Getenv("HOME"), "/.local/share/ortfo/lsp/logs/"))
	},
}

func init() {
	lspCmd.PersistentFlags().Bool("stdio", false, "Used for compatibility with VSCode. Ignored (the server is always started in stdio mode)")
	rootCmd.AddCommand(lspCmd)
}
