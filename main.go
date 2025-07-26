package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

const banner = `
   ____   _____  __      _    _   ______
  / ___| |___  ||  |    | |  | | |   ___|
 | |  _  ___ | ||  |    | |  | | \__ \__
 | |_| ||  ,,  ||  |___ | \__/ |  __\   |
  \____||______||______| \____/  |_____ |
`

const version = "1.0.0"

type Config struct {
	RootDir     string   `toml:"root_dir"`
	TmpDir      string   `toml:"tmp_dir"`
	IncludeExt  []string `toml:"include_ext"`
	ExcludeDir  []string `toml:"exclude_dir"`
	BuildCmd    string   `toml:"build_cmd"`
	BinaryName  string   `toml:"binary_name"`
	CommandArgs []string `toml:"command_args"`
}

const defaultConfig = `
root_dir = "."
tmp_dir = "tmp"
include_ext = ["go"]
exclude_dir = [".git", "vendor", "tmp"]
build_cmd = "go build -o ./tmp/main.exe ."
binary_name = "tmp/main.exe"
command_args = ["tmp/main.exe"]
`

func main() {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	rootCmd := &cobra.Command{
		Use:   "galus",
		Short: "Galus - Live Reloading for Go Applications",
		Long:  fmt.Sprintf("%s\n%s", cyan(banner), cyan("Galus - Live Reloading for Go Applications")),
		Run: func(cmd *cobra.Command, args []string) {
			runLiveReload(green, yellow, red, cyan)
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default .galus.toml configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			configFile := ".galus.toml"
			if _, err := os.Stat(configFile); err == nil {
				fmt.Println(red("Error: Config file already exists:", configFile))
				os.Exit(1)
			}
			if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
				fmt.Println(red("Error creating config file:", err))
				os.Exit(1)
			}
			fmt.Println(green("Created config file:", configFile))
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of Galus",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(green("Galus version:", version))
		},
	}

	rootCmd.AddCommand(initCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(red("Error:", err))
		os.Exit(1)
	}
}

func runLiveReload(green, yellow, red, cyan func(a ...interface{}) string) {
	config := Config{
		RootDir:     ".",
		TmpDir:      "tmp",
		IncludeExt:  []string{"go"},
		ExcludeDir:  []string{".git", "vendor", "tmp"},
		BuildCmd:    "go build -o ./tmp/main .",
		BinaryName:  "tmp/main",
		CommandArgs: []string{"tmp/main"},
	}

	configFile := ".galus.toml"
	if _, err := os.Stat(configFile); err == nil {
		if _, err := toml.DecodeFile(configFile, &config); err != nil {
			fmt.Println(red("Error reading config file:", err))
			os.Exit(1)
		}
	}

	if err := validateConfig(&config); err != nil {
		fmt.Println(red("Configuration error:", err))
		os.Exit(1)
	}

	hasGoFiles, err := checkGoFiles(config.RootDir)
	if err != nil {
		fmt.Println(red("Error checking Go files:", err))
		os.Exit(1)
	}
	if !hasGoFiles {
		fmt.Println(red("Error: No .go files found in", config.RootDir))
		os.Exit(1)
	}

	if err := os.MkdirAll(config.TmpDir, 0755); err != nil {
		fmt.Println(red("Error creating tmp directory:", err))
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(red("Error creating watcher:", err))
		os.Exit(1)
	}
	defer watcher.Close()

	var currentProcess *exec.Cmd

	rebuildAndRun := func() {
		if currentProcess != nil && currentProcess.Process != nil {
			fmt.Println(yellow("Stopping current process..."))
			currentProcess.Process.Signal(syscall.SIGTERM)
			done := make(chan error, 1)
			go func() { done <- currentProcess.Wait() }()
			select {
			case <-time.After(5 * time.Second):
				fmt.Println(red("Force killing process..."))
				currentProcess.Process.Kill()
			case <-done:
			}
		}

		fmt.Println(yellow("Building..."))
		args := strings.Fields(config.BuildCmd)
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(red("Build failed:", err))
			fmt.Println(red("Build output:", string(output)))
			return
		}

		fmt.Println(green("Running:", config.CommandArgs[0]))
		cmd = exec.Command(config.CommandArgs[0], config.CommandArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		currentProcess = cmd
		if err := cmd.Start(); err != nil {
			fmt.Println(red("Failed to start process:", err))
			return
		}
	}

	// Auto build binary first if it doesn't exist
	if _, err := os.Stat(config.BinaryName); os.IsNotExist(err) {
		fmt.Println(yellow("Binary not found, performing initial build..."))
		rebuildAndRun()
	}

	// Initial run regardless
	fmt.Println(yellow("Initial run..."))
	rebuildAndRun()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if isIncludedExt(event.Name, config.IncludeExt) && (event.Op&fsnotify.Write == fsnotify.Write) {
					fmt.Println(cyan("File changed:", event.Name))
					time.Sleep(100 * time.Millisecond)
					rebuildAndRun()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println(red("Watcher error:", err))
			}
		}
	}()

	err = filepath.Walk(config.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && !isExcludedDir(path, config.ExcludeDir) {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println(red("Error walking directory:", err))
		os.Exit(1)
	}

	fmt.Println(green("Watching for changes in", config.RootDir))
	<-make(chan struct{}) // Block forever
}

func validateConfig(config *Config) error {
	if config.BuildCmd == "" {
		return fmt.Errorf("build_cmd cannot be empty")
	}
	if !strings.HasPrefix(strings.TrimSpace(config.BuildCmd), "go build") {
		return fmt.Errorf("build_cmd must start with 'go build'")
	}
	if len(config.CommandArgs) == 0 {
		return fmt.Errorf("command_args cannot be empty")
	}
	return nil
}

func checkGoFiles(rootDir string) (bool, error) {
	hasGoFiles := false
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".go") {
			hasGoFiles = true
		}
		return nil
	})
	return hasGoFiles, err
}

func isIncludedExt(name string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(strings.ToLower(name), "."+strings.ToLower(ext)) {
			return true
		}
	}
	return false
}

func isExcludedDir(path string, excludeDirs []string) bool {
	for _, dir := range excludeDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}
	return false
}
