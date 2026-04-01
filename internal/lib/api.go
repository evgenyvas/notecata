package lib

import (
	"os"
)

type APIStatus struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type APINoteOrDirectory interface {
	APIType() string
}

type APINote struct {
	Path    string   `json:"path"`
	Format  string   `json:"format"`
	Date    string   `json:"date"`
	Title   string   `json:"title"`
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
	Type    string   `json:"type"`
}

func (m APINote) APIType() string {
	return "note"
}

type APIDirectory struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (m APIDirectory) APIType() string {
	return "directory"
}

type APINoteSingle struct {
	APIStatus
	APINote
}

type APINoteList struct {
	APIStatus
	Notes []APINoteOrDirectory `json:"notes"`
}

type APINotePost struct {
	Format  string   `json:"format"`
	Title   string   `json:"title"`
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
}

func ToAPI(n NoteOrDirectory) APINoteOrDirectory {
	switch n := n.(type) {
	case Note:
		return APINote{
			Format:  n.Metadata.Format,
			Date:    n.Metadata.Date.Format("2006-01-02 15:04:05"),
			Title:   n.Metadata.Title,
			Tags:    n.Metadata.Tags,
			Path:    n.Path,
			Content: string(n.Content),
			Type:    n.Type(),
		}
	case DirectoryNote:
		return APIDirectory{
			Path: n.Path + string(os.PathSeparator),
			Name: n.Name,
			Type: n.Type(),
		}
	default:
		return nil
	}
}
