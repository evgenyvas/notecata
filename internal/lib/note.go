// Package lib
package lib

import (
	"log"
	"slices"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
)

type NoteOrDirectory interface {
	Type() string
}

type Note struct {
	Path        string
	FullContent string
	Content     []byte
	Metadata    MetadataNote
}

func (m Note) Type() string {
	return "note"
}

type DirectoryNote struct {
	Path string
	Name string
}

func (m DirectoryNote) Type() string {
	return "directory"
}

type MetadataNote struct {
	Format string
	Date   time.Time
	Title  string
	Tags   []string
}

func ParseMetadata(input string) (meta MetadataNote, rest []byte) {
	model := new(FrontmatterMeta)
	rest, err := frontmatter.Parse(strings.NewReader(input), model)
	if err != nil {
		log.Fatalf("Error parsing metadata: %v\n", err)
		panic("metadata")
	}
	meta = MetadataNote{
		Format: model.Format,
		Title:  model.Title,
		Tags:   model.Tags,
	}
	if meta.Format == "" || !slices.Contains(
		[]string{"Markdown", "reStructuredText"},
		meta.Format) {
		meta.Format = "Markdown"
	}

	if model.Date != "" {
		t, err := time.Parse("2006-01-02 15:04:05", model.Date)
		if err != nil {
			log.Println("Error parsing date:", err)
		} else {
			meta.Date = t
		}
	}

	return meta, rest
}
