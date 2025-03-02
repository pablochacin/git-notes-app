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
		return widget.NewLabel("Note Title")
	    },
	    func(id widget.ListItemID, obj fyne.CanvasObject) {
		label := obj.(*widget.Label)
		if id < len(notes) {
		    label.SetText(notes[id].Title)
		}
	    },
	)
	
	// Function to refresh the notes list
	refreshNotesList := func() {
	    var err error
	    notes, err = listNotes(config.RepoPath)
	    if err != nil {
		dialog.ShowError(err, w)
		return
	    }
	    
	    // This is the key change - fully refresh the list widget
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
	
	// List container with fixed width
	listContainer := container.NewVBox(
	    widget.NewLabel("Notes:"),
	    container.NewScroll(notesList),
	    container.NewHBox(
		pushButton,
		pullButton,
	    ),
	)
	
	// Set minimum size for list container
	listScroll := container.NewScroll(listContainer)
	listScroll.SetMinSize(fyne.NewSize(200, 0))
	
	// Split view
	split := container.NewHSplit(
	    listScroll,
	    editorContainer,
	)
	split.SetOffset(0.25) // 25% for list, 75% for editor
	
	// Set main container
	w.SetContent(container.New(layout.NewMaxLayout(), split))
	
	// Initial refresh
	refreshNotesList()
	
	// Show and run
	w.ShowAndRun()
    }
    