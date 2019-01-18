package sessionmanager

import (
	"fmt"
	"github.com/Davidc2525/messager/log"
	"github.com/segmentio/ksuid"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var GlobalSessions *Manager
var provides = make(map[string]Provider)
var (
	log  = mlog.New()
	once sync.Once
)

type Manager struct {
	cookieName  string     //private cookiename
	lock        sync.Mutex // protects session
	provider    Provider
	maxlifetime int64
}

func (manager *Manager) sessionId() string {
	return ksuid.New().String()
}

func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		sid := manager.sessionId()
		session, _ = manager.provider.SessionInit(sid)
		cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(manager.maxlifetime)}

		http.SetCookie(w, &cookie)
	} else {
		sid := cookie.Value
		sess, err := manager.provider.SessionRead(sid)
		if err != nil {
			sid := manager.sessionId()
			session, _ = manager.provider.SessionInit(sid)
			cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(manager.maxlifetime)}
			http.SetCookie(w, &cookie)
		} else {
			session = sess
		}
	}
	return
}

func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		manager.lock.Lock()
		defer manager.lock.Unlock()
		manager.provider.SessionDestroy(cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

func (manager *Manager) GetSession(w http.ResponseWriter, r *http.Request, create bool) (Session, error) {

	cookie, err := r.Cookie(manager.cookieName)

	if err != nil || cookie.Value == "" {
		if create {
			return manager.SessionStart(w, r), nil
		} else {
			return nil, fmt.Errorf("no existe session")
		}
	} else {
		sess, err := manager.provider.SessionRead(cookie.Value)
		if err != nil {
			if create {
				return manager.SessionStart(w, r), nil
			} else {
				return nil, err
			}
		} else {
			return sess, nil
		}
	}

	return nil, nil

}

type ErrorSessison struct {
	Msg string
}

func (this *ErrorSessison) Error() string {
	return this.Msg
}

func GetInstance(provideName, cookieName string, maxlifetime int64) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	once.Do(func() {

		GlobalSessions = &Manager{provider: provider, cookieName: cookieName, maxlifetime: maxlifetime}
		go GlobalSessions.GC()
	})

	return GlobalSessions, nil
}

func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.SessionGC(manager.maxlifetime)
	time.AfterFunc(time.Duration(manager.maxlifetime), func() { manager.GC() })
}

// Then, initialize the session manager
func init() {
	/*log.Warning.Println("Iniciando session manager")

	GlobalSessions,err := GetInstance("memory", "gosessionid", 3600)
	if err!=nil{
		log.Error.Panic(err)
	}
	log.Warning.Printf("m: %#v",GlobalSessions)*/
}

type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64)
}

// Register makes a session provider available by the provided name.
// If a Register is called twice with the same name or if the driver is nil,
// it panics.
func Register(name string, provider Provider) {
	log.Warning.Printf("Register %s %$v", name, provider)
	if provider == nil {
		panic("session: Register provider is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provider
}

type Session interface {
	Set(key, value interface{}) error //set session value
	Get(key interface{}) interface{}  //get session value
	Delete(key interface{}) error     //delete session value
	SessionID() string                //back current sessionID
}
