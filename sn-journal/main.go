// Command sn-journal reads and writes a daily diary note in a Standard Notes
// tag, reusing sn-cli's keychain session.
//
//	sn-journal title
//	sn-journal pull --tag T --title X --out FILE [--password-stdin]   # prints "found" or "new"
//	sn-journal push --tag T --title X --file FILE [--created RFC3339] [--protect]
//	sn-journal unlock   # checks the account password on stdin
//
// A note protected in Standard Notes is only pulled with --password-stdin and
// the account password, the same one the app asks for.
package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/crypto"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
	"github.com/ketankunkalikar/sn-journal/journal"
)

const appName = "sn-journal"

// Exit codes the Neovim plugin relies on.
const (
	exitErr           = 1
	exitUsage         = 2
	exitAmbiguous     = 3
	exitLocked        = 4
	exitWrongPassword = 5
)

var (
	errLocked        = errors.New("entry is protected; the account password is required")
	errWrongPassword = errors.New("incorrect password")
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "title":
		fmt.Println(journal.Title(time.Now()))
	case "pull":
		err = pull(os.Args[2:])
	case "push":
		err = push(os.Args[2:])
	case "unlock":
		err = unlock()
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "sn-journal:", err)
		switch {
		case errors.Is(err, journal.ErrAmbiguous):
			os.Exit(exitAmbiguous)
		case errors.Is(err, errLocked):
			os.Exit(exitLocked)
		case errors.Is(err, errWrongPassword):
			os.Exit(exitWrongPassword)
		}
		os.Exit(exitErr)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: sn-journal title | pull --tag T --title X --out FILE [--password-stdin] | push --tag T --title X --file FILE [--created RFC3339] [--protect] | unlock")
	os.Exit(exitUsage)
}

func pull(args []string) error {
	fs := flag.NewFlagSet("pull", flag.ExitOnError)
	tag := fs.String("tag", "", "tag title")
	title := fs.String("title", "", "note title")
	out := fs.String("out", "", "file to write")
	pwStdin := fs.Bool("password-stdin", false, "read the account password from stdin, to open a protected entry")
	_ = fs.Parse(args)
	if *tag == "" || *title == "" || *out == "" {
		usage()
	}
	var password string
	if *pwStdin {
		var err error
		if password, err = readPassword(); err != nil {
			return err
		}
	}

	st, err := openStore()
	if err != nil {
		return err
	}
	defer st.Close()

	text, found, protected, err := journal.Pull(st, *tag, *title)
	if err != nil {
		return err
	}
	if protected {
		if !*pwStdin {
			return errLocked
		}
		if err := checkPassword(st.s.Session, password); err != nil {
			return err
		}
	}
	if err := os.WriteFile(*out, []byte(text), 0o600); err != nil {
		return err
	}
	if found {
		fmt.Println("found")
	} else {
		fmt.Println("new")
	}
	return nil
}

func push(args []string) error {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	tag := fs.String("tag", "", "tag title")
	title := fs.String("title", "", "note title")
	file := fs.String("file", "", "file to read")
	createdS := fs.String("created", "", "creation time for a new note (RFC3339)")
	protect := fs.Bool("protect", false, "mark the note protected in Standard Notes")
	_ = fs.Parse(args)
	if *tag == "" || *title == "" || *file == "" {
		usage()
	}
	var created time.Time
	if *createdS != "" {
		var err error
		if created, err = time.Parse(time.RFC3339, *createdS); err != nil {
			return fmt.Errorf("--created: %w", err)
		}
	}

	text, err := os.ReadFile(*file)
	if err != nil {
		return err
	}

	st, err := openStore()
	if err != nil {
		return err
	}
	defer st.Close()

	return journal.Push(st, *tag, *title, string(text), created, time.Now().UTC(), *protect)
}

// unlock checks the password on stdin without touching any note, so the
// plugin can let the user retry, and can guard an entry left on disk.
func unlock() error {
	password, err := readPassword()
	if err != nil {
		return err
	}
	s, err := loadSession()
	if err != nil {
		return err
	}
	return checkPassword(s.Session, password)
}

func readPassword() (string, error) {
	b, err := io.ReadAll(io.LimitReader(os.Stdin, 4096))
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}
	return strings.TrimRight(string(b), "\r\n"), nil
}

// checkPassword verifies the account password the way Standard Notes does
// before showing a protected note: derive the root key from it and compare
// with the session's. Nothing is sent to the server.
func checkPassword(s *session.Session, password string) error {
	kp := s.KeyParams
	nonce := kp.PwNonce
	if nonce == "" {
		nonce = s.PasswordNonce
	}
	if kp.Version != "004" || kp.Identifier == "" || nonce == "" || s.MasterKey == "" {
		return errors.New("session lacks 004 key params to check the password; run `sn session --add` again")
	}
	mk, _, err := crypto.GenerateMasterKeyAndServerPassword004(crypto.GenerateEncryptedPasswordInput{
		UserPassword:  password,
		Identifier:    kp.Identifier,
		PasswordNonce: nonce,
	})
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(mk), []byte(s.MasterKey)) != 1 {
		return errWrongPassword
	}
	return nil
}

// gosnStore is the Standard Notes implementation of journal.Store. Load
// leaves the cache DB open so Save can write into it before syncing.
type gosnStore struct {
	s  cache.Session
	db *storm.DB
}

func loadSession() (cache.Session, error) {
	session.SetKeyringIdentity("sn-cli", "Session")
	s, _, err := cache.GetSession(nil, true, "", "", false)
	if err != nil {
		return s, fmt.Errorf("loading sn-cli session (run `sn session --add`): %w", err)
	}
	return s, nil
}

func openStore() (*gosnStore, error) {
	s, err := loadSession()
	if err != nil {
		return nil, err
	}
	if s.CacheDBPath, err = cache.GenCacheDBPath(s, "", appName); err != nil {
		return nil, err
	}
	return &gosnStore{s: s}, nil
}

func (g *gosnStore) Load() (items.Tags, items.Notes, error) {
	// AlwaysSync: the base must be current or saving creates a conflict copy.
	so, err := cache.Sync(cache.SyncInput{Session: &g.s, AlwaysSync: true})
	if err != nil {
		return nil, nil, fmt.Errorf("sync: %w", err)
	}
	g.db = so.DB

	var all cache.Items
	if err := g.db.All(&all); err != nil {
		return nil, nil, err
	}
	its, err := all.ToItems(&g.s)
	if err != nil {
		return nil, nil, err
	}
	return its.Tags(), its.Notes(), nil
}

func (g *gosnStore) Protected(uuid string) (bool, error) {
	c, err := g.rawContent(uuid)
	if err != nil {
		return false, err
	}
	var p bool
	if v, ok := c["protected"]; ok {
		if err := json.Unmarshal(v, &p); err != nil {
			return false, fmt.Errorf("note %s: protected: %w", uuid, err)
		}
	}
	return p, nil
}

func (g *gosnStore) Save(notes items.Notes, tags items.Tags, protect bool) error {
	var eis items.EncryptedItems
	for i := range notes {
		e, err := g.encrypt(&notes[i], protect)
		if err != nil {
			return err
		}
		eis = append(eis, e)
	}
	for i := range tags {
		e, err := g.encrypt(&tags[i], false)
		if err != nil {
			return err
		}
		eis = append(eis, e)
	}
	if err := cache.SaveEncryptedItems(g.db, eis, false); err != nil {
		return err
	}
	g.Close()
	if _, err := cache.Sync(cache.SyncInput{Session: &g.s, Close: true}); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	return nil
}

func (g *gosnStore) Close() {
	if g.db != nil {
		_ = g.db.Close()
		g.db = nil
	}
}

// rawContent decrypts the cached item's content JSON as stored, including
// fields gosn-v2 drops when parsing, such as a note's "protected".
func (g *gosnStore) rawContent(uuid string) (map[string]json.RawMessage, error) {
	var ci cache.Item
	if err := g.db.One("UUID", uuid, &ci); err != nil {
		return nil, fmt.Errorf("item %s: %w", uuid, err)
	}
	di, err := items.DecryptItem(items.EncryptedItem{
		UUID:        ci.UUID,
		Content:     ci.Content,
		ContentType: ci.ContentType,
		ItemsKeyID:  ci.ItemsKeyID,
		EncItemKey:  ci.EncItemKey,
	}, g.s.Session, g.s.ItemsKeys)
	if err != nil {
		return nil, err
	}
	var c map[string]json.RawMessage
	if err := json.Unmarshal([]byte(di.Content), &c); err != nil {
		return nil, fmt.Errorf("item %s: %w", uuid, err)
	}
	return c, nil
}

// encrypt encrypts the item's content laid over what Standard Notes already
// has for it, so that saving from here keeps fields gosn-v2 does not model.
func (g *gosnStore) encrypt(it items.Item, protect bool) (items.EncryptedItem, error) {
	typed, err := json.Marshal(it.GetContent())
	if err != nil {
		return items.EncryptedItem{}, err
	}
	var content map[string]json.RawMessage
	if err := json.Unmarshal(typed, &content); err != nil {
		return items.EncryptedItem{}, err
	}
	switch base, err := g.rawContent(it.GetUUID()); {
	case err == nil:
		content = merge(base, content)
	case !errors.Is(err, storm.ErrNotFound):
		return items.EncryptedItem{}, err
	}
	if protect {
		content["protected"] = json.RawMessage("true")
	}
	b, err := json.Marshal(content)
	if err != nil {
		return items.EncryptedItem{}, err
	}
	return items.EncryptItem(&rawItem{Item: it, content: b}, g.s.DefaultItemsKey, g.s.Session)
}

// merge lays over onto base, recursing into objects both have, so keys only
// base has survive.
func merge(base, over map[string]json.RawMessage) map[string]json.RawMessage {
	for k, v := range over {
		var bo, oo map[string]json.RawMessage
		if json.Unmarshal(base[k], &bo) == nil && json.Unmarshal(v, &oo) == nil && bo != nil && oo != nil {
			if nested, err := json.Marshal(merge(bo, oo)); err == nil {
				base[k] = nested
				continue
			}
		}
		base[k] = v
	}
	return base
}

// rawItem is an item whose content encrypts as the given JSON.
type rawItem struct {
	items.Item
	content rawContent
}

func (r *rawItem) GetContent() items.Content { return r.content }

type rawContent []byte

func (c rawContent) MarshalJSON() ([]byte, error)     { return c, nil }
func (rawContent) References() items.ItemReferences   { return nil }
func (rawContent) SetReferences(items.ItemReferences) {}
