package histutil

import "strings"

// DB is the interface of the storage database.
type DB interface {
	NextCmdSeq() (int, error)
	AddCmd(cmd string) (int, error)
	Cmds(from, upto int) ([]string, error)
	PrevCmd(upto int, prefix string) (int, string, error)
}

// TestDB is an implementation of the DB interface that can be used for testing.
type TestDB struct {
	AllCmds []string

	OneOffError error
}

func (s *TestDB) error() error {
	err := s.OneOffError
	s.OneOffError = nil
	return err
}

func (s *TestDB) NextCmdSeq() (int, error) {
	return len(s.AllCmds), s.error()
}

func (s *TestDB) AddCmd(cmd string) (int, error) {
	if s.OneOffError != nil {
		return -1, s.error()
	}
	s.AllCmds = append(s.AllCmds, cmd)
	return len(s.AllCmds) - 1, nil
}

func (s *TestDB) Cmds(from, upto int) ([]string, error) {
	return s.AllCmds[from:upto], s.error()
}

func (s *TestDB) PrevCmd(upto int, prefix string) (int, string, error) {
	if s.OneOffError != nil {
		return -1, "", s.error()
	}
	if upto < 0 || upto > len(s.AllCmds) {
		upto = len(s.AllCmds)
	}
	for i := upto - 1; i >= 0; i-- {
		if strings.HasPrefix(s.AllCmds[i], prefix) {
			return i, s.AllCmds[i], nil
		}
	}
	return -1, "", ErrEndOfHistory
}
