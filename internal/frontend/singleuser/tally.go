package singleuser

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/horazont/webskat/internal/skat"
)

const (
	userDirectoryName = "users"
)

type TallyUser struct {
	ClientID     string `json:"-"`
	ClientSecret string `json:"client_secret"`
	DisplayName  string `json:"display_name"`
}

type Tally struct {
	rootDirectory string
	userCache     map[string]*TallyUser
	currentGame   *skat.GameState
}

func newClientID() (string, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(buf), nil
}

func NewUser(clientSecret string, displayName string) (*TallyUser, error) {
	clientID, err := newClientID()
	if err != nil {
		return nil, err
	}
	return &TallyUser{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		DisplayName:  displayName,
	}, nil
}

func ReadUser(filename string) (*TallyUser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	result := &TallyUser{}
	if err := dec.Decode(result); err != nil {
		return nil, err
	}
	if !strings.HasSuffix(filename, ".json") {
		panic("not a .json file!")
	}
	basename := filepath.Base(filename)
	result.ClientID = basename[:len(basename)-len(".json")]
	return result, nil
}

func (u *TallyUser) Write(userdir string) error {
	outname := filepath.Join(userdir, u.ClientID+".json")
	tmpfile, err := ioutil.TempFile(userdir, "."+u.ClientID+".*")
	if err != nil {
		return err
	}
	tmpfileName := tmpfile.Name()
	defer tmpfile.Close()
	defer os.Remove(tmpfileName)

	enc := json.NewEncoder(tmpfile)
	err = enc.Encode(u)
	if err != nil {
		return err
	}
	err = tmpfile.Sync()
	if err != nil {
		return err
	}
	err = tmpfile.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tmpfileName, outname)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tally) userDirectory() string {
	return filepath.Join(t.rootDirectory, userDirectoryName)
}

func (t *Tally) initDirectories() error {
	if err := os.Mkdir(t.rootDirectory, 0700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	if err := os.Mkdir(t.userDirectory(), 0700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	return nil
}

func (t *Tally) readUsers() error {
	userdir := t.userDirectory()
	files, err := ioutil.ReadDir(userdir)
	if err != nil {
		return err
	}
	userCache := make(map[string]*TallyUser)
	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		user, err := ReadUser(filepath.Join(userdir, name))
		if err != nil {
			return err
		}
		userCache[user.ClientID] = user
	}
	t.userCache = userCache
	return nil
}

func (t *Tally) Register(clientSecret, displayName string) (string, error) {
	user, err := NewUser(clientSecret, displayName)
	if err != nil {
		return "", err
	}
	err = user.Write(t.userDirectory())
	if err != nil {
		return "", err
	}
	t.userCache[user.ClientID] = user
	return user.ClientID, nil
}

func (t *Tally) NewGame(playerIDs [3]string, dealerID string) (*skat.GameState, error) {
	if t.currentGame != nil {
		return nil, errors.New("game in progress!")
	}
	var err error
	t.currentGame, err = skat.NewGame(dealerID != "", skat.LeagueScoreDefinition())
	return t.currentGame, err
}

func NewTally(dataDirectory string) (*Tally, error) {
	result := &Tally{
		rootDirectory: dataDirectory,
	}
	if err := result.initDirectories(); err != nil {
		return nil, err
	}
	if err := result.readUsers(); err != nil {
		return nil, err
	}
	return result, nil
}
