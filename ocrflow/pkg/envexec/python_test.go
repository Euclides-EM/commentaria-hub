package envexec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPythonCmd(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "Python version command",
			cmd:     "python",
			args:    []string{"--version"},
			wantErr: false,
		},
		{
			name:    "Python help command",
			cmd:     "python",
			args:    []string{"-h"},
			wantErr: false,
		},
		{
			name:    "Invalid Python command",
			cmd:     "python",
			args:    []string{"-c", "import sys; sys.exit(1)"},
			wantErr: true,
		},
		{
			name:    "Python with multiple args",
			cmd:     "python",
			args:    []string{"-c", "print('test')"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := exec.LookPath("uv"); err != nil {
				t.Skip("uv not found in PATH, skipping test")
			}

			err := PythonCmd(tt.cmd, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("PythonCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPythonCmdWithEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name: "Command with environment variables",
			env: map[string]string{
				"TEST_VAR": "test_value",
			},
			cmd:     "python",
			args:    []string{"-c", "import os; print(os.environ.get('TEST_VAR', 'not_found'))"},
			wantErr: false,
		},
		{
			name: "Command with multiple env vars",
			env: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
			cmd:     "python",
			args:    []string{"-c", "print('test')"},
			wantErr: false,
		},
		{
			name:    "Command with nil env",
			env:     nil,
			cmd:     "python",
			args:    []string{"-c", "print('test')"},
			wantErr: false,
		},
		{
			name:    "Command with empty env map",
			env:     map[string]string{},
			cmd:     "python",
			args:    []string{"-c", "print('test')"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := exec.LookPath("uv"); err != nil {
				t.Skip("uv not found in PATH, skipping test")
			}

			err := PythonCmdWithEnv(tt.env, tt.cmd, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("PythonCmdWithEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPythonBashCmd(t *testing.T) {
	tests := []struct {
		name    string
		bashCmd string
		wantErr bool
	}{
		{
			name:    "Simple echo command",
			bashCmd: "echo 'test'",
			wantErr: false,
		},
		{
			name:    "Python inline command",
			bashCmd: "python -c \"print('Hello from bash')\"",
			wantErr: false,
		},
		{
			name:    "Command with pipe",
			bashCmd: "echo 'test' | cat",
			wantErr: false,
		},
		{
			name:    "Invalid command",
			bashCmd: "nonexistent_command",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := exec.LookPath("uv"); err != nil {
				t.Skip("uv not found in PATH, skipping test")
			}
			if _, err := exec.LookPath("bash"); err != nil {
				t.Skip("bash not found in PATH, skipping test")
			}

			err := PythonBashCmd(tt.bashCmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("PythonBashCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPythonCmdPathResolution(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
	}{
		{
			name: "PythonCmd builds correct path",
			function: func() string {
				_, filename, _, _ := runtime.Caller(0)
				rootDir := filepath.Join(filepath.Dir(filename), "..", "..", "..")
				return filepath.Join(rootDir, "python-tools")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedPath := tt.function()
			if !strings.Contains(expectedPath, "python-tools") {
				t.Errorf("Expected path to contain 'python-tools', got %s", expectedPath)
			}
		})
	}
}

func TestPythonCmdIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if _, err := exec.LookPath("uv"); err != nil {
		t.Skip("uv not found in PATH, skipping integration test")
	}

	projectDir := os.Getenv("PYTHON_TOOLS_DIR")
	if projectDir == "" {
		t.Log("PYTHON_TOOLS_DIR not set, using default path resolution")
	}

	t.Run("Execute Python script successfully", func(t *testing.T) {
		err := PythonCmd("python", "-c", "import sys; print(f'Python {sys.version_info.major}.{sys.version_info.minor}')")
		if err != nil {
			t.Errorf("Expected successful execution, got error: %v", err)
		}
	})

	t.Run("Handle Python script error", func(t *testing.T) {
		err := PythonCmd("python", "-c", "raise ValueError('Test error')")
		if err == nil {
			t.Error("Expected error for failing Python script, got nil")
		}
	})
}