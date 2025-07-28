package dto

type Config struct {
	RootDir     string   `toml:"root_dir"`
	TmpDir      string   `toml:"tmp_dir"`
	IncludeExt  []string `toml:"include_ext"`
	ExcludeDir  []string `toml:"exclude_dir"`
	BuildCmd    string   `toml:"build_cmd"`
	BinaryName  string   `toml:"binary_name"`
	CommandArgs []string `toml:"command_args"`
}
