package ranger_utils

import (
	"bufio"
	"os"
	"strings"
)

func ExportEnvVars(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = line[7:]
		}

		token := strings.SplitN(line, "=", 2)
		// Remove spaces, ' and "
		key := strings.Trim(strings.Trim(strings.TrimSpace(token[0]), "'"), "\"")
		value := strings.Trim(strings.Trim(strings.TrimSpace(token[1]), "'"), "\"")

		os.Setenv(key, value)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
