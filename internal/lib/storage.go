package lib

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"notecata/internal/config"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"text/template"
	"time"
)

type Store interface {
	Note(path string) (*Note, error)
	Notes(path string) ([]NoteOrDirectory, error)
	SaveNote(path string, content []byte, metaInput MetadataNote) (*Note, error)
	CreateDirectory(path string) error
	DeleteNote(path string) (*Note, error)
}

type storageService struct {
	Store Store
}

var (
	instance *storageService
	once     sync.Once
)

func GetStorageService() *storageService {
	once.Do(func() {
		config := config.GetConfig()
		var store Store
		if config.Storage == "file" {
			store = &FileStore{}
		} else {
			panic("not implemented")
		}
		instance = &storageService{Store: store}
	})
	return instance
}

type FileStore struct{}

var ErrPathIsDirectory = errors.New("by note path found a directory")

func (s *FileStore) getFullPath(path string) string {
	config := config.GetConfig()
	return filepath.Join(config.Root, path)
}

func (s *FileStore) checkEmptyPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("path must not be empty")
	}
	return nil
}

func (s *FileStore) Note(path string) (*Note, error) {
	fileOrDir, err := os.Open(s.getFullPath(path))
	if err != nil {
		return nil, err
	}
	defer fileOrDir.Close()

	fileInfo, err := fileOrDir.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, ErrPathIsDirectory
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(fileOrDir)
	content := buf.String()
	meta, rest := ParseMetadata(content)

	var note = Note{
		Path:        path,
		FullContent: content,
		Content:     rest,
		Metadata:    meta,
	}

	return &note, nil
}

func (s *FileStore) Notes(path string) ([]NoteOrDirectory, error) {
	fullPath := s.getFullPath(path)
	files, err := os.ReadDir(fullPath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	notes := []NoteOrDirectory{}
	for _, file := range files {
		if file.IsDir() {
			notes = append(notes, DirectoryNote{
				Path: path,
				Name: file.Name(),
			})
		} else {
			fileName := file.Name()
			if !slices.Contains([]string{".gitignore"}, fileName) {
				note, err := s.Note(filepath.Join(path, fileName))
				if err != nil {
					return nil, err
				}
				notes = append(notes, *note)
			}
		}
	}

	return notes, nil
}

func (s *FileStore) SaveNote(path string, content []byte, meta MetadataNote) (*Note, error) {
	fullPath := s.getFullPath(path)

	if meta.Format == "" || !slices.Contains(
		[]string{"Markdown", "reStructuredText"},
		meta.Format) {
		meta.Format = "Markdown"
	}
	meta.Date = time.Now()

	var out []byte
	var tmplFile = "internal/templates/frontmatter/yaml.tmpl"
	tmpl, err := template.New("yaml.tmpl").ParseFiles(tmplFile)
	if err != nil {
		log.Printf("Error parsing template: %v\n", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, meta)
	if err != nil {
		log.Printf("Error rendering template: %v\n", err)
	}

	buf.Write(content)
	out = buf.Bytes()

	dir, _ := filepath.Split(fullPath)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory %s: %v\n", dir, err)
		return nil, err
	}

	err = os.WriteFile(fullPath, out, 0644)
	if err != nil {
		log.Printf("Can't write to file %s, %s\n", fullPath, err)
		return nil, err
	}

	var note = Note{
		Path:        path,
		FullContent: buf.String(),
		Content:     content,
		Metadata:    meta,
	}

	return &note, nil
}

func (s *FileStore) CreateDirectory(path string) error {
	err := s.checkEmptyPath(path)
	if err != nil {
		log.Printf("Error creating directory %s: %v\n", path, err)
		return err
	}
	fullPath := s.getFullPath(path)
	err = os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory %s: %v\n", fullPath, err)
		return err
	}
	return nil
}

func (s *FileStore) DeleteNote(path string) (*Note, error) {
	err := s.checkEmptyPath(path)
	if err != nil {
		log.Printf("Error removing directory %s: %v\n", path, err)
		return nil, err
	}
	fullPath := s.getFullPath(path)
	note, err := s.Note(path)
	if errors.Is(err, ErrPathIsDirectory) {
		fmt.Printf("%s is a directory\n", path)

		// delete directory
		err = os.RemoveAll(fullPath)
		if err != nil {
			fmt.Printf("Error removing directory: %v\n", err)
			return nil, err
		}
		fmt.Printf("Directory %s deleted successfully\n", path)
	} else {
		// delete file
		err = os.Remove(fullPath)
		if err != nil {
			fmt.Printf("Error removing note: %v\n", err)
			return nil, err
		}
		fmt.Printf("Note %s deleted successfully\n", path)
	}
	return note, nil
}
