package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

// loadConfig loads the configuration from .git-notes.conf file
func loadConfig() (AppConfig, error) {
	config := AppConfig{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("failed to get user home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".git-notes.conf")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, create it
		return createConfigFile(homeDir, configPath)
	}
	
	// Read the config file
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}
	
	// Parse config file
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "REPO_PATH=") {
			config.RepoPath = strings.TrimPrefix(line, "REPO_PATH=")
		}
	}
	
	// Validate config
	if config.RepoPath == "" {
		return config, fmt.Errorf("repository path not found in config file")
	}
	
	return config, nil
}

// createConfigFile creates a new configuration file with user input
func createConfigFile(homeDir, configPath string) (AppConfig, error) {
	config := AppConfig{}
	
	// Default repository path
	defaultRepoPath := filepath.Join(homeDir, "notes-repo")
	
	// Ask user for repository path using dialog
	a := app.New()
	w := a.NewWindow("Git Notes Configuration")
	w.Resize(fyne.NewSize(400, 200))
	
	// Entry for repository path
	repoPathEntry := widget.NewEntry()
	repoPathEntry.SetText(defaultRepoPath)
	
	// Form for the dialog
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Repository Path:", Widget: repoPathEntry},
		},
	}
	
	// Channel to get the result
	done := make(chan bool)
	
	// Create and show dialog
	dialog.ShowCustomConfirm("Git Notes Configuration", "Save", "Cancel", form, func(save bool) {
		if save {
			config.RepoPath = repoPathEntry.Text
			
			// Write to config file
			configContent := fmt.Sprintf("REPO_PATH=%s\n", config.RepoPath)
			err := ioutil.WriteFile(configPath, []byte(configContent), 0644)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to write config file: %v", err), w)
			}
		} else {
			// Use default if canceled
			config.RepoPath = defaultRepoPath
			
			// Write default to config file
			configContent := fmt.Sprintf("REPO_PATH=%s\n", config.RepoPath)
			err := ioutil.WriteFile(configPath, []byte(configContent), 0644)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to write config file: %v", err), w)
			}
		}
		done <- true
	}, w)
	
	<-done // Wait for dialog to complete
	w.Close()
	
	return config, nil
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

// sortNotesByDateAndTitle sorts notes first by creation date (newest first)
// and then by title (alphabetically) for notes with the same date
func sortNotesByDateAndTitle(notes []Note) {
	sort.Slice(notes, func(i, j int) bool {
	    // First compare by date (newest first)
	    if !notes[i].Created.Equal(notes[j].Created) {
		return notes[i].Created.After(notes[j].Created)
	    }
	    // If dates are equal, compare by title (alphabetically)
	    return strings.ToLower(notes[i].Title) < strings.ToLower(notes[j].Title)
	})    
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
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
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
	w.Resize(fyne.NewSize(900, 700))
	
	// UI elements
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Note Title")
	
	tagsEntry := widget.NewEntry()
	tagsEntry.SetPlaceHolder("Tags (comma separated)")
	
	// Create multiline content entry with proper scrolling
	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetPlaceHolder("Write your note content here (Markdown supported)")
	contentEntry.Wrapping = fyne.TextWrapWord  // Enable word wrapping
	
	// Content entry should take up all available space
	contentEntryScroll := container.NewScroll(contentEntry)
	contentEntryScroll.SetMinSize(fyne.NewSize(500, 400))  // Set minimum size for content area
	
	// Initialize notes slice
	var notes []Note
	
	// Create list widget with proper binding to notes slice
	notesList := widget.NewList(
		func() int { 
			return len(notes) // This will update when notes slice changes
		},
		func() fyne.CanvasObject {
			// Make the template item more visible with proper styling
			return container.NewHBox(
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(notes) {
				// Get the label from the container
				label := obj.(*fyne.Container).Objects[0].(*widget.Label)
				label.SetText(notes[id].Title)
				// Make text bold and properly styled
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.Alignment = fyne.TextAlignLeading
			}
		},
	)
	
	// Custom item size to make list items taller and more visible
	notesList.OnSelected = func(id widget.ListItemID) {
		if id < len(notes) {
			selectedNote := notes[id]
			titleEntry.SetText(selectedNote.Title)
			tagsEntry.SetText(strings.Join(selectedNote.Tags, ", "))
			contentEntry.SetText(selectedNote.Content)
		}
	}
	
	// Function to refresh the notes list
	refreshNotesList := func() {
		var err error
		notes, err = listNotes(config.RepoPath)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		
		// Sort notes by creation time (newest first)
		sortNotesByDateAndTitle(notes)
		
		// Fully refresh the list widget
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
		
		// Refresh list (including sorting)
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
		
		// Refresh list (including sorting)
		refreshNotesList()
		
		dialog.ShowInformation("Success", "Changes pulled from remote repository", w)
	})
	
	newButton := widget.NewButtonWithIcon("New Note", theme.FileIcon(), func() {
		// Clear fields
		titleEntry.SetText("")
		tagsEntry.SetText("")
		contentEntry.SetText("")
	})
	
	// Layout
	// Create a form layout for title and tags
	formContainer := container.NewVBox(
		widget.NewLabel("Title:"),
		titleEntry,
		widget.NewLabel("Tags:"),
		tagsEntry,
	)
	
	// Content area with label
	contentContainer := container.NewVBox(
		widget.NewLabel("Content:"),
		contentEntryScroll,  // Use the scrollable container
	)
	
	// Buttons container
	buttonContainer := container.NewHBox(
		saveButton,
		newButton,
	)
	
	// Stack everything in the editor area
	editorContainer := container.NewBorder(
		formContainer,  // Top
		buttonContainer, // Bottom
		nil,            // Left
		nil,            // Right
		contentContainer, // Center (fills remaining space)
	)
	
	// List panel header
	listHeader := widget.NewLabelWithStyle("Notes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	// Git operation buttons in the list panel
	gitButtonsContainer := container.NewHBox(
		pushButton,
		pullButton,
	)
	
	// Make the notes list fill all available space
	listContent := container.NewBorder(
		listHeader,       // Top
		gitButtonsContainer, // Bottom
		nil,              // Left
		nil,              // Right
		notesList,        // Center (fills remaining space)
	)
	
	// Set minimum size for list container
	// The list panel should take about 25% of the window width, but at least 200px
	split := container.NewHSplit(
		listContent,
		editorContainer,
	)
	split.SetOffset(0.25) // 25% for list, 75% for editor
	
	// Set main container
	w.SetContent(split)
	
	// Initial refresh (including sorting)
	refreshNotesList()
	
	// Show and run
	w.ShowAndRun()
}