// Package journal holds the diary logic: which note is "today", and how to
// read and overwrite it inside a tag. Standard Notes access sits behind Store.
package journal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

// RolloverHour is when a new diary day starts; writing at 1 AM still counts
// as the previous day.
const RolloverHour = 4

var (
	ErrTagNotFound = errors.New("tag not found")
	ErrAmbiguous   = errors.New("ambiguous match")
)

// Store loads a freshly synced view of the account and persists changes.
type Store interface {
	Load() (items.Tags, items.Notes, error)
	Save(items.Notes, items.Tags) error
}

// Title returns the note title for the diary day containing t,
// e.g. "Sunday, November 1, '26".
func Title(t time.Time) string {
	return t.Add(-RolloverHour * time.Hour).Format("Monday, January 2, '06")
}

// Pull returns the text of the note titled noteTitle in tag tagTitle.
// found is false when the tag has no such note.
func Pull(s Store, tagTitle, noteTitle string) (text string, found bool, err error) {
	tags, notes, err := s.Load()
	if err != nil {
		return "", false, err
	}
	_, n, err := find(tags, notes, tagTitle, noteTitle)
	if err != nil || n == nil {
		return "", false, err
	}
	return n.Content.Text, true, nil
}

// Push overwrites the note's text, creating it in the tag if it does not
// exist. created is used only when creating; now becomes the modified time.
func Push(s Store, tagTitle, noteTitle, text string, created, now time.Time) error {
	tags, notes, err := s.Load()
	if err != nil {
		return err
	}
	tg, n, err := find(tags, notes, tagTitle, noteTitle)
	if err != nil {
		return err
	}

	if n != nil {
		if n.Content.Text == text {
			return nil
		}
		n.Content.Text = text
		n.Content.PreviewPlain = preview(text)
		n.Content.SetUpdateTime(now)
		return s.Save(items.Notes{*n}, nil)
	}

	if strings.TrimSpace(text) == "" {
		return nil
	}
	nn, err := items.NewNote(noteTitle, text, nil)
	if err != nil {
		return err
	}
	nn.Content.NoteType = "plain-text"
	nn.Content.EditorIdentifier = "com.standardnotes.plain-text"
	nn.Content.PreviewPlain = preview(text)
	nn.Content.SetUpdateTime(now)
	if !created.IsZero() {
		nn.CreatedAt = created.UTC().Format(common.TimeLayout)
		nn.CreatedAtTimestamp = created.UnixMicro()
	}
	tg.Content.UpsertReferences(items.ItemReferences{{UUID: nn.UUID, ContentType: common.SNItemTypeNote}})
	return s.Save(items.Notes{nn}, items.Tags{*tg})
}

// find locates the tag by exact title and, among its references, the
// non-trashed note with the exact title. More than one of either is an error.
func find(tags items.Tags, notes items.Notes, tagTitle, noteTitle string) (*items.Tag, *items.Note, error) {
	var tg *items.Tag
	for i := range tags {
		if tags[i].Content.Title != tagTitle || tags[i].Deleted {
			continue
		}
		if tg != nil {
			return nil, nil, fmt.Errorf("%w: more than one tag titled %q", ErrAmbiguous, tagTitle)
		}
		tg = &tags[i]
	}
	if tg == nil {
		return nil, nil, fmt.Errorf("%w: %q", ErrTagNotFound, tagTitle)
	}

	inTag := map[string]bool{}
	for _, r := range tg.Content.References() {
		if r.ContentType == common.SNItemTypeNote {
			inTag[r.UUID] = true
		}
	}

	var n *items.Note
	for i := range notes {
		c := notes[i].Content
		if !inTag[notes[i].UUID] || notes[i].Deleted || c.Title != noteTitle || (c.Trashed != nil && *c.Trashed) {
			continue
		}
		if n != nil {
			return nil, nil, fmt.Errorf("%w: more than one note titled %q in tag %q", ErrAmbiguous, noteTitle, tagTitle)
		}
		n = &notes[i]
	}
	return tg, n, nil
}

func preview(text string) string {
	p := strings.Join(strings.Fields(text), " ")
	if r := []rune(p); len(r) > 100 {
		p = string(r[:100])
	}
	return p
}
