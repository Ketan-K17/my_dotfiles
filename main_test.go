package main

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/jonhadfield/gosn-v2/auth"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/crypto"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
)

const testPassword = "correct horse battery staple"

func testSession(t *testing.T) *session.Session {
	t.Helper()
	kp := auth.KeyParams{Identifier: "me@example.com", PwNonce: "0123456789abcdef0123456789abcdef", Version: "004"}
	mk, _, err := crypto.GenerateMasterKeyAndServerPassword004(crypto.GenerateEncryptedPasswordInput{
		UserPassword: testPassword, Identifier: kp.Identifier, PasswordNonce: kp.PwNonce,
	})
	if err != nil {
		t.Fatal(err)
	}
	key, _ := crypto.GenerateItemKey(64)
	ik := session.SessionItemsKey{UUID: items.GenUUID(), ItemsKey: key, Version: "004", Default: true}
	return &session.Session{MasterKey: mk, KeyParams: kp, ItemsKeys: []session.SessionItemsKey{ik}, DefaultItemsKey: ik}
}

func TestCheckPassword(t *testing.T) {
	s := testSession(t)
	if err := checkPassword(s, testPassword); err != nil {
		t.Fatalf("right password: %v", err)
	}
	if err := checkPassword(s, testPassword+"x"); !errors.Is(err, errWrongPassword) {
		t.Fatalf("wrong password: %v", err)
	}
	s.KeyParams.Version = "003"
	if err := checkPassword(s, testPassword); err == nil || errors.Is(err, errWrongPassword) {
		t.Fatalf("unsupported key params should be an error, got %v", err)
	}
}

func TestMergeKeepsUnknownFields(t *testing.T) {
	var base, over map[string]json.RawMessage
	_ = json.Unmarshal([]byte(`{"protected":true,"text":"old","appData":{"org.standardnotes.sn":{"pinned":true,"archived":true}}}`), &base)
	_ = json.Unmarshal([]byte(`{"text":"new","appData":{"org.standardnotes.sn":{"pinned":false}}}`), &over)
	got, _ := json.Marshal(merge(base, over))
	want := `{"appData":{"org.standardnotes.sn":{"archived":true,"pinned":false}},"protected":true,"text":"new"}`
	if string(got) != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
}

// openTestStore returns a store over a temporary cache DB holding n with the
// raw content given.
func openTestStore(t *testing.T, n items.Note, raw string) *gosnStore {
	t.Helper()
	db, err := storm.Open(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	g := &gosnStore{s: cache.Session{Session: testSession(t)}, db: db}
	e, err := items.EncryptItem(&rawItem{Item: &n, content: rawContent(raw)}, g.s.DefaultItemsKey, g.s.Session)
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.SaveEncryptedItems(db, items.EncryptedItems{e}, false); err != nil {
		t.Fatal(err)
	}
	return g
}

// saved decrypts what encrypt produced.
func saved(t *testing.T, g *gosnStore, e items.EncryptedItem) map[string]json.RawMessage {
	t.Helper()
	if err := cache.SaveEncryptedItems(g.db, items.EncryptedItems{e}, false); err != nil {
		t.Fatal(err)
	}
	c, err := g.rawContent(e.UUID)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEncryptKeepsProtection(t *testing.T) {
	n, _ := items.NewNote("Friday", "old", nil)
	g := openTestStore(t, n, `{"title":"Friday","text":"old","protected":true,"futureField":"x"}`)
	if p, err := g.Protected(n.UUID); err != nil || !p {
		t.Fatalf("Protected = %v, %v", p, err)
	}

	n.Content.Text = "new"
	e, err := g.encrypt(&n, false)
	if err != nil {
		t.Fatal(err)
	}
	c := saved(t, g, e)
	if string(c["protected"]) != "true" || string(c["futureField"]) != `"x"` || string(c["text"]) != `"new"` {
		t.Fatalf("saved content lost fields: %v", c)
	}
}

func TestEncryptProtects(t *testing.T) {
	n, _ := items.NewNote("Friday", "text", nil)
	g := openTestStore(t, n, `{"title":"Friday","text":"text"}`)
	if p, _ := g.Protected(n.UUID); p {
		t.Fatal("should start unprotected")
	}
	e, err := g.encrypt(&n, true)
	if err != nil {
		t.Fatal(err)
	}
	saved(t, g, e)
	if p, err := g.Protected(n.UUID); err != nil || !p {
		t.Fatalf("Protected = %v, %v", p, err)
	}

	fresh, _ := items.NewNote("Saturday", "new note", nil)
	e, err = g.encrypt(&fresh, true)
	if err != nil {
		t.Fatal(err)
	}
	if c := saved(t, g, e); string(c["protected"]) != "true" || string(c["text"]) != `"new note"` {
		t.Fatalf("new note content %v", c)
	}
}
