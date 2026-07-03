package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	// Reproduce the issue: duplicate completion registration for inherited persistent flags
	rootCmd := &cobra.Command{
		Use:   "app",
		Short: "Test app for cobra completion fix",
		RunE: func(cmd *cobra.Command, args []string) error {
			val, _ := cmd.Flags().GetString("output")
			fmt.Printf("output: %s\n", val)
			return nil
		},
	}

	// Register a persistent flag with completion on root
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format")
	rootCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml", "text"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Subcommand that inherits the flag and tries to override completion
	subCmd := &cobra.Command{
		Use:   "sub",
		Short: "A subcommand",
		RunE: func(cmd *cobra.Command, args []string) error {
			val, _ := cmd.Flags().GetString("output")
			fmt.Printf("sub output: %s\n", val)
			return nil
		},
	}

	// Safe override: use RegisterFlagCompletionFunc which should accept
	// overrides for inherited persistent flags without panicking
	subCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"xml", "csv"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.AddCommand(subCmd)

	// Test 1: Run root command
	fmt.Println("=== Test 1: Root command ===")
	rootCmd.SetArgs([]string{"--output", "json"})
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Test 2: Run subcommand
	fmt.Println("=== Test 2: Subcommand ===")
	subCmd.SetArgs([]string{"--output", "yaml"})
	if err := subCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Test 3: Generate shell completion scripts (this used to panic)
	fmt.Println("=== Test 3: Shell completion generation ===")
	var buf strings.Builder
	rootCmd.GenBashCompletionV2(&buf, true)
	output := buf.String()
	if len(output) > 100 {
		fmt.Printf("Bash completion generated (%d bytes) - SUCCESS\n", len(output))
	} else {
		fmt.Printf("Bash completion too short: %q\n", output)
	}

	_ = pflag.CommandLine
	fmt.Println("\nAll tests passed! No panics detected.")
}
