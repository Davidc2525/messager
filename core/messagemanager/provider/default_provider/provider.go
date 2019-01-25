package default_provider

import (
	"github.com/Davidc2525/messager/core/messagemanager/inboxitem"
	"github.com/Davidc2525/messager/core/messagemanager/storeprovider"
	"github.com/Davidc2525/messager/core/messagemanager/storeprovider/storememoryprovider"
	"github.com/Davidc2525/messager/core/messagemanager/storeprovider/storeredisprovider"
	"github.com/Davidc2525/messager/core/user"
	"github.com/Davidc2525/messager/log"
	"sync"
)

var (
	log = mlog.New()
)

type DefaultProvider struct {
	//Op chan messagemanager.Op

	StoreProvider storeprovider.StoreProvider
	lock          sync.Mutex
}

func (this *DefaultProvider) UpdateAndGetItem(usr *user.User, itemId string) (*inboxitem.InboxItem, error) {
	var members []*user.User
	if in, err := this.StoreProvider.GetInbox(usr); err == nil {
		if item, err := in.Get(in.IndexOfById(itemId)); err == nil {
			in.MoveFromt(item)

			for _, member := range item.GetMembers() {
				members = append(members, member)
			}

			for _, member := range members { //update all members inbox by itemId

				if inMember, err := this.StoreProvider.GetInbox(member); err == nil {

					if inItemMember, err := inMember.Get(inMember.IndexOfById(itemId)); err == nil {
						inMember.MoveFromt(inItemMember)


					} else {
						log.Warning.Printf("miembro %s no item %s", member.Id, itemId)
					}

				} else {
					log.Warning.Printf("miembro %s no inbox", member.Id)
				}

			}

			return item,nil
		} else { //error en get item
			return nil, err
		}
	} else { //error en get inbox
		return nil, err
	}
}

func (this *DefaultProvider) GetMembersOfItemWithoutMe(usr *user.User, itemId string) ([]*user.User, error) {
	var memberswom []*user.User
	if members, err := this.GetMembersOfItem(usr, itemId); err == nil {
		for _, m := range members {
			if usr.Id != m.Id {
				memberswom = append(memberswom, m)
			}
		}
		return memberswom, nil
	} else {
		return nil, err
	}
}

func (this *DefaultProvider) GetMembersOfItem(usr *user.User, itemId string) ([]*user.User, error) {

	var members []*user.User
	if in, err := this.StoreProvider.GetInbox(usr); err == nil {
		if item, err := in.Get(in.IndexOfById(itemId)); err == nil {
			for _, member := range item.GetMembers() {
				members = append(members, member)
			}
			return members, nil
		} else { //error en get item
			return nil, err
		}
	} else { //error en get inbox
		return nil, err
	}
}

func (this *DefaultProvider) UpdateAndGetMembersOfItem(usr *user.User, itemId string) ([]*user.User, error) {
	//panic("implement me")
	var members []*user.User
	if in, err := this.StoreProvider.GetInbox(usr); err == nil {
		if item, err := in.Get(in.IndexOfById(itemId)); err == nil {
			in.MoveFromt(item)

			for _, member := range item.GetMembers() {
				members = append(members, member)
			}

			for _, member := range members { //update all members inbox by itemId

				if inMember, err := this.StoreProvider.GetInbox(member); err == nil {

					if inItemMember, err := inMember.Get(inMember.IndexOfById(itemId)); err == nil {
						inMember.MoveFromt(inItemMember)


					} else {
						log.Warning.Printf("miembro %s no item %s", member.Id, itemId)
					}

				} else {
					log.Warning.Printf("miembro %s no inbox", member.Id)
				}

			}

			return this.GetMembersOfItemWithoutMe(usr,itemId)
		} else { //error en get item
			return nil, err
		}
	} else { //error en get inbox
		return nil, err
	}
}

func (this *DefaultProvider) DeleteItem(usr *user.User, itemId string) error {
	if in, err := this.StoreProvider.GetInbox(usr); err == nil {
		index := in.IndexOfById(itemId)
		if _, err := in.Get(index); err == nil {
			in.DeleteIn(index)
			return nil
		} else { //error en get item
			return err
		}
	} else { //error en get inbox
		return err
	}
}

func (this *DefaultProvider) GetStore() storeprovider.StoreProvider {
	return this.StoreProvider
}

func (this *DefaultProvider) Start() {

	/*go func() {
		for {
			select {
			case op := <-this.Op:
				fmt.Println("        DefaultProvider", op)
				switch v := op.(type) {
				case *messagemanager.GetMembers: //obtener miembros de una conversacion en el inbox de un usuario

				//actualiazar la el inbox de el usuario y de los miembros colocando
				//la conversacion en el tope de la lista
				case *messagemanager.UpdateAndGetMembers:
					fmt.Println("Update and get members", v.GetUser(), v.GetInboxItemId())
					getInboxOp := storeprovider.NewGetInbox(v.GetUser())

					this.StoreProvider.MakeOp() <- getInboxOp

					if r, ok := <-getInboxOp.Receive; ok { //exist
						getInboxOp.ReleaseOp()

						iigetop := inbox.NewGet()
						iigetop.Id = v.GetInboxItemId()

						r.Op <- iigetop

						if reg, ok := <-iigetop.Receive; ok {
							fmt.Println(" 		iigetop", reg, ok)
							//r.Op <- inbox.Move{Method: inbox.FROMT, Input: reg.Item}

							getMembers := &inboxitem.GetMembers{Receive: make(chan []*user.User)}
							reg.Item.InOp <- getMembers
							if mbs, ok := <-getMembers.Receive; ok {
								getMembers.ReleaseOpInboxItem()
								for _, m := range mbs { //update imbox to members
									getInboxOp := storeprovider.NewGetInbox(m)
									this.StoreProvider.MakeOp() <- getInboxOp
									if r, ok := <-getInboxOp.Receive; ok { //exist
										getInboxOp.ReleaseOp()
										iigetop := inbox.NewGet()
										iigetop.Id = v.GetInboxItemId()

										r.Op <- iigetop
										if reg, ok := <-iigetop.Receive; ok {
											close(iigetop.Receive)
											r.Op <- inbox.Move{Method: inbox.FROMT, Input: reg.Item}
										}
									}
								}
								v.Receive <- mbs

							}
						} else {
							fmt.Println(" 	No existe conversacion", v.GetInboxItemId())
							v.Release()
						}
					} else { //no exist inbox, search in db
						fmt.Println(" 	No existe conversacion", v.GetInboxItemId())
						v.Release()
					}

				}
			}
		}
	}()*/
}

func NewDefaultProvider() *DefaultProvider {
	p := &DefaultProvider{
		//Op:            make(chan messagemanager.Op),
		StoreProvider: storememoryprovider.NewMemoryStorage(),
	}
	p.StoreProvider.SetSubProvider(storeredisprovider.NewRedisStoreProvider())
	p.Start()
	return p
}
func init() {
	//messagemanager.Register("memory",NewMemoryProvider())
}

/*
func (this *DefaultProvider) MakeOp() chan messagemanager.Op {
	return this.Op
}*/
