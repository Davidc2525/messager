package etcd

//TODO detalles
import (
	"container/list"
	"fmt"
	"github.com/go-redis/redis"

	"github.com/Davidc2525/messager/core/sessionmanager"
	"github.com/Davidc2525/messager/log"
	"sync"
	"time"
)

var (
	log           = mlog.New()
	keyPrefix     = "SESS"
	keyPrefixAttr = "SESS_"
)

func newKey(sid string) string {
	if len(sid) == 0 {
		return keyPrefix
	} else {
		return keyPrefix + "_" + sid
	}
}

func newKeyAttr(sid string, attr string) string {
	if len(sid) == 0 {
		return keyPrefix
	} else {
		if len(attr) == 0 {
			return keyPrefix + "_" + sid
		} else {
			return keyPrefix + "_" + sid + "_" + attr
		}
	}
}

var pder = &Provider{list: list.New()}

type SessionStore struct {
	sid          string                      // unique session id
	timeAccessed time.Time                   // last access time
	value        map[interface{}]interface{} // session value stored inside
}

func (st *SessionStore) Set(key, value interface{}) error {
	pder.cli.Set(newKeyAttr(st.sid, key.(string)), value.(string), 0)

	//st.value[key] = value
	pder.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) Get(key interface{}) interface{} {

	pder.SessionUpdate(st.sid)
	if res, err := pder.cli.Get(newKeyAttr(st.sid, key.(string))).Result(); err == nil {
		return res
	} else {
		return nil
	}

	return nil
}

func (st *SessionStore) Delete(key interface{}) error {

	if err := pder.cli.Del(newKeyAttr(st.sid, key.(string))).Err(); err == nil {
	} else {
		return err
	}
	pder.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) SessionID() string {
	return st.sid
}

type Provider struct {
	cli      *redis.ClusterClient
	lock     sync.Mutex // lock
	sessions map[string]sessionmanager.Session
	list     *list.List // gc
}

func (pder *Provider) SessionInit(sid string) (sessionmanager.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	//SESS:sid = timeAccessed
	//SESS:sid_<name attr> = data
	if err := pder.cli.Set(newKey(sid), time.Now().String(), 0).Err(); err == nil {
		newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
		pder.sessions[sid] = newsess
		return newsess, nil
	} else {
		return nil, err
	}

	return nil, nil
}

func (pder *Provider) SessionRead(sid string) (sessionmanager.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if s, ok := pder.sessions[sid]; !ok {
		if res, err := pder.cli.Get(newKey(sid)).Result(); err == nil {
			if len(res) > 0 {
				sess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: make(map[interface{}]interface{})}
				return sess, nil
			} else {
				delete(pder.sessions, sid)
				return nil, fmt.Errorf("no session")
			}
			/*datak := res.Kvs[0].Key
			datav := res.Kvs[0].Key

			//SESS_sid
			conposeKey := string(datak)
			partsKey := strings.Split(conposeKey,"_")
			key := partsKey[1]*/

		} else {
			return nil, err
		}
	} else {
		return s, nil
	}
	return nil, nil
}

func (pder *Provider) SessionDestroy(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if _, ok := pder.sessions[sid]; ok {
		if keys, err := pder.cli.Keys(newKey(sid) + "*").Result(); err == nil {
			for _, v := range keys {
				pder.cli.Del(v)
			}
			delete(pder.sessions, sid)
		} else {
			return nil
		}
		return nil
	}
	return nil
}

func (pder *Provider) SessionGC(maxlifetime int64) {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*SessionStore).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
			//pder.list.Remove(element)
			delete(pder.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

func (pder *Provider) SessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if _, ok := pder.sessions[sid]; ok {
		//element.Value.(*SessionStore).timeAccessed = time.Now()
		//pder.list.MoveToFront(element)
		return nil
	}
	return nil
}

func GetProvider() sessionmanager.Provider {
	log.Warning.Println("1")
	return pder
}

func init() {
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"192.168.0.1:7000",
			"192.168.0.1:7001",
			"192.168.0.1:7002",
			"192.168.0.1:7003",
			"192.168.0.1:7004",
			"192.168.0.1:7005"},
	})
	cli.Ping()
	//cli.ZAdd("",redis.Z{})
	/*cli := redis.NewClient(&redis.Options{

		Addr:     "localhost:30001",
		Password: "", // no password set
		DB:       0,  // use default DB

	})*/

	log.Warning.Println("Starting REDIS session provider")
	pder.sessions = make(map[string]sessionmanager.Session, 0)
	pder.cli = cli
	sessionmanager.Register("redis", pder)
}
