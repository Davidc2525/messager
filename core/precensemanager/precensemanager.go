package precensemanager

import (
	"github.com/Davidc2525/messager/log"
	"sync"
)

var (
	log             = mlog.New()
	ProcenceManager *Manager
	provides        = make(map[string]Provider)
	once            sync.Once
)

//interface para proveedor de presencia
//este proveedor sirve para saber en q lugares del cluster
//esta conectado un usuario
type Provider interface {
	Add(string, string) error
	Remove(string, string) error
	Get(string) ([]string, error)
}

type Manager struct {
	Provider Provider
}

func New(providerName string) {
	log.Warning.Println("Starting precense manager with provider: ", providerName)
	once.Do(func() {
		provider, ok := provides[providerName]
		if !ok {
			log.Error.Panicln("no provider found", providerName)
		}
		log.Warning.Printf("PROVIDER %#v\n", provider)
		ProcenceManager = &Manager{provider}
	})
}
func Register(name string, provider Provider) {
	log.Warning.Printf("Register %s %$v", name, provider)
	if provider == nil {
		log.Error.Panicln("session: Register provider is nil")
	}
	if _, dup := provides[name]; dup {
		log.Error.Panicln("session: Register called twice for provider " + name)
	}
	provides[name] = provider
}

func init() {

}
