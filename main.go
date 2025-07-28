package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/aliftech/galus/src/dto"
	"github.com/aliftech/galus/src/function"
)

func main() {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	rootCmd := &cobra.Command{
		Use:   "galus",
		Short: "Galus - Live Reloading for Go Applications",
		Long:  fmt.Sprintf("%s\n%s", cyan(dto.Banner), cyan("Galus - Live Reloading for Go Applications")),
		Run: func(cmd *cobra.Command, args []string) {
			function.RunLiveReload(green, yellow, red, cyan)
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: cyan("Create a default .galus.toml configuration file"),
		Long:  fmt.Sprintf("%s \n%s", cyan(dto.Banner), cyan("Create a default .galus.toml configuration file")),
		Run: func(cmd *cobra.Command, args []string) {
			configFile := ".galus.toml"
			if _, err := os.Stat(configFile); err == nil {
				fmt.Println(red("Error: Config file already exists:", configFile))
				os.Exit(1)
			}
			if err := os.WriteFile(configFile, []byte(dto.DefaultConfig), 0644); err != nil {
				fmt.Println(red("Error creating config file:", err))
				os.Exit(1)
			}
			fmt.Println(green("Created config file:", configFile))
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: cyan("Print the current version of Galus"),
		Long:  fmt.Sprintf("%s\n%s", cyan(dto.Banner), cyan("Print the current version of Galus")),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(green("Galus version:", dto.Version))
		},
	}

	rootCmd.AddCommand(initCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(red("Error:", err))
		os.Exit(1)
	}
}
