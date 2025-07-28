package dto

const DefaultConfig = `
root_dir = "."
tmp_dir = "tmp"
include_ext = ["go"]
exclude_dir = [".git", "vendor", "tmp"]
build_cmd = "go build -o ./tmp/main.exe ."
binary_name = "tmp/main.exe"
command_args = ["tmp/main.exe"]
`
