# Application created using claude chat

The application was created from scratch by claude 3.7 sonet.

The process was intentionally "hands off" only general indications were given.  No direct indications about how to modify the code.

I Only a did a couple of minor changes manually to fix compilation errors, mostly to save time and context space.

I did not review the code and based my feedback only in the behavior.


## observations

1. The application worked since almost the first version!
2. The application included functions I did not explicitly asked for, such as updating existing notes.
3. Claude introduced many unsolicited changes
4. I asked to fix an issue with the list of notes not being updated, but the list was actually updated but the space did not allow to show the new note(s). Claude "fixed" the issue and "hallucinated" [an explanation](CHAT.md#fix-note-list-is-not-updated):

```
I've reviewed the code and see the issue with the list of notes not updating properly after a note is added. Let me analyze the problem and provide a solution.

The issue is in the refreshNotesList function. Here's what's happening:

When the app loads, the refreshNotesList function is called once to initially populate the list.
This function sets up the Length and UpdateItem functions for the notesList widget and calls notesList.Refresh().
When a note is saved, the function is called again to refresh the list, but there's a problem with how the list is being updated.
The main issue is that while you're updating the notes slice correctly with the new notes from the repository, and you're setting the Length and UpdateItem functions properly, there's no explicit mechanism to tell the list that its data has changed significantly (beyond just refreshing the UI).
```


The process went thru several iterations. 

1. [Initial implementation](#initial-version)
2. [Display the content of the note with multiple lines](#second-iteration-with-multi-line-text-entry)
3. [Refresh the list of notes after inserting a note](#third-iteration-with-list-refresh)
4. [Display the full list of notes](#fourth-iteration-with-full-note-list-panel)
5. [Sort notes by creation date](#v5-sort-note-list)
5.1 Second attempt to sort, didn't work (apparently, the previous version did work, but the ordering wasn't obvious and I didn't realized)

At this point the application was capable of:
- add notes
- update notes



The screen shots bellow show the progress.

## Screenshots

### initial-version

![main view](./screenshots/v1/main-view.png)

![note entry](./screenshots/v1/note-entry.png)

![notes list](./screenshots/v1/note-list.png)

### second-iteration-with-multi-line-text-entry

![main view](./screenshots/v2/main-view.png)

![note entry](./screenshots/v2/note-entry.png)


### third-iteration-with-list-refresh

![main view](./screenshots/v3/main-view.png)

![note entry](./screenshots/v3/note-entry.png)

### fourth-iteration-with-full-note-list-panel

![main view](./screenshots/v4/main-view.png)

![list updated](./screenshots/v4/list-updated.png)

### v5-sort-note-list

![main view](./screenshots/v5/main-view.png)


