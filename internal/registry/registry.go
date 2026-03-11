package registry

import (
	"errors"
	"sync"

	"localgate/internal/logger"
)

// ServiceEntry はサービスのエントリを表す
type ServiceEntry struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

// ErrNotFound は未登録サービスへの操作時に返されるエラー
var ErrNotFound = errors.New("service not found")

// ErrAlreadyExists は同名サービスが既に存在する場合に返されるエラー
var ErrAlreadyExists = errors.New("service already exists")

// ServiceRegistry はサービスのルーティングテーブルを管理するインターフェース
type ServiceRegistry interface {
	Register(name, target string, force bool) error
	Deregister(name string) error
	Lookup(name string) (target string, found bool)
	List() []ServiceEntry
}

type inMemoryRegistry struct {
	mu       sync.RWMutex
	services map[string]string
}

// NewServiceRegistry は新しいインメモリサービスレジストリを返す
func NewServiceRegistry() ServiceRegistry {
	return &inMemoryRegistry{
		services: make(map[string]string),
	}
}

func (r *inMemoryRegistry) Register(name, target string, force bool) error {
	if name == "" || target == "" {
		return errors.New("name and target are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.services[name]; exists && !force {
		logger.Debug("registry.Register: 既に登録済み", "name", name)
		return ErrAlreadyExists
	}
	r.services[name] = target
	logger.Debug("registry.Register: 登録完了", "name", name, "target", target, "force", force)
	return nil
}

func (r *inMemoryRegistry) Deregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.services[name]; !ok {
		logger.Debug("registry.Deregister: 未登録", "name", name)
		return ErrNotFound
	}
	delete(r.services, name)
	logger.Debug("registry.Deregister: 解除完了", "name", name)
	return nil
}

func (r *inMemoryRegistry) Lookup(name string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	target, ok := r.services[name]
	logger.Debug("registry.Lookup", "name", name, "found", ok, "target", target)
	return target, ok
}

func (r *inMemoryRegistry) List() []ServiceEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entries := make([]ServiceEntry, 0, len(r.services))
	for name, target := range r.services {
		entries = append(entries, ServiceEntry{Name: name, Target: target})
	}
	return entries
}
