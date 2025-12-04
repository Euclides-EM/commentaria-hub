package python

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RunScript(pythonExecutable, script string) error {
	tmp, err := os.CreateTemp("", "script-*.py")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	tmp.WriteString(script)
	tmp.Close()

	cmd := exec.Command(pythonExecutable, tmp.Name())

	cmd.Stdin = strings.NewReader(script)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python error: %v\n%s", err, out)
	}

	log.Printf("Python output:\n%s", out)
	return nil
}
