package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type Note struct {
	Title   string
	Tags    []string
	Content string
	Created time.Time
}

type AppConfig struct {
	RepoPath string
}

// ensureRepoExists checks if the repo exists and is a git repo
func ensureRepoExists(path string) (*git.Repository, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create directory
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}
		
		// Initialize git repository
		repo, err := git.PlainInit(path, false)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize git repository: %v", err)
		}
		
		return repo, nil
	}
	
	// Open existing repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("not a valid git repository: %v", err)
	}
	
	return repo, nil
}

// saveNote saves a note to the repository
func saveNote(note Note, repo *git.Repository, repoPath string) error {
	// Format the filename: YYYY-MM-DD-title.md
	fileName := fmt.Sprintf("%04d-%02d-%02d-%s.md", 
		note.Created.Year(), 
		note.Created.Month(), 
		note.Created.Day(), 
		strings.ReplaceAll(note.Title, " ", "-"))
	
	// Create the file content
	content := fmt.Sprintf("# %s\n\nTags: %s\n\n%s", 
		note.Title, 
		strings.Join(note.Tags, ", "), 
		note.Content)
	
	// Write to file
	filePath := filepath.Join(repoPath, fileName)
	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	
	// Get the worktree
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}
	
	// Add file to git
	_, err = w.Add(fileName)
	if err != nil {
		return fmt.Errorf("git add failed: %v", err)
	}
	
	// Commit changes
	commitMsg := fmt.Sprintf("Add note: %s", note.Title)
	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Notes App",
			Email: "notes@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("git commit failed: %v", err)
	}
	
	return nil
}

// listNotes retrieves all notes from the repository
func listNotes(repoPath string) ([]Note, error) {
	var notes []Note
	
	// Get all .md files
	files, err := filepath.Glob(filepath.Join(repoPath, "*.md"))
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}
	
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}
		
		// Parse note from file
		note, err := parseNoteFromContent(content, filepath.Base(file))
		if err != nil {
			continue
		}
		
		notes = append(notes, note)
	}
	
	return notes, nil
}

// parseNoteFromContent extracts note data from file content
func parseNoteFromContent(content []byte, filename string) (Note, error) {
	var note Note
	
	// Parse creation date and title from filename (YYYY-MM-DD-title.md)
	parts := strings.Split(filename, "-")
	if len(parts) < 4 {
		return note, fmt.Errorf("invalid filename format")
	}
	
	year := parts[0]
	month := parts[1]
	day := parts[2]
	
	// Extract title (join remaining parts and remove .md)
	titleParts := parts[3:]
	title := strings.Join(titleParts, "-")
	title = strings.TrimSuffix(title, ".md")
	title = strings.ReplaceAll(title, "-", " ")
	
	// Parse date
	dateStr := fmt.Sprintf("%s-%s-%s", year, month, day)
	created, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return note, fmt.Errorf("invalid date format: %v", err)
	}
	
	// Parse content
	contentStr := string(content)
	scanner := bufio.NewScanner(strings.NewReader(contentStr))
	
	// First line should be title
	if scanner.Scan() {
		titleLine := scanner.Text()
		if strings.HasPrefix(titleLine, "# ") {
			note.Title = strings.TrimPrefix(titleLine, "# ")
		}
	}
	
	// Look for tags
	var contentBuilder strings.Builder
	foundTags := false
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if !foundTags && strings.HasPrefix(line, "Tags: ") {
			tagsStr := strings.TrimPrefix(line, "Tags: ")
			tags := strings.Split(tagsStr, ", ")
			note.Tags = tags
			foundTags = true
			continue
		}
		
		// Add to content
		contentBuilder.WriteString(line)
		contentBuilder.WriteString("\n")
	}
	
	note.Content = contentBuilder.String()
	note.Created = created
	
	if note.Title == "" {
		note.Title = title // Use filename-derived title if not found in content
	}
	
	return note, nil
}

// pushToRemote pushes changes to remote repository
func pushToRemote(repo *git.Repository) error {
	// Push using go-git
	err := repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	
	if err != nil && err != transport.ErrEmptyRemoteRepository {
		return fmt.Errorf("git push failed: %v", err)
	}
	
	return nil
}

// pullFromRemote pulls changes from remote repository
func pullFromRemote(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}
	
	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("git pull failed: %v", err)
	}
	
	return nil
}

func main() {
	// Configuration
	homeDir, _ := os.UserHomeDir()
	config := AppConfig{
		RepoPath: filepath.Join(homeDir, "notes-repo"),
	}
	
	// Ensure repository exists
	repo, err := ensureRepoExists(config.RepoPath)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		os.Exit(1)
	}
	
	// Create Fyne app
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Notes Manager")
	w.Resize(fyne.NewSize(800, 600))
	
	// UI elements
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Note Title")
	
	tagsEntry := widget.NewEntry()
	tagsEntry.SetPlaceHolder("Tags (comma separated)")
	
	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetPlaceHolder("Write your note content here (Markdown supported)")
	
	notesList := widget.NewList(
		func() int { return 0 }, // Will be updated when we load notes
		func() fyne.CanvasObject {
			return widget.NewLabel("Note Title")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// Will be updated when we load notes
		},
	)
	
	// Load notes initially
	var notes []Note
	
	refreshNotesList := func() {
		notes, err = listNotes(config.RepoPath)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		
		notesList.Length = func() int {
			return len(notes)
		}
		
		notesList.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < len(notes) {
				label.SetText(notes[id].Title)
			}
		}
		
		notesList.Refresh()
	}
	
	// Buttons
	saveButton := widget.NewButtonWithIcon("Save Note", theme.DocumentSaveIcon(), func() {
		if titleEntry.Text == "" {
			dialog.ShowInformation("Error", "Title cannot be empty", w)
			return
		}
		
		// Create note
		note := Note{
			Title:   titleEntry.Text,
			Content: contentEntry.Text,
			Created: time.Now(),
		}
		
		// Parse tags
		if tagsEntry.Text != "" {
			tagsList := strings.Split(tagsEntry.Text, ",")
			for i, tag := range tagsList {
				tagsList[i] = strings.TrimSpace(tag)
			}
			note.Tags = tagsList
		}
		
		// Save note
		if err := saveNote(note, repo, config.RepoPath); err != nil {
			dialog.ShowError(err, w)
			return
		}
		
		// Clear fields
		titleEntry.SetText("")
		tagsEntry.SetText("")
		contentEntry.SetText("")
		
		// Refresh list
		refreshNotesList()
		
		dialog.ShowInformation("Success", "Note saved successfully", w)
	})
	
	pushButton := widget.NewButtonWithIcon("Push to Remote", theme.UploadIcon(), func() {
		// Push to remote repository
		if err := pushToRemote(repo); err != nil {
			dialog.ShowError(err, w)
			return
		}
		
		dialog.ShowInformation("Success", "Changes pushed to remote repository", w)
	})
	
	pullButton := widget.NewButtonWithIcon("Pull from Remote", theme.DownloadIcon(), func() {
		// Pull from remote repository
		if err := pullFromRemote(repo); err != nil {
			dialog.ShowError(err, w)
			return
		}
		
		// Refresh list
		refreshNotesList()
		
		dialog.ShowInformation("Success", "Changes pulled from remote repository", w)
	})
	
	newButton := widget.NewButtonWithIcon("New Note", theme.FileIcon(), func() {
		// Clear fields
		titleEntry.SetText("")
		tagsEntry.SetText("")
		contentEntry.SetText("")
	})
	
	// Note selection handling
	notesList.OnSelected = func(id widget.ListItemID) {
		if id < len(notes) {
			selectedNote := notes[id]
			titleEntry.SetText(selectedNote.Title)
			tagsEntry.SetText(strings.Join(selectedNote.Tags, ", "))
			contentEntry.SetText(selectedNote.Content)
		}
	}
	
	// Layout
	editorContainer := container.NewVBox(
		widget.NewLabel("Title:"),
		titleEntry,
		widget.NewLabel("Tags:"),
		tagsEntry,
		widget.NewLabel("Content:"),
		container.NewScroll(contentEntry),
		container.NewHBox(
			saveButton,
			newButton,
		),
	)
	
	listContainer := container.NewVBox(
		widget.NewLabel("Notes:"),
		container.NewScroll(notesList),
		container.NewHBox(
			pushButton,
			pullButton,
		),
	)
	
	// Split view
	split := container.NewHSplit(
		listContainer,
		editorContainer,
	)
	split.SetOffset(0.3) // 30% for list, 70% for editor
	
	// Set main container
	w.SetContent(container.New(layout.NewMaxLayout(), split))
	
	// Initial refresh
	refreshNotesList()
	
	// Show and run
	w.ShowAndRun()
}