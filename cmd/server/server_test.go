package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"notecata/internal/lib"
	"strings"
	"testing"
	"time"
)

type mockStore struct{}

func (m *mockStore) Note(path string) (*lib.Note, error) {
	content := `
---
format: "Markdown"
date: "2026-05-23 00:57:15"
title: "Test title"
tags: ["test", "mock"]
---
#Test note

note body
`

	meta, rest := lib.ParseMetadata(content)
	return &lib.Note{
		Path:        path,
		FullContent: content,
		Content:     rest,
		Metadata:    meta,
	}, nil
}

func (m *mockStore) Notes(path string) ([]lib.NoteOrDirectory, error) {
	notes := []lib.NoteOrDirectory{}
	notes = append(notes, lib.DirectoryNote{
		Path: path,
		Name: "testdir",
	})
	note, err := m.Note("doc.md")
	if err != nil {
		return nil, err
	}
	notes = append(notes, *note)
	return notes, nil
}

func (m *mockStore) SaveNote(path string, content []byte, meta lib.MetadataNote) (*lib.Note, error) {
	meta.Date, _ = time.Parse("2006-01-02 15:04:05", "2026-05-23 00:57:15")
	return &lib.Note{
		Path:        path,
		FullContent: string(content),
		Content:     content,
		Metadata:    meta,
	}, nil
}

func (m *mockStore) CreateDirectory(path string) error {
	return nil
}

func (m *mockStore) DeleteNote(path string) (*lib.Note, error) {
	return m.Note(path)
}

func TestApiGetNotes(t *testing.T) {
	api := &API{
		Store: &mockStore{},
	}

	req := httptest.NewRequest(http.MethodGet, "/noteslist/docs", nil)
	req.SetPathValue("path", "docs")
	rec := httptest.NewRecorder()

	api.getNotes(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", res.Status)
	}

	body, _ := io.ReadAll(res.Body)

	expected := `{` +
		`"status":200,` +
		`"message":"ok",` +
		`"notes":[` +
		`{` +
		`"path":"docs/",` +
		`"name":"testdir",` +
		`"type":"directory"` +
		`},{` +
		`"path":"doc.md",` +
		`"format":"Markdown",` +
		`"date":"2026-05-23 00:57:15",` +
		`"title":"Test title",` +
		`"tags":["test","mock"],` +
		`"content":"#Test note\n\nnote body\n",` +
		`"type":"note"` +
		`}` +
		`]` +
		`}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("expected %s; got %s", expected, body)
	}
}

func TestApiGetNote(t *testing.T) {
	api := &API{
		Store: &mockStore{},
	}

	req := httptest.NewRequest(http.MethodGet, "/notes/docs.md", nil)
	req.SetPathValue("path", "docs.md")
	rec := httptest.NewRecorder()

	api.getNote(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", res.Status)
	}

	body, _ := io.ReadAll(res.Body)

	expected := `{` +
		`"status":200,` +
		`"message":"ok",` +
		`"path":"docs.md",` +
		`"format":"Markdown",` +
		`"date":"2026-05-23 00:57:15",` +
		`"title":"Test title",` +
		`"tags":["test","mock"],` +
		`"content":"#Test note\n\nnote body\n",` +
		`"type":"note"` +
		`}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("expected %s; got %s", expected, body)
	}
}

func TestApiSaveNote(t *testing.T) {
	api := &API{
		Store: &mockStore{},
	}

	payload := lib.APINotePost{
		Format:  "Markdown",
		Title:   "api test",
		Tags:    []string{"go", "yaml", "json"},
		Content: "Note create content",
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/notes/docs.md", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("path", "docs.md")
	rec := httptest.NewRecorder()

	api.saveNote(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", res.Status)
	}

	body, _ := io.ReadAll(res.Body)

	expected := `{` +
		`"status":200,` +
		`"message":"Note docs.md saved successfully",` +
		`"path":"docs.md",` +
		`"format":"Markdown",` +
		`"date":"2026-05-23 00:57:15",` +
		`"title":"api test",` +
		`"tags":["go","yaml","json"],` +
		`"content":"Note create content",` +
		`"type":"note"` +
		`}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("expected %s; got %s", expected, body)
	}
}

func TestApiDeleteNote(t *testing.T) {
	api := &API{
		Store: &mockStore{},
	}

	req := httptest.NewRequest(http.MethodGet, "/notes/docs.md", nil)
	req.SetPathValue("path", "docs.md")
	rec := httptest.NewRecorder()

	api.deleteNote(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", res.Status)
	}

	body, _ := io.ReadAll(res.Body)

	expected := `{` +
		`"status":200,` +
		`"message":"Note docs.md deleted successfully"` +
		`}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("expected %s; got %s", expected, body)
	}
}
