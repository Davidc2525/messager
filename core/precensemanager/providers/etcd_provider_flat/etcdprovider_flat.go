package etcd_provider_flat

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/Davidc2525/messager/core/precensemanager"
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/user"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"regexp"
	"time"
)

var (
	log       = mlog.New()
	pder      = &EtcdProvider{}
	mapPrefix = "PRECENSE2"

	//el primer grupo obtiene el usuario y el segundo el host
	//el tercero el id de la conexcion en el endpoint
	//PRECENSE2:55_orchi:2626_1FLl8ywl7Dn1K0cd6iI48WzbbTd ->
	//[":55_orchi:2626_1FLl8ywl7Dn1K0cd6iI48WzbbTd" "55" "orchi:2626" "1FLl8ywl7Dn1K0cd6iI48WzbbTd"]
	reg, _ = regexp.Compile(`:(.+)_(.+)_(.+)`)
)

/*Funciones de ayuda*/
//Indice en de string
func indexOf(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

//eliminar en indice de un []string
func remove(slice []string, s int) []string {
	log.Error.Printf("REMOVE %s %s", slice, s)
	if len(slice) == 0 || s < 0 {
		return slice
	}
	return append(slice[:s], slice[s+1:]...)
}

//obtener todos menos los repetidos
func disct(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

func toGob(entry *Entry) *bytes.Buffer {
	b := new(bytes.Buffer)
	gob.NewEncoder(b).Encode(entry)
	return b
}

func toEntry(entry []byte) *Entry {
	db := new(bytes.Buffer)
	db.Write(entry)
	newEntry := new(Entry)
	gob.NewDecoder(db).Decode(newEntry)
	return newEntry
}

/*Funciones de ayuda*/

type Entry struct {
	Uid   string
	Hosts []string
}

func (this *Entry) add(host string) {
	hs := this.Hosts
	if indexOf(hs, host) == -1 {
		hs = append(hs, host)
		this.Hosts = hs
	}
}

func (this *Entry) Remove(host string) {
	hs := this.Hosts
	if len(hs) > 0 {
		hs = remove(hs, indexOf(hs, host))
		this.Hosts = hs
	}
}

func (this *Entry) AllHosts() []string {
	return disct(this.Hosts)
}

func NewEntry() *Entry {
	return &Entry{Hosts: []string{}}
}

func newKey(key string, host string) (prefixKey string) {
	if len(key) == 0 {
		prefixKey = mapPrefix + ":"
	} else {
		if len(host) == 0 {
			prefixKey = mapPrefix + ":" + key
		} else {
			prefixKey = mapPrefix + ":" + key + "_" + host
		}
	}

	return
}

type EtcdProvider struct {
	cli  *clientv3.Client
	ctx  context.Context
	sess *concurrency.Session
	lock *concurrency.Mutex
}

func (this *EtcdProvider) Add(uid string, hostid string) error {
	log.Info.Println("add", uid, hostid)

	_, errorPut := this.cli.Put(this.ctx, newKey(uid, hostid), "")
	if errorPut != nil {
		return errorPut
	}

	return nil
}
func (this *EtcdProvider) Remove(uid string, hostid string) error { //TODO borrar de etcd cuando no kede mas host

	_, errorDel := this.cli.Delete(this.ctx, newKey(uid, hostid))
	if errorDel != nil {
		return errorDel
	}
	return nil
}
func (this *EtcdProvider) Get(uid string) ([]string, error) {

	log.Info.Println("get", uid)
	var members []string = []string{}
	res, err := this.cli.Get(this.ctx, newKey(uid, ""), clientv3.WithPrefix())
	if err == nil {
		if len(res.Kvs) > 0 {
			for _, k := range res.Kvs {
				ke := string(k.Key)
				parts := reg.FindStringSubmatch(ke)
				members = append(members, parts[2])
				members = disct(members)

			}
			return members, nil
		} else {
			return members, nil
		}
	} else { //error
		return members, nil
	}

	return nil, nil
}

func init() {

	log.Warning.Println("starting etcd flatend precense provider")

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"orchi:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Error.Panicln(err.Error())
	}

	sess, err := concurrency.NewSession(cli)
	if err != nil {
		log.Error.Fatal(err)
	}

	m1 := concurrency.NewMutex(sess, mapPrefix)

	pder.lock = m1
	pder.sess = sess
	pder.cli = cli
	pder.ctx = context.Background()

	go func() {
		for {
			log.Warning.Println("show data")
			//m1.Lock(pder.ctx)
			res, e := cli.Get(pder.ctx, newKey("", ""), clientv3.WithPrefix())
			//m1.Unlock(pder.ctx)
			n := 0
			if e == nil {
				for _, k := range res.Kvs {
					n++

					log.Warning.Printf("ITEM: %#v", string(k.Key))
				}
				log.Info.Println("CONEXIONES ", n)
			}
			time.Sleep(time.Second * 60)
		}
	}()

	if false {

		davidk := newKey("david", "")
		b := new(bytes.Buffer)
		gob.NewEncoder(b).Encode(user.User{Name: "david", Id: "23034087"})
		ctx := context.Background()
		cli.Put(ctx, davidk, string(b.Bytes()))
		x := 0
		go func() {
			for {
				c := context.Background()
				r, _ := cli.Get(c, davidk)

				db := new(bytes.Buffer)
				db.Write(r.Kvs[0].Value)
				david := new(user.User)
				gob.NewDecoder(db).Decode(&david)
				if x%1000 == 0 {
					fmt.Println(x, david)
				}
				x++
			}
		}()
	}

	precensemanager.Register("etcd_flat", pder)
}
