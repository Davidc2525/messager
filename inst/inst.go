package inst

import "fmt"
import "sync"

var (
	r    *repository
	once sync.Once
)

type repository struct {
	items map[string]interface{}
	mu    sync.RWMutex
}

func (r *repository) Set(key string, data interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[key] = data
}

func (r *repository) Get(key string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[key]
	if !ok {
		return "", fmt.Errorf("The '%s' is not presented", key)
	}
	return item, nil
}

func Repository() *repository {
	once.Do(func() {
		r = &repository{
			items: make(map[string]interface{}),
		}
	})

	return r
}
