package function

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/aliftech/galus/src/dto"
	"github.com/fsnotify/fsnotify"
)

func RunLiveReload(green, yellow, red, cyan func(a ...interface{}) string) {
	config := dto.Config{
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

	if err := ValidateConfig(&config); err != nil {
		fmt.Println(red("Configuration error:", err))
		os.Exit(1)
	}

	hasGoFiles, err := CheckGoFiles(config.RootDir)
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
				if IsIncludedExt(event.Name, config.IncludeExt) && (event.Op&fsnotify.Write == fsnotify.Write) {
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
		if info.IsDir() && !IsExcludedDir(path, config.ExcludeDir) {
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
