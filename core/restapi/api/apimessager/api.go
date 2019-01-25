package apimessager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Davidc2525/messager/core/messagemanager"
	"github.com/Davidc2525/messager/core/messagemanager/inbox"
	"github.com/Davidc2525/messager/core/messagemanager/inboxitem"
	"github.com/Davidc2525/messager/core/restapi"
	"github.com/Davidc2525/messager/core/sessionmanager"
	user2 "github.com/Davidc2525/messager/core/user"
	"github.com/Davidc2525/messager/core/userprovider"
	"github.com/Davidc2525/messager/log"
	"github.com/gorilla/mux"
	"github.com/segmentio/ksuid"
	"html/template"
	"net/http"
	"strings"
)

var (
	log        = mlog.New()
	HostByUser map[string][]string
)

func sendJson(w http.ResponseWriter, data interface{}) {
	b, _ := json.Marshal(data)

	buf := bytes.Buffer{}
	json.Indent(&buf, b, "", "   ")
	w.Write(buf.Bytes())
}

func pos(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func remove(slice []string, s int) []string {
	if len(slice) < 0 || s < 0 {
		return slice
	}
	return append(slice[:s], slice[s+1:]...)
}

func filterMembers(slice []*user2.User, filter func(*user2.User) bool) []*user2.User {
	filtered := []*user2.User{}
	for _, m := range slice {
		if filter(m) {
			filtered = append(filtered, m)
		}
	}

	return filtered
}

func init() {
	HostByUser = make(map[string][]string)
	fmt.Println("iniciando api messager")

	restapi.Register2("/api/messager", func(messager *mux.Router) {
		messager.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			out := `
Api messager, descripcion:

	inbox # inbox, todas las conversaciones
	inbox/{item} # ir por conversacion
	inbox/range/{range} # ir por conversaciones en rango; inbox/range/0,-1
	inbox/seek/{seek} # ir por conversaciones desde seek; inbox/range/0
	inbox/create/private/{with} # crear una conversacion privada con un usuario
	inbox/create/group/{members} # crear una conversacion
	inbox/delete/{item} # eliminar conversacion en inbox 
	inbox/{item}/members # obtener miembros de una conversacion		
	inbox/{item}/addmembers/{members} # agregar miembros a una conversacion		
	inbox/{item}/removemembers/{members} # eliminar miembros de una conversacion			
	inbox/{item}/messages/ # optener los ultmimos mensajes de esa conversacion (20)
	inbox/{item}/messages/range/{range} # optener los los mensajes de esa conversacion por rango; inbox/123/messages/range/mid,mid
	inbox/{item}/messages/seek/{seek} # optener los ultmimos mensajes de esa conversacion luego de (seeks) mensajes; inbox/123/messages/seek/mid
			`
			writer.Write([]byte(out))
		})
		messager.HandleFunc("/inbox/create/private/{with}", func(w http.ResponseWriter, r *http.Request) {

			s, err := sessionmanager.GlobalSessions.GetSession(w, r, false)
			if err != nil {
				//json.NewEncoder(w).Encode()
				sendJson(w, struct {
					Status bool   `json:"status" xml:"status"`
					Msg    string `json:"msg" xml:"msg"`
				}{
					Status: false,
					Msg:    "u no logged",
				})

			} else {
				if s.Get("uid").(string) == mux.Vars(r)["with"] {
					sendJson(w, struct {
						Status bool         `json:"status"`
						Msg    string       `json:"msg"`
						User   *user2.User  `json:"user"`
						Inbox  *inbox.Inbox `json:"inbox"`
					}{Status: false, Msg: "no puedes crear una conversacion contigo mismo."})
					return
				}

				if u, err := userprovider.GetInstance().GetUserById(s.Get("uid").(string)); err == nil {
					w.Header().Set("content-type", "application/json")
					withId := mux.Vars(r)["with"]

					if with, err := userprovider.GetInstance().GetUserById(withId); err == nil {

						if inOwner, err := messagemanager.Manager.Pder.GetStore().GetInbox(u); err == nil {

							if inGuest, err := messagemanager.Manager.Pder.GetStore().GetInbox(with); err == nil {
								if pin, err := messagemanager.Manager.Pder.GetStore().GetPrivateItem(u, with); err == nil {
									sendJson(w, struct {
										Status bool         `json:"status"`
										Msg    string       `json:"msg"`
										User   *user2.User  `json:"user"`
										Inbox  *inbox.Inbox `json:"inbox"`
									}{Status: false, Msg: "ya tienes una conversaicon creada con " + with.Id+": "+pin})
									return
								}
								newId := ksuid.New().String()
								item := inboxitem.NewInboxItem()
								item.Id = newId
								item.Owner = u.Id
								item.Type = inboxitem.PRIVATE
								item.AddMember(u)
								item.AddMember(with)
								inOwner.PutFromt(item)

								item = nil

								item = inboxitem.NewInboxItem()
								item.Id = newId
								item.Owner = u.Id
								item.Type = inboxitem.PRIVATE
								item.AddMember(u)
								item.AddMember(with)
								inGuest.PutFromt(item)

								if err := messagemanager.Manager.Pder.GetStore().MapPrivateItem(item); err != nil {
									log.Warning.Printf("no se pudo mapear %v", item.Id)
								}

								if err := messagemanager.Manager.Pder.GetStore().SaveInbox(u, inOwner); err != nil {
									log.Warning.Printf("no se pudo guardar el inbox de %v", inOwner.Owner.Id)
								}
								if err := messagemanager.Manager.Pder.GetStore().SaveInbox(with, inGuest); err != nil {
									log.Warning.Printf("no se pudo guardar el inbox de %v", inGuest.Owner.Id)
								}

								sendJson(w, struct {
									Status bool         `json:"status"`
									Msg    string       `json:"msg"`
									User   *user2.User  `json:"user"`
									Inbox  *inbox.Inbox `json:"inbox"`
								}{Status: true, User: u, Inbox: inOwner})

							} else {
								//el usuario invitado no tiene imbox
								sendJson(w, struct {
									Status bool         `json:"status"`
									Msg    string       `json:"msg"`
									User   *user2.User  `json:"user"`
									Inbox  *inbox.Inbox `json:"inbox"`
								}{Status: false, Msg: err.Error()})

							}

						} else {
							//el usuario q esta creando no tiene inbox
							sendJson(w, struct {
								Status bool         `json:"status"`
								Msg    string       `json:"msg"`
								User   *user2.User  `json:"user"`
								Inbox  *inbox.Inbox `json:"inbox"`
							}{Status: false, Msg: err.Error()})
						}
					} else {
						//no existe user
						sendJson(w, struct {
							Status bool         `json:"status"`
							Msg    string       `json:"msg"`
							User   *user2.User  `json:"user"`
							Inbox  *inbox.Inbox `json:"inbox"`
						}{Status: false, Msg: err.Error()})
					}

				} else {
					//no existe el usuario q esta creando la conversacion
					sendJson(w, struct {
						Status bool         `json:"status"`
						Msg    string       `json:"msg"`
						User   *user2.User  `json:"user"`
						Inbox  *inbox.Inbox `json:"inbox"`
					}{Status: false, Msg: err.Error()})
				}

			}

		})

		messager.HandleFunc("/inbox/create/group/{members}", func(w http.ResponseWriter, r *http.Request) {

			s, err := sessionmanager.GlobalSessions.GetSession(w, r, false)
			if err != nil {
				//json.NewEncoder(w).Encode()
				sendJson(w, struct {
					Status bool   `json:"status" xml:"status"`
					Msg    string `json:"msg" xml:"msg"`
				}{
					Status: false,
					Msg:    "u no logged",
				})

			} else {
				if u, error := userprovider.GetInstance().GetUserById(s.Get("uid").(string)); error == nil {
					w.Header().Set("content-type", "application/json")
					membersId := mux.Vars(r)["members"]
					var members []*user2.User
					for _, id := range strings.Split(membersId, ",") {
						idt := strings.TrimSpace(id)
						if _, err := userprovider.GetInstance().GetUserById(id); err == nil {
							members = append(members, &user2.User{Id: idt})
						} else {
							log.Warning.Printf("el usuario %s no existe", id)
						}
					}

					if in, err := messagemanager.Manager.Pder.GetStore().GetInbox(u); err == nil {
						/*
							item := inboxitem.NewInboxItem()
							item.Id = ksuid.New().String()
							item.AddMembers(members)
							in.PutFromt(item)*/

						membersWithOutMe := filterMembers(members, func(usr *user2.User) bool {
							if usr.Id != u.Id {
								return true
							} else {
								return false
							}
						})

						newId := ksuid.New().String()
						item := inboxitem.NewInboxItem()
						item.Id = newId
						item.Owner = u.Id
						item.Type = inboxitem.GROUP
						item.AddMembers(membersWithOutMe)
						item.AddMember(u)
						in.PutFromt(item)

						for _, member := range members {
							if member.Id != u.Id {
								if in, err := messagemanager.Manager.Pder.GetStore().GetInbox(member); err == nil {
									item := inboxitem.NewInboxItem()
									item.Id = newId
									item.Owner = u.Id
									item.Type = inboxitem.GROUP
									item.AddMembers(filterMembers(members, func(usr *user2.User) bool {
										if usr.Id != member.Id {
											return true
										} else {
											return false
										}
									}))
									item.AddMember(u)
									item.AddMember(member)
									in.PutFromt(item)
								} else {
									log.Warning.Printf("el usuario %s no tiene inbox", member.Id)
								}
							}

						}

						sendJson(w, struct {
							Status bool         `json:"status"`
							Msg    string       `json:"msg"`
							User   *user2.User  `json:"user"`
							Inbox  *inbox.Inbox `json:"inbox"`
						}{Status: true, User: u, Inbox: in})
					} else {
						sendJson(w, struct {
							Status bool         `json:"status"`
							Msg    string       `json:"msg"`
							User   *user2.User  `json:"user"`
							Inbox  *inbox.Inbox `json:"inbox"`
						}{Status: true, User: u, Msg: err.Error()})
					}
				}

			}

		})

		messager.HandleFunc("/inbox", func(w http.ResponseWriter, r *http.Request) {

			s, err := sessionmanager.GlobalSessions.GetSession(w, r, false)
			if err != nil {
				//json.NewEncoder(w).Encode()
				sendJson(w, struct {
					Status bool   `json:"status" xml:"status"`
					Msg    string `json:"msg" xml:"msg"`
				}{
					Status: false,
					Msg:    "u no logged",
				})

			} else {
				if u, error := userprovider.GetInstance().GetUserById(s.Get("uid").(string)); error == nil {
					w.Header().Set("content-type", "application/json")

					if in, err := messagemanager.Manager.Pder.GetStore().GetInbox(u); err == nil {
						sendJson(w, struct {
							status bool
							User   *user2.User  `json:"user"`
							Inbox  *inbox.Inbox `json:"inbox"`
						}{status: true, User: u, Inbox: in})
					} else {
						sendJson(w, struct {
							Error  bool   `json:"erro"`
							Reason string `json:"reason"`
						}{
							Error:  true,
							Reason: err.Error(),
						})
					}
				} else {
					sendJson(w, struct {
						Error  bool   `json:"erro"`
						Reason string `json:"reason"`
					}{
						Error:  true,
						Reason: error.Error(),
					})
				}

			}

		})

		messager.HandleFunc("/inbox/{item}", func(w http.ResponseWriter, r *http.Request) {

			s, err := sessionmanager.GlobalSessions.GetSession(w, r, false)
			if err != nil {
				//json.NewEncoder(w).Encode()
				sendJson(w, struct {
					Status bool   `json:"status" xml:"status"`
					Msg    string `json:"msg" xml:"msg"`
				}{
					Status: false,
					Msg:    "u no logged",
				})

			} else {
				if u, error := userprovider.GetInstance().GetUserById(s.Get("uid").(string)); error == nil {
					w.Header().Set("content-type", "application/json")

					if in, err := messagemanager.Manager.Pder.GetStore().GetInbox(u); err == nil || in != nil {
						if inItem, err := in.Get(in.IndexOfById(mux.Vars(r)["item"])); err == nil || inItem != nil {
							sendJson(w, inItem)
						} else { //no item
							sendJson(w, struct {
								Error  bool   `json:"erro"`
								Reason string `json:"reason"`
							}{
								Error:  true,
								Reason: err.Error(),
							})
						}
					} else { //no inbox
						sendJson(w, struct {
							Error  bool   `json:"erro"`
							Reason string `json:"reason"`
						}{
							Error:  true,
							Reason: err.Error(),
						})
					}

				} else {
					sendJson(w, struct {
						Error  bool   `json:"erro"`
						Reason string `json:"reason"`
					}{
						Error:  true,
						Reason: error.Error(),
					})
				}

			}

		})

	})
	//restapi.Register("/api/user",*r)
}

var form = `
<html>
    <head>
    <title></title>
    </head>
    <body>
        <form id="form" action="/api/user/login" method="post">
            Username:<input type="text" name="username">
            Password:<input type="password" name="password">
            <input type="submit" onclick="return login()" value="Login">
        </form>
    </body>
    <script>
        var f = document.getElementById("form");
        login = ()=>{
            
            var user = f.elements.username;
            var pass = f.elements.password;
            
            var xhr = new XMLHttpRequest();
            var data = new FormData();
                data.append("user",user.value);
                data.append("pass",pass.value);
            xhr.open("post","/api/user/login");
            xhr.onload = (e)=>{
                var data = JSON.parse(e.target.response);
                if (data.Loging){
                    document.location = "/api/user/get"
                }else{
                    alert(data.msg)
                }
            };
            xhr.send(data);
            
            return false
        }
    </script>
</html>
`
var tplForm, _ = template.New("formLogin").Parse(form)
