// Command sn-journal reads and writes a daily diary note in a Standard Notes
// tag, reusing sn-cli's keychain session.
//
//	sn-journal title
//	sn-journal pull --tag T --title X --out FILE   # prints "found" or "new"
//	sn-journal push --tag T --title X --file FILE [--created RFC3339]
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
	"github.com/ketankunkalikar/sn-journal/journal"
)

const appName = "sn-journal"

// Exit codes the shell wrapper relies on.
const (
	exitErr       = 1
	exitUsage     = 2
	exitAmbiguous = 3
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
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "sn-journal:", err)
		if errors.Is(err, journal.ErrAmbiguous) {
			os.Exit(exitAmbiguous)
		}
		os.Exit(exitErr)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: sn-journal title | pull --tag T --title X --out FILE | push --tag T --title X --file FILE [--created RFC3339]")
	os.Exit(exitUsage)
}

func pull(args []string) error {
	fs := flag.NewFlagSet("pull", flag.ExitOnError)
	tag := fs.String("tag", "", "tag title")
	title := fs.String("title", "", "note title")
	out := fs.String("out", "", "file to write")
	_ = fs.Parse(args)
	if *tag == "" || *title == "" || *out == "" {
		usage()
	}

	st, err := openStore()
	if err != nil {
		return err
	}
	defer st.Close()

	text, found, err := journal.Pull(st, *tag, *title)
	if err != nil {
		return err
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

	return journal.Push(st, *tag, *title, string(text), created, time.Now().UTC())
}

// gosnStore is the Standard Notes implementation of journal.Store. Load
// leaves the cache DB open so Save can write into it before syncing.
type gosnStore struct {
	s  cache.Session
	db *storm.DB
}

func openStore() (*gosnStore, error) {
	session.SetKeyringIdentity("sn-cli", "Session")
	s, _, err := cache.GetSession(nil, true, "", "", false)
	if err != nil {
		return nil, fmt.Errorf("loading sn-cli session (run `sn session --add`): %w", err)
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

func (g *gosnStore) Save(notes items.Notes, tags items.Tags) error {
	if len(notes) > 0 {
		if err := cache.SaveNotes(&g.s, g.db, notes, false); err != nil {
			return err
		}
	}
	if len(tags) > 0 {
		if err := cache.SaveTags(g.db, &g.s, tags, false); err != nil {
			return err
		}
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
