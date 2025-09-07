package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	lineCount = "7145"
	wordCount = "58164"
	byteCount = "342190"
	charCount = "339292"
)

var (
	binPath   string
	inputPath = filepath.Join("testdata", "input.txt")
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "ccwc-*")
	if err != nil {
		os.Exit(1)
	}
	binPath = filepath.Join(tmpDir, "ccwc")
	// Add .exe suffix if windows
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runErr := cmd.Run()
	// Ensure cancel runs before exiting (defer won't run after os.Exit)
	cancel()
	if runErr != nil {
		_ = os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}

type runResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func runCLI(ctx context.Context, args []string, stdin []byte) runResult {
	cmd := exec.CommandContext(ctx, binPath, args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = 127 // spawn/ctx failure
		}
	}
	return runResult{stdout: out.String(), stderr: errb.String(), exitCode: code}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func TestFileCases(t *testing.T) {
	type (
		argFn      func(path string) []string
		expectedFn func(path string) string
	)

	cases := []struct {
		name     string
		args     argFn
		expected expectedFn
	}{
		{
			name: "default",
			args: func(path string) []string { return []string{path} },
			expected: func(path string) string {
				return strings.Join([]string{lineCount, wordCount, byteCount, path}, "\t") + "\n"
			},
		},
		{
			name: "lines",
			args: func(path string) []string { return []string{"-l", path} },
			expected: func(path string) string {
				return strings.Join([]string{lineCount, path}, "\t") + "\n"
			},
		},
		{
			name: "words",
			args: func(path string) []string { return []string{"-w", path} },
			expected: func(path string) string {
				return strings.Join([]string{wordCount, path}, "\t") + "\n"
			},
		},
		{
			name: "bytes",
			args: func(path string) []string { return []string{"-c", path} },
			expected: func(path string) string {
				return strings.Join([]string{byteCount, path}, "\t") + "\n"
			},
		},
		{
			name: "chars",
			args: func(path string) []string { return []string{"-m", path} },
			expected: func(path string) string {
				return strings.Join([]string{charCount, path}, "\t") + "\n"
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			res := runCLI(ctx, tc.args(inputPath), nil)
			if res.exitCode != 0 {
				t.Fatalf("exit=%d stderr=%q", res.exitCode, res.stderr)
			}

			want := tc.expected(inputPath)
			if res.stdout != want {
				t.Fatalf("stdout mismatch.\n got: %q\nwant: %q", res.stdout, want)
			}
		})
	}
}

func TestStdinCases(t *testing.T) {
	type (
		expectedFn func() string
	)

	cases := []struct {
		name     string
		args     []string
		expected expectedFn
	}{
		{
			name: "default",
			args: nil,
			expected: func() string {
				return strings.Join([]string{lineCount, wordCount, byteCount}, "\t") + "\n"
			},
		},
		{
			name: "lines",
			args: []string{"-l"},
			expected: func() string {
				return lineCount + "\n"
			},
		},
		{
			name: "words",
			args: []string{"-w"},
			expected: func() string {
				return wordCount + "\n"
			},
		},
		{
			name: "bytes",
			args: []string{"-c"},
			expected: func() string {
				return byteCount + "\n"
			},
		},
		{
			name: "chars",
			args: []string{"-m"},
			expected: func() string {
				return charCount + "\n"
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			stdin := readFile(t, inputPath)
			res := runCLI(ctx, tc.args, stdin)
			if res.exitCode != 0 {
				t.Fatalf("exit=%d stderr=%q", res.exitCode, res.stderr)
			}

			want := tc.expected()
			if want != res.stdout {
				t.Fatalf("stdout mismatch.\ngot: %q\nwant: %q", res.stdout, want)
			}
		})
	}
}

func TestErrorCases(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{
			name: "nonexistent-file",
			args: []string{"-c", filepath.Join("testdata", "does-not-exist.txt")},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			res := runCLI(ctx, tc.args, nil)
			if res.exitCode == 0 {
				t.Fatalf("expected non-zero exit, got 0; stderr=%q", res.stderr)
			}

			ls := strings.ToLower(res.stderr)
			if !strings.HasPrefix(ls, "error:") {
				t.Fatalf("stderr does not start with %q. got %q", "error:", res.stderr)
			}
		})
	}
}
