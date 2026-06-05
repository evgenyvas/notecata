package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"notecata/internal/config"
	"notecata/internal/lib"
	"notecata/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
)

type API struct {
	Store lib.Store
	//Logger     Logger
	//Config     Config
	//Mailer     Mailer
	//Cache      Cache
}

func (api *API) getNotes(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	notes, err := api.Store.Notes(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiNotes := make([]lib.APINoteOrDirectory, len(notes))
	for i, note := range notes {
		apiNotes[i] = lib.ToAPI(note)
	}
	apiNotesList := lib.APINoteList{
		APIStatus: lib.APIStatus{
			Status:  http.StatusOK,
			Message: "ok",
		},
		Notes: apiNotes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiNotesList)
}

func (api *API) getNote(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	note, err := api.Store.Note(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiNote := lib.APINoteSingle{
		APIStatus: lib.APIStatus{
			Status:  http.StatusOK,
			Message: "ok",
		},
		APINote: lib.ToAPI(*note).(lib.APINote),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiNote)
}

func (api *API) saveNote(w http.ResponseWriter, r *http.Request) {
	var note lib.APINotePost
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metaInput := lib.MetadataNote{
		Format: note.Format,
		Title:  note.Title,
		Tags:   note.Tags,
	}
	path := r.PathValue("path")

	newNote, err := api.Store.SaveNote(path, []byte(note.Content), metaInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiNote := lib.APINoteSingle{
		APIStatus: lib.APIStatus{
			Status:  http.StatusOK,
			Message: fmt.Sprintf("Note %s saved successfully", path),
		},
		APINote: lib.ToAPI(*newNote).(lib.APINote),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiNote)
}

func (api *API) deleteNote(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	_, err := api.Store.DeleteNote(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiStatus := lib.APIStatus{
		Status:  http.StatusOK,
		Message: fmt.Sprintf("Note %s deleted successfully", path),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiStatus)
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	storageService := lib.GetStorageService()

	var prevPath = func() string {
		prevPath, _ := filepath.Split(path)
		return prevPath
	}

	// try to get single note
	note, error := storageService.Store.Note(path)
	if error != nil {
		if errors.Is(error, lib.ErrPathIsDirectory) {
			// get notes list in directory
			notes, err := storageService.Store.Notes(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			tmpl := template.New("noteslist")
			template.Must(tmpl.ParseFiles("internal/templates/base.html"))
			template.Must(tmpl.Funcs(template.FuncMap{
				"IsDirectory": func(i any) bool {
					_, ok := i.(lib.DirectoryNote)
					return ok
				},
				"PrevPath": prevPath,
				"PathSeparator": func() string {
					return string(os.PathSeparator)
				},
				"JoinPath": filepath.Join,
			}).ParseFiles("internal/templates/noteslist.html"))

			if err := tmpl.ExecuteTemplate(w, "base", struct {
				Path  string
				Notes []lib.NoteOrDirectory
			}{Path: path + string(os.PathSeparator), Notes: notes}); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		} else {
			http.Error(w, error.Error(), http.StatusBadRequest)
		}
		return
	}

	// render single note as markdown
	buf := new(bytes.Buffer)
	if err := goldmark.Convert(note.Content, buf); err != nil {
		http.Error(w, "Error converting markdown", http.StatusInternalServerError)
		return
	}

	tmpl := template.New("note")
	template.Must(tmpl.ParseFiles("internal/templates/base.html"))
	template.Must(tmpl.Funcs(template.FuncMap{
		"PrevPath": prevPath,
		"JoinPath": filepath.Join,
		"unsafe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseFiles("internal/templates/note.html"))

	if err := tmpl.ExecuteTemplate(w, "base", struct {
		Path     string
		Note     string
		Metadata lib.MetadataNote
	}{Path: path, Note: buf.String(), Metadata: note.Metadata}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func manageNote(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	// try to get single note
	storageService := lib.GetStorageService()
	note, error := storageService.Store.Note(path)
	var isAdd bool
	if error != nil {
		if errors.Is(error, lib.ErrPathIsDirectory) {
			isAdd = true
		} else {
			http.Error(w, error.Error(), http.StatusBadRequest)
			return
		}
	}

	if r.Method == http.MethodGet {
		tmpl := template.New("notemanage")
		template.Must(tmpl.ParseFiles("internal/templates/base.html"))
		template.Must(tmpl.Funcs(template.FuncMap{
			"JoinPath": filepath.Join,
			"PathSeparator": func() string {
				return string(os.PathSeparator)
			},
		}).ParseFiles("internal/templates/noteform.html"))

		if isAdd {
			path = path + string(os.PathSeparator)
		}
		if err := tmpl.ExecuteTemplate(w, "base", struct {
			Path  string
			Note  *lib.Note
			IsAdd bool
		}{Path: path, Note: note, IsAdd: isAdd}); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	r.ParseForm()
	nameInput := utils.KeepASCII(strings.TrimSpace(r.FormValue("name")))
	titleInput := strings.TrimSpace(r.FormValue("title"))
	tagsInput := strings.TrimSpace(r.FormValue("tags"))
	noteInput := strings.TrimSpace(r.FormValue("note"))

	if isAdd {
		if nameInput == "" {
			http.Error(w, "note name is empty", http.StatusBadRequest)
			return
		}
		if !strings.HasSuffix(nameInput, ".md") {
			nameInput = nameInput + ".md"
		}
	}

	if noteInput == "" {
		http.Error(w, "note content is empty", http.StatusBadRequest)
		return
	}

	if titleInput == "" {
		// get title from first line
		scanner := bufio.NewScanner(bytes.NewReader([]byte(noteInput)))
		if scanner.Scan() {
			titleInput = scanner.Text()
		}
	}

	var pathInput string
	var metaInput = lib.MetadataNote{}
	if isAdd {
		pathInput = filepath.Join(path, nameInput)
		metaInput = lib.MetadataNote{
			Title: titleInput,
			Tags:  strings.Fields(tagsInput),
		}
	} else {
		pathInput = path
		metaInput = note.Metadata
		metaInput.Title = titleInput
		metaInput.Tags = strings.Fields(tagsInput)
	}

	_, err := storageService.Store.SaveNote(pathInput, []byte(noteInput), metaInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, filepath.Join("/note/", path), http.StatusSeeOther)
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	storageService := lib.GetStorageService()
	_, err := storageService.Store.DeleteNote(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	prevPath, _ := filepath.Split(path)
	http.Redirect(w, r, filepath.Join("/note/", prevPath), http.StatusSeeOther)
}

func createDir(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	if r.Method == http.MethodGet {
		tmpl := template.New("createdir")
		template.Must(tmpl.ParseFiles("internal/templates/base.html"))
		template.Must(tmpl.Funcs(template.FuncMap{
			"JoinPath": filepath.Join,
			"PathSeparator": func() string {
				return string(os.PathSeparator)
			},
		}).ParseFiles("internal/templates/dirform.html"))

		if err := tmpl.ExecuteTemplate(w, "base", struct {
			Path string
		}{Path: path}); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	r.ParseForm()
	nameInput := utils.KeepASCII(strings.TrimSpace(r.FormValue("name")))
	dir := filepath.Join(path, nameInput)
	storageService := lib.GetStorageService()
	err := storageService.Store.CreateDirectory(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/note/"+dir, http.StatusSeeOther)
}

func main() {
	config := config.GetConfig()

	storage := lib.GetStorageService()
	api := &API{
		Store: storage.Store,
	}

	m := http.NewServeMux()
	m.HandleFunc("GET /api/noteslist/{path...}", api.getNotes)
	m.HandleFunc("GET /api/notes/{path...}", api.getNote)
	m.HandleFunc("POST /api/notes/{path...}", api.saveNote)
	m.HandleFunc("DELETE /api/notes/{path}", api.deleteNote)

	m.HandleFunc("/", getNotes)
	m.HandleFunc("/note/{path...}", getNotes)
	m.HandleFunc("/notemanage/{path...}", manageNote)
	m.HandleFunc("/notedelete/{path...}", deleteNote)
	m.HandleFunc("/createdir/{path...}", createDir)

	// Serve files from the "./static" directory at the "/static/" URL path
	fileServer := http.FileServer(http.Dir("./static"))
	m.Handle("/static/", http.StripPrefix("/static", fileServer))

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), m))
}
