package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type Store interface {
	Save(ctx context.Context, key string, r io.Reader) (string, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
}

type LocalStore struct {
	root string
}

func NewLocalStore(root string) *LocalStore {
	return &LocalStore{root: root}
}

func (s *LocalStore) Save(ctx context.Context, key string, r io.Reader) (string, error) {
	target := filepath.Join(s.root, filepath.Clean(key))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	file, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, r); err != nil {
		return "", err
	}
	return target, ctx.Err()
}

func (s *LocalStore) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.Open(path)
}

type S3Store struct{}

func NewS3Store() *S3Store { return &S3Store{} }
