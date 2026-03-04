package generator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"git-recon-viz/pkg/types"
)

//go:embed template/index.html
var indexHTML string

const (
	placeholder = "null/* GIT_RECON_DATA_INJECT */"
)

func GenerateReport(data *types.ReconData, outputPath string) error {
	if indexHTML == "" {
		return fmt.Errorf("embedded HTML template is empty - run 'make build-ui' first")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if !strings.Contains(indexHTML, placeholder) {
		return fmt.Errorf("placeholder not found in template")
	}

	output := strings.Replace(indexHTML, placeholder, string(jsonData), 1)

	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

func ValidateTemplate() error {
	if indexHTML == "" {
		return fmt.Errorf("template not embedded - build UI first")
	}
	if !strings.Contains(indexHTML, placeholder) {
		return fmt.Errorf("placeholder marker missing in template")
	}
	if !strings.Contains(indexHTML, "GIT_RECON_DATA") {
		return fmt.Errorf("template missing GIT_RECON_DATA variable")
	}
	return nil
}

func GetTemplateSize() int {
	return len(indexHTML)
}
