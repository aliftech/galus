package function

import (
	"fmt"
	"strings"

	"github.com/aliftech/galus/src/dto"
)

func ValidateConfig(config *dto.Config) error {
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
