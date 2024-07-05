package cache

import (
	"github.com/zhex/local-bbp/internal/common"
	"github.com/zhex/local-bbp/internal/models"
	"path"
)

type Store struct {
	Path string
	data models.Caches
}

func NewStore(path string, caches models.Caches) *Store {
	return &Store{
		Path: path,
		data: common.MergeMaps(defaultCaches, caches),
	}
}

func (s *Store) Get(key string) *models.Cache {
	if cache, ok := s.data[key]; ok {
		return cache
	}
	return nil
}

func (s *Store) HasCachePath(key string) bool {
	p := path.Join(s.Path, key)
	return common.IsDirExists(p)
}

func (s *Store) HasHashPath(key string, hash string) bool {
	p := s.GetHashPath(key, hash)
	return common.IsDirExists(p)
}

func (s *Store) GetHashPath(key string, hash string) string {
	return path.Join(s.Path, key, hash)
}
