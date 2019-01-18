package etcd

//TODO detalles
import (
	"container/list"
	"context"
	"fmt"

	"github.com/Davidc2525/messager/core/sessionmanager"
	"github.com/Davidc2525/messager/log"
	"go.etcd.io/etcd/clientv3"

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
	if _, err := pder.cli.Put(context.Background(), newKeyAttr(st.sid, key.(string)), value.(string)); err == nil {

	} else {
		return err
	}
	//st.value[key] = value
	pder.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) Get(key interface{}) interface{} {

	pder.SessionUpdate(st.sid)
	if res, err := pder.cli.Get(context.Background(), newKeyAttr(st.sid, key.(string))); err == nil {
		if len(res.Kvs) > 0 {
			return string(res.Kvs[0].Value)
		} else {
			return nil
		}
	} else {
		return nil
	}

	return nil
}

func (st *SessionStore) Delete(key interface{}) error {

	if _, err := pder.cli.Delete(context.Background(), newKeyAttr(st.sid, key.(string))); err == nil {

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
	cli      *clientv3.Client
	lock     sync.Mutex // lock
	sessions map[string]sessionmanager.Session
	list     *list.List // gc
}

func (pder *Provider) SessionInit(sid string) (sessionmanager.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	//SESS:sid = timeAccessed
	//SESS:sid_<name attr> = data
	if _, err := pder.cli.Put(context.Background(), newKey(sid), time.Now().String()); err == nil {
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
		if res, err := pder.cli.Get(context.Background(), newKey(sid)); err == nil {
			if len(res.Kvs) > 0 {
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
		if _, err := pder.cli.Delete(context.Background(), newKey(sid), clientv3.WithPrefix()); err == nil {
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

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"orchi:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Error.Panicln(err.Error())
	}

	log.Warning.Println("Starting ETCD session provider")
	pder.sessions = make(map[string]sessionmanager.Session, 0)
	pder.cli = cli
	sessionmanager.Register("etcd", pder)
}
