package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func DocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "docs",
		Hidden: true,
		Short:  "Generates mmomni documentation",
		Args:   cobra.NoArgs,
		Run:    docsCmdF,
	}

	cmd.Flags().StringP("directory", "d", "docs", "The directory where the docs will be generated in")

	return cmd
}

func docsCmdF(cmd *cobra.Command, _ []string) {
	outDir, _ := cmd.Flags().GetString("directory")

	fileInfo, err := os.Stat(outDir)
	if err != nil {
		if !os.IsNotExist(err) {
			errAndExit(fmt.Errorf("error checking the output directory %q: %w", outDir, err))
		}
		if createErr := os.Mkdir(outDir, 0755); createErr != nil {
			errAndExit(fmt.Errorf("error creating the output directory %q: %w", outDir, err))
		}
	} else if !fileInfo.IsDir() {
		errAndExit(fmt.Errorf("file %q exists and is not a directory", outDir))
	}

	if err := doc.GenReSTTree(RootCmd(), outDir); err != nil {
		errAndExit(fmt.Errorf("error generating the documentation at %q: %w", outDir, err))
	}
}
