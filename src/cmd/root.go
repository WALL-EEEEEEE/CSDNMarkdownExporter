package cmd

import (
	"github.com/spf13/cobra"
)

var (
	avaliable_sites   []string = []string{"csdn"}
	default_cfgFile   string
	default_outputDir string = "blogs"
	cfgFile           string
	outputDir         string
	rootCmd           = &cobra.Command{
		Use:   "BlogExporter",
		Short: "BlogExporter is a blog exporter tool.",
		Long:  `A simple exporter tool for export online blogs hosted on different blog website to markdown files.`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", default_outputDir, "ouptut directory for markdown blogs")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runCmd)
}
