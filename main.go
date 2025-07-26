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

// Banner ASCII art untuk Galus
const banner = `
   ____   _____  __      _    _   ______
  / ___| |___  ||  |    | |  | | |   ___|
 | |  _  ___ | ||  |    | |  | | \__ \__
 | |_| ||  ,,  ||  |___ | \__/ |  __\   |
  \____||______||______| \____/  |_____ |
`

// Version aplikasi
const version = "1.0.0"

// Config menyimpan pengaturan dari file .galus.toml
type Config struct {
	RootDir     string   `toml:"root_dir"`
	TmpDir      string   `toml:"tmp_dir"`
	IncludeExt  []string `toml:"include_ext"`
	ExcludeDir  []string `toml:"exclude_dir"`
	BuildCmd    string   `toml:"build_cmd"`
	BinaryName  string   `toml:"binary_name"`
	CommandArgs []string `toml:"command_args"`
}

// DefaultConfig adalah template konfigurasi default untuk .galus.toml
const defaultConfig = `
root_dir = "."
tmp_dir = "tmp"
include_ext = ["go"]
exclude_dir = [".git", "vendor", "tmp"]
build_cmd = "go build -o ./tmp/main ."
binary_name = "tmp/main"
command_args = ["tmp/main"]
`

func main() {
	// Inisialisasi warna untuk output
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Inisialisasi root command dengan Cobra
	rootCmd := &cobra.Command{
		Use:   "galus",
		Short: "Galus - Live Reloading for Go Applications",
		Long:  fmt.Sprintf("%s\n%s\n\nA live reloading tool for Go applications, similar to Air or CompileDaemon.", cyan(banner), cyan("Galus - Live Reloading for Go Applications")),
		Run: func(cmd *cobra.Command, args []string) {
			runLiveReload(green, yellow, red, cyan)
		},
	}

	// Perintah init
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default .galus.toml configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			configFile := ".galus.toml"
			if _, err := os.Stat(configFile); err == nil {
				fmt.Println(red("Error: Config file already exists: ", configFile))
				os.Exit(1)
			}
			err := os.WriteFile(configFile, []byte(defaultConfig), 0644)
			if err != nil {
				fmt.Println(red("Error creating config file: ", err))
				os.Exit(1)
			}
			fmt.Println(green("Created config file: ", configFile))
		},
	}

	// Perintah version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of Galus",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(green("Galus version: ", version))
		},
	}

	// Tambahkan subcommands ke root command
	rootCmd.AddCommand(initCmd, versionCmd)

	// Jalankan root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(red("Error: ", err))
		os.Exit(1)
	}
}

func runLiveReload(green, yellow, red, cyan func(a ...interface{}) string) {
	// Mode live reload: baca file konfigurasi .galus.toml
	config := Config{
		RootDir:     ".",            // Default: direktori saat ini
		TmpDir:      "tmp",          // Default: direktori sementara
		IncludeExt:  []string{"go"}, // Default: hanya file .go
		ExcludeDir:  []string{".git", "vendor", "tmp"},
		BuildCmd:    "go build -o ./tmp/main .",
		BinaryName:  "tmp/main",
		CommandArgs: []string{"tmp/main"},
	}

	configFile := ".galus.toml"
	if _, err := os.Stat(configFile); err == nil {
		if _, err := toml.DecodeFile(configFile, &config); err != nil {
			fmt.Println(red("Error reading config file: ", err))
			os.Exit(1)
		}
	}

	// Validasi konfigurasi
	if err := validateConfig(&config); err != nil {
		fmt.Println(red("Configuration error: ", err))
		os.Exit(1)
	}

	// Buat direktori sementara jika belum ada
	if err := os.MkdirAll(config.TmpDir, 0755); err != nil {
		fmt.Println(red("Error creating tmp directory: ", err))
		os.Exit(1)
	}

	// Membuat watcher untuk memantau perubahan file
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(red("Error creating watcher: ", err))
		os.Exit(1)
	}
	defer watcher.Close()

	// Variabel untuk menyimpan proses yang sedang berjalan
	var currentProcess *exec.Cmd

	// Fungsi untuk mengompilasi dan menjalankan ulang
	rebuildAndRun := func() {
		// Hentikan proses yang sedang berjalan dengan SIGTERM (jika ada)
		if currentProcess != nil && currentProcess.Process != nil {
			fmt.Println(yellow("Stopping current process..."))
			currentProcess.Process.Signal(syscall.SIGTERM)
			// Berikan waktu 5 detik untuk proses berhenti
			done := make(chan error, 1)
			go func() {
				done <- currentProcess.Wait()
			}()
			select {
			case <-time.After(5 * time.Second):
				fmt.Println(red("Process did not stop in time, killing..."))
				currentProcess.Process.Kill()
			case <-done:
			}
		}

		// Kompilasi ulang kode
		fmt.Println(yellow("Building..."))
		args := strings.Fields(config.BuildCmd)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(red("Build failed: ", err))
			return
		}

		// Jalankan binary
		fmt.Println(green("Running ", config.BinaryName))
		cmd = exec.Command(config.CommandArgs[0], config.CommandArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		currentProcess = cmd
		if err := cmd.Start(); err != nil {
			fmt.Println(red("Failed to start process: ", err))
			return
		}
	}

	// Jalankan kompilasi dan eksekusi pertama kali
	rebuildAndRun()

	// Memantau perubahan file
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Hanya tangani perubahan pada file dengan ekstensi yang diizinkan
				if isIncludedExt(event.Name, config.IncludeExt) && (event.Op&fsnotify.Write == fsnotify.Write) {
					fmt.Println(cyan("File changed: ", event.Name))
					// Tunggu sebentar untuk menghindari multiple trigger
					time.Sleep(100 * time.Millisecond)
					rebuildAndRun()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println(red("Error: ", err))
			}
		}
	}()

	// Tambahkan direktori yang akan dipantau
	err = filepath.Walk(config.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Tambahkan hanya direktori, abaikan direktori yang dikecualikan
		if info.IsDir() && !isExcludedDir(path, config.ExcludeDir) {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println(red("Error walking directory: ", err))
		os.Exit(1)
	}

	fmt.Println(green("Watching for changes in ", config.RootDir))
	// Jaga agar program tetap berjalan
	<-make(chan struct{})
}

// validateConfig memeriksa apakah konfigurasi valid
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
	if _, err := os.Stat(config.CommandArgs[0]); os.IsNotExist(err) {
		return fmt.Errorf("binary %s does not exist", config.CommandArgs[0])
	}
	return nil
}

// isIncludedExt memeriksa apakah file memiliki ekstensi yang diizinkan
func isIncludedExt(name string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(strings.ToLower(name), "."+strings.ToLower(ext)) {
			return true
		}
	}
	return false
}

// isExcludedDir memeriksa apakah direktori termasuk dalam daftar yang dikecualikan
func isExcludedDir(path string, excludeDirs []string) bool {
	for _, dir := range excludeDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}
	return false
}
