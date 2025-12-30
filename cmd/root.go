package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "template-backend-go",
	Short: "A CLI application with HTTP server capabilities",
	Long:  `A command-line application built with Cobra that provides HTTP server functionality using Gin.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add commands here
}
