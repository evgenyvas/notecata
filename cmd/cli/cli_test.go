package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
)

var update = flag.Bool("update", false, "update golden files")

var binaryName = "notecata-coverage"

var binaryPath = ""

func fixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), "testdata", fixture)
}

func writeFixture(t *testing.T, fixture string, content []byte) {
	err := os.WriteFile(fixturePath(t, fixture), content, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func loadInput(t *testing.T, input *string) *string {
	if input == nil {
		return nil
	}

	content, err := os.ReadFile(fixturePath(t, *input))
	if err != nil {
		t.Fatal(err)
	}

	return ptr(string(content))
}

func loadFixture(t *testing.T, fixture string) string {
	content, err := os.ReadFile(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(content))
}

func ptr(s string) *string {
	return &s
}

func waitFor(r *bufio.Reader, needle string) error {
	var buf strings.Builder

	for {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}

		buf.WriteByte(b)

		if strings.Contains(buf.String(), needle) {
			return nil
		}
	}
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\r\n", "\n")
}

func cleanup() {
	errStr := "could not clean data after tests: %v"
	_, err := runBinary([]string{"-d", "/test.md"}, nil)
	if err != nil {
		fmt.Printf(errStr, err)
	}
	_, err = runBinary([]string{"-d", "/subdir"}, nil)
	if err != nil {
		fmt.Printf(errStr, err)
	}
}

func TestCliArgs(t *testing.T) {
	defer cleanup()

	tests := []struct {
		name    string
		args    []string
		input   *string
		fixture string
	}{
		{"no arguments", []string{}, nil, "no-args.golden"},
		{"create in root", []string{"/test.md"}, ptr("create-nometa.input"), "create-nometa.golden"},
		{"view", []string{"/test.md"}, nil, "view.golden"},
		{"delete", []string{"-d", "/test.md"}, nil, "delete.golden"},
		{"create in subdirectory", []string{"/subdir/test.md"}, ptr("create-nometa.input"), "create-nometa-subdirectory.golden"},
		{"view in subdirectory", []string{"/subdir/test.md"}, nil, "view.golden"},
		{"view subdirectory", []string{"/subdir"}, nil, "view-subdirectory.golden"},
		{"delete subdirectory", []string{"-d", "/subdir"}, nil, "delete-subdirectory.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dateNow := time.Now()
			output, err := runBinary(tt.args, loadInput(t, tt.input))

			if err != nil {
				t.Fatal(err)
			}

			if *update {
				writeFixture(t, tt.fixture, output)
			}

			actual := normalize(string(output))

			expected := normalize(loadFixture(t, tt.fixture))
			expected = strings.ReplaceAll(expected, "{{ DATE_NOW }}", dateNow.Format("2006-01-02 15:04:05"))

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("actual = %s, expected = %s", actual, expected)
			}
		})
	}
}

func TestCliMetadata(t *testing.T) {
	fileName := "testint.md"

	defer func() {
		errStr := "could not clean data after tests: %v"
		_, err := runBinary([]string{"-d", "/" + fileName}, nil)
		if err != nil {
			fmt.Printf(errStr, err)
		}
	}()

	t.Run("create in root asking metadata", func(t *testing.T) {

		cmd := exec.Command(
			"bash",
			"-c",
			fmt.Sprintf("%s -m / < %s",
				binaryPath,
				fixturePath(t, "create-meta.input"),
			),
		)

		cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")

		ptmx, err := pty.Start(cmd)
		if err != nil {
			t.Fatal(err)
		}
		defer ptmx.Close()

		var transcript bytes.Buffer
		reader := bufio.NewReader(
			io.TeeReader(ptmx, &transcript),
		)

		// filename
		err = waitFor(reader, "Enter filename:")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		fmt.Fprintln(ptmx, fileName)

		// title
		err = waitFor(reader, "Enter title:")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		fmt.Fprintln(ptmx, "Test input")

		// tags
		err = waitFor(reader, "Enter tags separated by spaces:")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		fmt.Fprintln(ptmx, "input test")

		err = cmd.Wait()
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(&transcript, ptmx) // read remain output

		actual := normalize(transcript.String())
		fmt.Printf("\nCaptured Output:\n%s\n-----\n", actual)

		expected := normalize(loadFixture(t, "create-meta.golden"))

		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("actual = %s, expected = %s", actual, expected)
		}
	})
}

func TestMain(m *testing.M) {
	err := os.Chdir("../..")
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("could not get current dir: %v", err)
	}

	binaryPath = filepath.Join(dir, binaryName)

	os.Exit(m.Run())
}

func runBinary(args []string, input *string) ([]byte, error) {
	cmd := exec.Command(binaryPath, args...)

	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")

	if input != nil {
		cmd.Stdin = strings.NewReader(*input)
		return cmd.CombinedOutput()
	} else {
		// emulate tty
		ptm, err := pty.Start(cmd)
		if err != nil {
			panic(err)
		}
		defer ptm.Close()

		// Create a channel to wait for the output to finish being read
		done := make(chan struct{})
		var output []byte

		// Read from the PTY in real-time
		go func() {
			buf, error := io.ReadAll(ptm)
			if error != nil && error != io.EOF {
				fmt.Printf("error reading from pty: %v", error)
			}
			output = buf
			close(done)
		}()

		// Wait for the command to finish
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("command finished with error: %v", err)
			return nil, err
		}

		// Wait for the ReadAll goroutine to wrap up
		<-done

		fmt.Printf("\nCaptured Output:\n%s\n-----\n", strings.TrimSpace(string(output)))
		return output, nil
	}
}
