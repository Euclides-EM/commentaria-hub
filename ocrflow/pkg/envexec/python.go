package envexec

import (
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

func PythonCmd(name string, args ...string) error {
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	pythonToolsDir := filepath.Join(rootDir, "python-tools")

	fullArgs := []string{"run", "--project", pythonToolsDir, name}
	fullArgs = append(fullArgs, args...)

	log.Printf("Executing Python command: uv %s", strings.Join(fullArgs, " "))

	return Cmd("uv", fullArgs...)
}

func PythonCmdWithEnv(env map[string]string, name string, args ...string) error {
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	pythonToolsDir := filepath.Join(rootDir, "python-tools")

	fullArgs := []string{"run", "--project", pythonToolsDir, name}
	fullArgs = append(fullArgs, args...)

	log.Printf("Executing Python command with env: uv %s", strings.Join(fullArgs, " "))

	return CmdWithEnv(env, "uv", fullArgs...)
}

func PythonBashCmd(bashCmd string) error {
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	pythonToolsDir := filepath.Join(rootDir, "python-tools")

	wrappedCmd := "uv run --project " + pythonToolsDir + " " + bashCmd

	log.Printf("Executing Python bash command: bash -c \"%s\"", wrappedCmd)

	return Cmd("bash", "-c", wrappedCmd)
}
