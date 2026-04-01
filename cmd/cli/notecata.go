package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/mattn/go-isatty"
	"io"
	"notecata/internal/lib"
	"notecata/utils"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	args := os.Args[1:]

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
		fmt.Println("error: missing argument")
		os.Exit(1)
	}
	path := args[0]

	storageService := lib.GetStorageService()
	note, err := storageService.Store.Note(path)
	if err != nil && !errors.Is(err, lib.ErrPathIsDirectory) {
		if isTerminal() {
			fmt.Printf("Can't open %s, %s\n", path, err)
			os.Exit(1)
		} else {
			saveNote(path, askMeta)
			fmt.Printf("New note %s created successfully\n", path)
			os.Exit(0)
		}
	}

	if isDelete {
		_, err = storageService.Store.DeleteNote(path)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if errors.Is(err, lib.ErrPathIsDirectory) {
		fmt.Printf("%s is a directory\n", path)

		if isTerminal() {
			// show list of contents
			notes, err := storageService.Store.Notes(path)
			if err != nil {
				os.Exit(1)
			}
			if len(notes) > 0 {
				fmt.Println("\nContents:")
			} else {
				fmt.Println("\nDirectory is empty:")
			}
			for _, n := range notes {
				switch n := n.(type) {
				case lib.DirectoryNote:
					fmt.Println(filepath.Join(path, n.Name) + string(os.PathSeparator))
				case lib.Note:
					fmt.Println(n.Path)
				}
			}
			os.Exit(0)
		} else {
			tty := openTTY()
			defer tty.Close()

			// ask for filename to create
			fmt.Fprint(tty, "Enter filename: ")

			// Read from the TTY instead of os.Stdin
			reader := bufio.NewReader(tty)
			fileName, _ := reader.ReadString('\n')
			cleanFileName := utils.CleanString(fileName)

			namePath := filepath.Join(path, cleanFileName)
			saveNote(namePath, askMeta)
			fmt.Printf("New note %s created successfully\n", namePath)
			os.Exit(0)
		}
	} else {
		if isTerminal() {
			// terminal without input - show file content
			fmt.Print(note.FullContent)
			os.Exit(0)
		}

		saveNote(path, askMeta)
		fmt.Printf("Note %s updated successfully\n", path)
	}
}

// open the terminal directly
func openTTY() *os.File {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Printf("Error opening the terminal: %v\n", err)
		os.Exit(1)
	}
	return tty
}

func isTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

func readStdin() []byte {
	stdin, error := io.ReadAll(os.Stdin)
	if error != nil {
		fmt.Printf("Error reading stdin: %v\n", error)
		os.Exit(1)
	}
	return stdin
}

func saveNote(path string, askMeta bool) {
	stdin := readStdin()
	content := string(stdin)
	storageService := lib.GetStorageService()

	// metadata could be received from stdin or asked from user
	meta, rest := lib.ParseMetadata(content)
	if meta.Title == "" || len(meta.Tags) == 0 {
		if askMeta {
			tty := openTTY()
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
		os.Exit(1)
	}
}
