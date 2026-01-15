package filestore

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ondrej/chat/internal/user"
)

var log = slog.Default().With("component", "store")

type FileStore struct {
	path	string
	users   atomic.Value // map[string]UserRecord
	modTime atomic.Value // time.Time

	reloadMu sync.Mutex
}

type UserRecord struct {
	Username string
	Password string
	Role user.Role
	Disabled bool
}

func NewFileStore(path string) (*FileStore, error) {
	s := &FileStore{path: path}

	s.users.Store(map[string]UserRecord{})
	s.modTime.Store(time.Time{})

	if err := s.reload(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *FileStore) getRecord(username string) (UserRecord, error) {
	if err := s.maybeReload(); err != nil {
		return UserRecord{}, err
	}

	u := s.users.Load().(map[string]UserRecord)
	rec, ok := u[username]
	if !ok {
		return UserRecord{}, fmt.Errorf("User %s not found", username)
	}

	return rec, nil
}

func (s *FileStore) GetUserByUsername(username string) (*user.User, error) {
	if err := s.maybeReload(); err != nil {
		return nil, err
	}

	u := s.users.Load().(map[string]UserRecord)
	rec, ok := u[username]
	if !ok {
		return nil, fmt.Errorf("User %s not found", username)
	}

	return recordToDomain(rec), nil
}

func (s *FileStore) Authenticate(username, password string) (*user.User, error) {
	if err := s.maybeReload(); err != nil {
		return nil, err
	}

	m := s.users.Load().(map[string]UserRecord)
	rec, ok := m[username]
	if !ok || rec.Disabled {
		return nil, errors.New("")
	}

	if password != rec.Password {
		return nil, errors.New("")
	}

	return recordToDomain(rec), nil
}

func (s *FileStore) parseUsersFile() (map[string]UserRecord, error) {
	log.Debug("Parsing the file user DB")
	out := make(map[string]UserRecord)
	file, err := os.Open(s.path)
	re := regexp.MustCompile(`^(.*):(.*):(.*)$`)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := re.FindStringSubmatch(line); len(matches) > 0 {
			u := matches[1]
			p := matches[2]
			r := parseRole(matches[3])

			out[u] = UserRecord{
					Username: u,
					Password: p,
					Role: r,
				}
		}
	}

	return out, nil
}

func (s *FileStore) maybeReload() error {
	fi, err := os.Stat(s.path)
	if err != nil {
		return fmt.Errorf("stat users file: %w", err)
	}

	current := fi.ModTime()
	last := s.modTime.Load().(time.Time)

	if !current.After(last) {
		return nil
	}

	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()

	fi2, err := os.Stat(s.path)
	if err != nil {
		return fmt.Errorf("stat users file: %w", err)
	}

	current2 := fi2.ModTime()
	last2 := s.modTime.Load().(time.Time)

	if !current2.After(last2) {
		return nil
	}

	return s.reload()
}

func (s *FileStore) reload() error {
	fi, err := os.Stat(s.path)
	if err != nil {
		return fmt.Errorf("stat users file: %w", err)
	}

	f, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("open users file: %w", err)
	}
	defer f.Close()

	parsed, err := s.parseUsersFile()
	if err != nil {
		return fmt.Errorf("parse users file: %w", err)
	}

	s.users.Store(parsed)
	s.modTime.Store(fi.ModTime())
	return nil
}

func parseRole(str string) user.Role {
	role, ok := user.RoleFromString(str)
	if !ok {
		log.Error("Unknown role: " + str)
		return user.RoleUser
	}

	return role
}

func recordToDomain(rec UserRecord) *user.User {
	return &user.User{
		Username: rec.Username,
		Role:     rec.Role,
		Disabled: rec.Disabled,
	}
}
