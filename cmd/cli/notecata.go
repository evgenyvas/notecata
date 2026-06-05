package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"notecata/internal/lib"
	"notecata/utils"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mattn/go-isatty"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout))
}

func run(args []string, stdin io.Reader, stdout io.Writer) int {
	// search parameters
	var isDelete bool
	var askMeta bool
	for i, v := range args {
		if slices.Contains([]string{"--del", "-del", "-d"}, v) {
			isDelete = true
			args = slices.Delete(args, i, i+1)
		}
		if slices.Contains([]string{"--meta", "-meta", "-m"}, v) {
			askMeta = true
			args = slices.Delete(args, i, i+1)
		}
	}

	if len(args) == 0 {
		fmt.Fprintln(stdout, "error: missing argument")
		return 0
	}
	path := args[0]

	storageService := lib.GetStorageService()
	note, err := storageService.Store.Note(path)
	if err != nil && !errors.Is(err, lib.ErrPathIsDirectory) {
		if isTerminal() {
			fmt.Fprintf(stdout, "Can't open %s, %s\n", path, err)
			return 1
		} else {
			code := saveNote(stdin, stdout, path, askMeta)
			fmt.Fprintf(stdout, "New note %s created successfully\n", path)
			return code
		}
	}

	if isDelete {
		_, err = storageService.Store.DeleteNote(path)
		if err != nil {
			return 1
		}
		return 0
	}

	if errors.Is(err, lib.ErrPathIsDirectory) {
		fmt.Fprintf(stdout, "%s is a directory\n", path)

		if isTerminal() {
			// show list of contents
			notes, err := storageService.Store.Notes(path)
			if err != nil {
				return 1
			}
			if len(notes) > 0 {
				fmt.Fprintln(stdout, "\nContents:")
			} else {
				fmt.Fprintln(stdout, "\nDirectory is empty:")
			}
			for _, n := range notes {
				switch n := n.(type) {
				case lib.DirectoryNote:
					fmt.Fprintln(stdout, filepath.Join(path, n.Name)+string(os.PathSeparator))
				case lib.Note:
					fmt.Fprintln(stdout, n.Path)
				}
			}
			return 0
		} else {
			tty, err := openTTY()
			if err != nil {
				fmt.Fprintf(stdout, "Error opening the terminal: %v\n", err)
				return 1
			}
			defer tty.Close()

			// ask for filename to create
			fmt.Fprint(tty, "Enter filename: ")

			// Read from the TTY instead of os.Stdin
			reader := bufio.NewReader(tty)
			fileName, _ := reader.ReadString('\n')
			cleanFileName := utils.CleanString(fileName)

			namePath := filepath.Join(path, cleanFileName)
			code := saveNote(stdin, stdout, namePath, askMeta)
			fmt.Fprintf(stdout, "New note %s created successfully\n", namePath)
			return code
		}
	} else {
		if isTerminal() {
			// terminal without input - show file content
			fmt.Fprint(stdout, note.FullContent)
			return 0
		}

		code := saveNote(stdin, stdout, path, askMeta)
		fmt.Fprintf(stdout, "Note %s updated successfully\n", path)
		return code
	}
}

// open the terminal directly
func openTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

func isTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

func saveNote(stdin io.Reader, stdout io.Writer, path string, askMeta bool) int {
	stdinVal, error := io.ReadAll(stdin)
	if error != nil {
		fmt.Fprintf(stdout, "Error reading stdin: %v\n", error)
		return 1
	}
	content := string(stdinVal)
	storageService := lib.GetStorageService()

	// metadata could be received from stdin or asked from user
	meta, rest := lib.ParseMetadata(content)
	if meta.Title == "" || len(meta.Tags) == 0 {
		if askMeta {
			tty, err := openTTY()
			if err != nil {
				fmt.Fprintf(stdout, "Error opening the terminal: %v\n", err)
				return 1
			}
			defer tty.Close()

			if meta.Title == "" {
				// ask title
				fmt.Fprint(tty, "Enter title: ")

				// Read from the TTY instead of os.Stdin
				reader := bufio.NewReader(tty)
				titleInput, _ := reader.ReadString('\n')
				meta.Title = utils.CleanString(titleInput)
			}

			if len(meta.Tags) == 0 {
				// ask title
				fmt.Fprint(tty, "Enter tags separated by spaces: ")

				// Read from the TTY instead of os.Stdin
				reader := bufio.NewReader(tty)
				tagsInput, _ := reader.ReadString('\n')
				meta.Tags = strings.Fields(utils.CleanString(tagsInput))
			}
		} else {
			// don't want to ask metadata - get title from first line
			scanner := bufio.NewScanner(bytes.NewReader(rest))
			if scanner.Scan() {
				meta.Title = scanner.Text()
			}
		}
	}
	_, err := storageService.Store.SaveNote(path, rest, meta)
	if err != nil {
		return 1
	}
	return 0
}
