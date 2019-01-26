package storeredisprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Davidc2525/messager/core/messagemanager/inbox"
	"github.com/Davidc2525/messager/core/messagemanager/inboxitem"
	"github.com/Davidc2525/messager/core/messagemanager/message"
	"github.com/Davidc2525/messager/core/messagemanager/storeprovider"
	"github.com/Davidc2525/messager/log"

	"github.com/Davidc2525/messager/core/user"
	"github.com/go-redis/redis"
)

var (
	log = mlog.New()
)

func makeKeyByMap(usrs map[string]*user.User) (r []*user.User) {
	adds := 0
	for _, u := range usrs {
		if adds >= 2 {
			break
		}
		r = append(r, u)
		adds++
	}
	return
}

func makeKey(usr1 *user.User, usr2 *user.User) string {
	return "PRIVATES:" + usr1.Id + ":" + usr2.Id
}

type RedisStoreProvider struct {
	cli *redis.ClusterClient
}

func NewRedisStoreProvider() *RedisStoreProvider {
	pder := &RedisStoreProvider{}
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

	pder.cli = cli
	return pder
}

func (this *RedisStoreProvider) GetSubProvider() storeprovider.StoreProvider {
	return nil
}

func (this *RedisStoreProvider) SetSubProvider(sub storeprovider.StoreProvider) {

}

func (this *RedisStoreProvider) GetInbox(usr *user.User) (*inbox.Inbox, error) {
	//var dataB bytes.Buffer
	var inb inbox.Inbox
	status := this.cli.Get("inbox:" + usr.Id)
	if data, err := status.Result(); err == nil {
		if err := json.NewDecoder(bytes.NewBuffer([]byte(data))).Decode(&inb); err == nil {
			return &inb, nil
		} else {

			return nil, fmt.Errorf(err.Error())
		}
	} else {
		return nil, fmt.Errorf("usuario %v no tiene inbox", usr.Id)
	}

}

func (this *RedisStoreProvider) SaveInbox(usr *user.User, in *inbox.Inbox) error {
	var network bytes.Buffer // Stand-in for the network.
	json.NewEncoder(&network).Encode(in)
	this.cli.Set("inbox:"+usr.Id, network.String(), 0)
	return nil
}

func (this *RedisStoreProvider) CreateInbox(usr *user.User, members []*user.User) (*inbox.Inbox, error) {
	return nil, nil
}

func (this *RedisStoreProvider) CreateInboxWith(usr *user.User, in *inbox.Inbox) error {
	status := this.cli.Get("inbox:" + usr.Id)
	if err := status.Err(); err == redis.Nil { // si no existe, se crea un inbox para ese usuario

		var network bytes.Buffer // Stand-in for the network.
		json.NewEncoder(&network).Encode(in)
		this.cli.Set("inbox:"+usr.Id, network.String(), 0)

	} else {

		return fmt.Errorf("el usuario %v ya tiene inbox: %v, %v", usr, err)
	}

	return nil
}

func (this *RedisStoreProvider) DeleteInbox(usr *user.User, inbox *inbox.Inbox) error {
	this.cli.Del("inbox:" + usr.Id)
	return nil
}

func (this *RedisStoreProvider) SaveAllInbox() error {
	return nil
}

func (this *RedisStoreProvider) MapPrivateItem(item *inboxitem.InboxItem) error {
	if item.Type == inboxitem.PRIVATE {
		uss := makeKeyByMap(item.GetMembers())

		key1 := makeKey(uss[0], uss[1])
		key2 := makeKey(uss[1], uss[0])

		this.cli.Set(key1, item.Id, 0)
		this.cli.Set(key2, item.Id, 0)

		return nil

	} else {
		//no es privado
		return fmt.Errorf("el item %v no es privado", item.Id)
	}

}

func (this *RedisStoreProvider) UnMapPrivateItem(item *inboxitem.InboxItem) error {
	if item.Type == inboxitem.PRIVATE {
		uss := makeKeyByMap(item.GetMembers())

		key1 := makeKey(uss[0], uss[1])
		key2 := makeKey(uss[1], uss[0])

		this.cli.Del(key1, key2)

		return nil

	} else {
		//no es privado
		return fmt.Errorf("el item %v no es privado", item.Id)
	}
}

func (this *RedisStoreProvider) GetPrivateItem(usr1 *user.User, usr2 *user.User) (string, error) {
	key1 := makeKey(usr1, usr2)
	key2 := makeKey(usr2, usr1)
	if idItem, err := this.cli.Get(key1).Result(); err == nil && len(idItem) > 0 {
		return idItem, nil
	} else if idItem, err := this.cli.Get(key2).Result(); err == nil && len(idItem) > 0 {
		return idItem, nil
	} else {
		return "", fmt.Errorf("no se econtro ningun item")
	}

}

func (this *RedisStoreProvider) StoreMessage(item *inboxitem.InboxItem, msg *message.Message) error {

	for k, _ := range item.GetMembers() {
		key := k + ":" + item.Id
		var network bytes.Buffer // Stand-in for the network.
		json.NewEncoder(&network).Encode(msg)
		log.Warning.Printf("		STORE MESSAGE PERSISTENT %v %#v", key, network.String())

		this.cli.ZAdd(key, redis.Z{
			Score:  msg.Id,
			Member: network.String(),
		})
	}
	return nil
}
