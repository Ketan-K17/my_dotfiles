package journal

import (
	"errors"
	"testing"
	"time"

	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

func at(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

func TestTitle(t *testing.T) {
	cases := map[string]string{
		"2026-11-01 12:00": "Sunday, November 1, '26",
		"2026-11-01 04:00": "Sunday, November 1, '26",
		"2026-11-02 03:59": "Sunday, November 1, '26",
		"2026-11-02 00:30": "Sunday, November 1, '26",
		"2027-01-01 02:00": "Thursday, December 31, '26",
		"2026-11-09 09:00": "Monday, November 9, '26",
	}
	for in, want := range cases {
		if got := Title(at(in)); got != want {
			t.Errorf("Title(%s) = %q, want %q", in, got, want)
		}
	}
}

// fakeStore records what was saved.
type fakeStore struct {
	tags      items.Tags
	notes     items.Notes
	protected map[string]bool
	savedN    items.Notes
	savedT    items.Tags
	savedProt bool
	saveCalls int
	loadErr   error
}

func (f *fakeStore) Load() (items.Tags, items.Notes, error) { return f.tags, f.notes, f.loadErr }
func (f *fakeStore) Protected(uuid string) (bool, error)    { return f.protected[uuid], nil }
func (f *fakeStore) Save(n items.Notes, t items.Tags, protect bool) error {
	f.saveCalls++
	f.savedN, f.savedT, f.savedProt = n, t, protect
	return nil
}

func note(title, text string) items.Note {
	n, _ := items.NewNote(title, text, nil)
	return n
}

func tag(title string, notes ...items.Note) items.Tag {
	tg, _ := items.NewTag(title, nil)
	for _, n := range notes {
		tg.Content.UpsertReferences(items.ItemReferences{{UUID: n.UUID, ContentType: common.SNItemTypeNote}})
	}
	return tg
}

const day = "Sunday, November 1, '26"

func TestPullFindsExactTitleInTag(t *testing.T) {
	want := note(day, "dear diary")
	other := note(day+" draft", "nope")    // superstring: must not match
	outside := note(day, "in another tag") // same title, not in tag
	s := &fakeStore{
		tags:  items.Tags{tag("musings / nsfw", want, other), tag("other", outside)},
		notes: items.Notes{other, outside, want},
	}
	text, found, prot, err := Pull(s, "musings / nsfw", day)
	if err != nil || !found || prot || text != "dear diary" {
		t.Fatalf("got %q %v %v %v", text, found, prot, err)
	}
}

func TestPullMissingNote(t *testing.T) {
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw")}}
	text, found, _, err := Pull(s, "musings / nsfw", day)
	if err != nil || found || text != "" {
		t.Fatalf("got %q %v %v", text, found, err)
	}
}

func TestPullMissingTagIsError(t *testing.T) {
	s := &fakeStore{tags: items.Tags{tag("Musings / nsfw")}} // case differs
	if _, _, _, err := Pull(s, "musings / nsfw", day); !errors.Is(err, ErrTagNotFound) {
		t.Fatalf("want ErrTagNotFound, got %v", err)
	}
}

func TestPullDuplicateTagIsError(t *testing.T) {
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw"), tag("musings / nsfw")}}
	if _, _, _, err := Pull(s, "musings / nsfw", day); !errors.Is(err, ErrAmbiguous) {
		t.Fatalf("want ErrAmbiguous, got %v", err)
	}
}

func TestPullDuplicateNoteIsError(t *testing.T) {
	a, b := note(day, "a"), note(day, "b")
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw", a, b)}, notes: items.Notes{a, b}}
	if _, _, _, err := Pull(s, "musings / nsfw", day); !errors.Is(err, ErrAmbiguous) {
		t.Fatalf("want ErrAmbiguous, got %v", err)
	}
}

func TestPullIgnoresTrashed(t *testing.T) {
	trashed := note(day, "old")
	tr := true
	trashed.Content.Trashed = &tr
	live := note(day, "live")
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw", trashed, live)}, notes: items.Notes{trashed, live}}
	text, found, _, err := Pull(s, "musings / nsfw", day)
	if err != nil || !found || text != "live" {
		t.Fatalf("got %q %v %v", text, found, err)
	}
}

func TestPushUpdatesExisting(t *testing.T) {
	n := note(day, "old")
	tg := tag("musings / nsfw", n)
	s := &fakeStore{tags: items.Tags{tg}, notes: items.Notes{n}}
	now := at("2026-11-01 22:15").UTC()
	if err := Push(s, "musings / nsfw", day, "new text", time.Time{}, now, false); err != nil {
		t.Fatal(err)
	}
	if len(s.savedN) != 1 || s.savedN[0].UUID != n.UUID || s.savedN[0].Content.Text != "new text" {
		t.Fatalf("saved %+v", s.savedN)
	}
	if got, _ := s.savedN[0].Content.GetUpdateTime(); !got.Equal(now.Truncate(time.Millisecond)) {
		t.Fatalf("update time %v", got)
	}
	if len(s.savedT) != 0 {
		t.Fatalf("tag should not be re-saved")
	}
}

func TestPushUnchangedTextSkipsSave(t *testing.T) {
	n := note(day, "same")
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw", n)}, notes: items.Notes{n}}
	if err := Push(s, "musings / nsfw", day, "same", time.Time{}, time.Now(), false); err != nil {
		t.Fatal(err)
	}
	if s.saveCalls != 0 {
		t.Fatalf("expected no save")
	}
}

func TestPushCreatesPlainTextNoteInTag(t *testing.T) {
	tg := tag("musings / nsfw")
	s := &fakeStore{tags: items.Tags{tg}}
	created := at("2026-11-01 21:00").UTC()
	now := at("2026-11-01 21:05").UTC()
	if err := Push(s, "musings / nsfw", day, "first words", created, now, false); err != nil {
		t.Fatal(err)
	}
	if len(s.savedN) != 1 || len(s.savedT) != 1 {
		t.Fatalf("saved %d notes %d tags", len(s.savedN), len(s.savedT))
	}
	n := s.savedN[0]
	if n.Content.Title != day || n.Content.Text != "first words" {
		t.Fatalf("note %+v", n.Content)
	}
	if n.Content.NoteType != "plain-text" || n.Content.EditorIdentifier != "com.standardnotes.plain-text" {
		t.Fatalf("type %q editor %q", n.Content.NoteType, n.Content.EditorIdentifier)
	}
	if n.CreatedAtTimestamp != created.UnixMicro() || n.CreatedAt != created.Format(common.TimeLayout) {
		t.Fatalf("created %v %v", n.CreatedAt, n.CreatedAtTimestamp)
	}
	refs := s.savedT[0].Content.References()
	if s.savedT[0].UUID != tg.UUID || len(refs) != 1 || refs[0].UUID != n.UUID || refs[0].ContentType != common.SNItemTypeNote {
		t.Fatalf("tag refs %+v", refs)
	}
}

func TestPushCreateSkipsEmptyText(t *testing.T) {
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw")}}
	if err := Push(s, "musings / nsfw", day, "  \n", time.Now(), time.Now(), false); err != nil {
		t.Fatal(err)
	}
	if s.saveCalls != 0 {
		t.Fatalf("empty new entry should not be created")
	}
}

func TestPushRefusesDuplicates(t *testing.T) {
	a, b := note(day, "a"), note(day, "b")
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw", a, b)}, notes: items.Notes{a, b}}
	if err := Push(s, "musings / nsfw", day, "x", time.Now(), time.Now(), false); !errors.Is(err, ErrAmbiguous) {
		t.Fatalf("want ErrAmbiguous, got %v", err)
	}
	if s.saveCalls != 0 {
		t.Fatal("must not save")
	}
}

func TestPullReportsProtected(t *testing.T) {
	n := note(day, "secret")
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw", n)}, notes: items.Notes{n}, protected: map[string]bool{n.UUID: true}}
	_, found, prot, err := Pull(s, "musings / nsfw", day)
	if err != nil || !found || !prot {
		t.Fatalf("got %v %v %v", found, prot, err)
	}
}

func TestPushProtectsUnchangedNote(t *testing.T) {
	n := note(day, "same")
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw", n)}, notes: items.Notes{n}}
	if err := Push(s, "musings / nsfw", day, "same", time.Time{}, time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if s.saveCalls != 1 || !s.savedProt || s.savedN[0].UUID != n.UUID {
		t.Fatalf("saves %d protect %v", s.saveCalls, s.savedProt)
	}
}

func TestPushProtectSkipsAlreadyProtected(t *testing.T) {
	n := note(day, "same")
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw", n)}, notes: items.Notes{n}, protected: map[string]bool{n.UUID: true}}
	if err := Push(s, "musings / nsfw", day, "same", time.Time{}, time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if s.saveCalls != 0 {
		t.Fatal("expected no save")
	}
}

func TestPushProtectsNewNote(t *testing.T) {
	s := &fakeStore{tags: items.Tags{tag("musings / nsfw")}}
	if err := Push(s, "musings / nsfw", day, "first", time.Now(), time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if !s.savedProt {
		t.Fatal("new note should be saved protected")
	}
}
