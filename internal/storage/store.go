package storage

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("object not found")

type Store struct {
	baseDir string
}

func NewStore(baseDir string) (*Store, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	return &Store{
		baseDir: baseDir,
	}, nil
}

func (s *Store) objectPath(name string) string {
	return filepath.Join(s.baseDir, name)
}

func (s *Store) Put(name string, data []byte) error {
	path := s.objectPath(name)

	return os.WriteFile(path, data, 0644)
}

func (s *Store) Get(name string) ([]byte, error) {
	path := s.objectPath(name)

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}

	return data, err
}

func (s *Store) Delete(name string) error {
	path := s.objectPath(name)

	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}

	return err
}