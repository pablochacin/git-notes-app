# Chat with claude

## Initial prompt

I want to develop a desktop application for linux that can create text notes and save them to a github repository. 

Each note will have a title, some tags (list of words that categorize the content) and a content. 

Each note will be saved in the repository as a separated markdown file. The file name is the date of creation and the title concatenated with any space character in the title will be replaced by a - character.  For example the note with title "note example" created on 2023-03-02 will be saved as "2025-03-02-note-example.md"

The content will be edited in the application as an markdown document. simple text editing is sufficient, no need for pre-visualization of content, but that would be a nice to have feature. 

Additionally, I want to be able to list existing documents in the git repository.

The programming language should go, preferably. The code should be portable across linux distruibutions. But if this is not possible, ubuntu should be preferable.

## claude response

I'll help you develop a desktop application for Linux that manages text notes and syncs them with a GitHub repository. Let's create a Go-based solution that will be portable across Linux distributions.

Go Note-Taking App with GitHub Integration

This is a complete note-taking application written in Go with GitHub integration. Here's an overview of its features:

[code](main.go)

### Core Note Management:

1. Create notes with title, tags, and markdown content
* Save notes with automatic date-based filenames (YYYY-MM-DD-title.md)
* View and edit existing notes

2. GitHub Integration:
* Local git repository management
* Push changes to GitHub
* Pull changes from GitHub

3. User Interface:
* Clean, split-panel design
* List of existing notes on the left
* Note editor on the right
* Markdown content editing

### Setup & Dependencies
To run this application, you'll need:

Go installed on your system
The Fyne UI toolkit: go get fyne.io/fyne/v2
The go-git package: go get github.com/go-git/go-git/v5
Git installed on your system
To build the application:

bash

```
go mod init note-manager
go get fyne.io/fyne/v2
go get github.com/go-git/go-git/v5
go build
```

### How It Works
The app creates/uses a git repository in your home directory (~/notes-repo)
Each note is saved as a separate markdown file
Files are automatically committed to the local git repository
You can manually push/pull to sync with GitHub
Would you like me to explain any specific part of the code in more detail or make any modifications to the application?

