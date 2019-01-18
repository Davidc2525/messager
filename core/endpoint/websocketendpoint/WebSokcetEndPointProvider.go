package websocketendpoint

import (
	"github.com/Davidc2525/messager/cluster/clustermanager"
	"github.com/Davidc2525/messager/core/endpoint"
	"github.com/Davidc2525/messager/core/packets"
	"github.com/Davidc2525/messager/core/precensemanager"
	"github.com/Davidc2525/messager/core/processor"
	"github.com/Davidc2525/messager/core/restapi"
	"github.com/Davidc2525/messager/core/sessionmanager"
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/new_user_conn/src"
	user2 "github.com/Davidc2525/messager/user"
	"github.com/Davidc2525/messager/userprovider"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"html/template"
	"net/http"
	"time"
)

var (
	log        = mlog.New()
	HostByUser map[string][]string
)

var (
	name = "websocket"

	upgrader websocket.Upgrader
)

type WSConnectionEndPoint struct {
	id        string
	open      bool
	uid       string
	hostid    string
	MessageIn chan []byte
	//MessageOut chan []byte//borrar

	onOpen        chan bool //borrar
	onClose       chan bool //borrar
	handleOnCLose func()

	wsc *websocket.Conn
	w   http.ResponseWriter
	r   *http.Request
}

func (this *WSConnectionEndPoint) Host() string {
	return this.hostid
}

func (this *WSConnectionEndPoint) Kind() packets.ClassConnection {
	return packets.Local
}

func (this *WSConnectionEndPoint) IsOpen() bool {
	return this.open
}

func (this *WSConnectionEndPoint) GetId() string {
	return this.id
}

func (this *WSConnectionEndPoint) GetUid() string {
	return this.uid
}

func (this *WSConnectionEndPoint) SetHandleClose(f func()) {
	this.handleOnCLose = f
}

func (this *WSConnectionEndPoint) GetWriter() http.ResponseWriter {
	return this.w
}

func (this *WSConnectionEndPoint) GetRequest() *http.Request {
	return this.r
}

func NewWSConnectionEndPoint(wsc *websocket.Conn, w http.ResponseWriter, r *http.Request) endpoint.ConnectionEndPoint {
	wscc := &WSConnectionEndPoint{wsc: wsc, w: w, r: r}
	wscc.id = ksuid.New().String()
	wscc.MessageIn = make(chan []byte)
	//wscc.MessageOut = make(chan []byte)
	wscc.onOpen = make(chan bool)
	wscc.onClose = make(chan bool)
	wscc.handleOnCLose = func() {
		log.Warning.Println("Funcion Vacia")
	}
	wscc.hostid = clustermanager.GetInstance().Server.GetConf().Addr
	wscc.Start()
	return wscc
}

func (this *WSConnectionEndPoint) Start() {

	this.open = true
	this.wsc.SetCloseHandler(func(code int, text string) error {
		this.open = false
		log.Warning.Println("Señal para desconectar", this.id)
		processor.Pross.ConnEpOut <- this
		this.onClose <- true
		this.handleOnCLose()
		//close(this.MessageOut)
		close(this.MessageIn)
		close(this.onClose)
		close(this.onOpen)

		return nil
	})

	if sess, err := sessionmanager.GlobalSessions.GetSession(this.w, this.r, false); err == nil {
		if iuid := sess.Get("uid"); iuid != nil {
			this.uid = iuid.(string)

			processor.Pross.ConnEpIn <- this

			go func() {
				for this.open {
					select {
					case messageData := <-this.MessageIn:
						//log.Warning.Printf("SEND TO WS EP %v",messageData)
						this.wsc.WriteMessage(websocket.TextMessage, messageData)
					case <-this.onClose:
						this.open = false
					}

				}
			}()
			go func() {
				for this.open {
					if _, m, err := this.wsc.ReadMessage(); err == nil {
						//	log.Warning.Println(m)
						//this.MessageOut <- message
						var container = packets.NewDContainer(packets.Local, "", false, m)
						container.Cid = this.id
						processor.Pross.StreamIn <- container

						/*
								if len(m) > 0 {
								if m[1] == 33 { // 33 == !
									switch m[0] {
									case 109: // 109 == message
										mp := packets.NewDMessagePacket()
										data := bytes.NewReader(m[2:])
										jd := json.NewDecoder(bufio.NewReader(data))
										jd.Decode(mp)
										mp.SetAttr("from", "local")
										//set attr id host
										//cp := packets.NewDefaultContainer("","remote","",mp)

										processor.Pross.MessageIn <- mp
									case 101: // 101 = event
										eventpacket := packets.NewDEventPacket()
										data := bytes.NewReader(m[2:])
										jd := json.NewDecoder(bufio.NewReader(data))
										jd.Decode(eventpacket)
										eventpacket.SetAttr("from", "local")
										//set attr id host
										//cp := packets.NewDefaultContainer("","remote","",eventpacket)

										processor.Pross.EventIn <- eventpacket
									}

								}
							}

						*/
					} else {
						//this.MessageOut = nil
						this.open = false
						log.Error.Println(err)
						return
					}
				}
			}()
		} else {
			//close
			this.open = false
			this.wsc.Close()
		}

	} else {
		//close por no tener session
		this.open = false
		this.wsc.Close()
	}
}

func (this *WSConnectionEndPoint) Close() {

}

func (this *WSConnectionEndPoint) Send() chan<- []byte {
	return this.MessageIn
}

func (this *WSConnectionEndPoint) OnClose() <-chan bool {
	return this.onClose
}

func (this *WSConnectionEndPoint) OnOpen() <-chan bool {
	return this.onOpen
}

type (
	Conf struct {
		Addr string
	}
)

type WebSokcetEndPointProvider struct {
	conf Conf
}

func (this *WebSokcetEndPointProvider) Stop() error {
	panic("implement me")
}

func (this *WebSokcetEndPointProvider) Start() error {

	log.Info.Println("Starting endpoint", this.GetName())
	http.HandleFunc("/ws", handler)

	http.HandleFunc("/ep", func(writer http.ResponseWriter, r *http.Request) {
		var user *user2.User
		login := false
		sess, e := sessionmanager.GlobalSessions.GetSession(writer, r, false)
		if e != nil {
			log.Warning.Println(e)
		} else {
			if s := sess.Get("uid"); s != nil {
				login = true
				user, _ = userprovider.GetInstance().GetUserById(s.(string))
			} else {
				login = false
			}

			//log.Warning.Println("session ", sess)
		}

		//rs:=hola{"david"}
		//json.NewEncoder(writer).Encode(rs)

		writer.Header().Add("content-type", "text/html")
		data := struct {
			Title    string
			Items    []string
			Addr     string
			Cookies  []*http.Cookie
			Login    bool
			User     *user2.User
			RestAddr string
		}{
			RestAddr: restapi.REST.Conf.Addr,
			Title:    "My page",
			Items: []string{
				"My photos",
				"My blog",
			},
			Addr:    this.conf.Addr,
			Cookies: r.Cookies(),
			Login:   login,
			User:    user,
		}
		tplHome.Execute(writer, data)

	})

	err := http.ListenAndServe(this.conf.Addr, nil)

	if err != nil {
		log.Error.Panicf("Can´t start endpoint "+this.GetName()+", ", err)
	}
	return err

}

func New(conf Conf) *WebSokcetEndPointProvider {
	wep := &WebSokcetEndPointProvider{}
	wep.conf = conf

	return wep
}

func (this *WebSokcetEndPointProvider) GetName() string {
	return name
}

func (this *WebSokcetEndPointProvider) SendTo() error {

	return &SendError{What: "No se pudo enviar el mensaje"}
}

func handler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error.Println(err)
	}
	log.Info.Printf("nueva cliente %v\n", conn.RemoteAddr())
	go NewWSConnectionEndPoint(conn, w, r)
	//go newcon(conn, w, r)

}

type Fase1 struct {
	Fase int    `json:"fase"`
	Uid  string `json:"uid"`
}
type Fase1r struct {
	Fase  int    `json:"fase"`
	Uid   string `json:"uid"`
	Pased bool   `json:"pased"`
}

type Fase3 struct {
	Fase int    `json:"fase"`
	Uid  string `json:"uid"`
	Msg  string `json:"msg"`
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
	return append(slice[:s], slice[s+1:]...)
}
func closeOnError(err error, conn *websocket.Conn, er *bool) {
	if err != nil {

		conn.Close()
		*er = true
		log.Error.Println("cliente desconectado", conn.RemoteAddr())
	}
}

type ClientConn struct {
	id            string
	host          string
	connected     bool
	login         bool
	startingChan  chan bool
	disconectChan chan bool

	conn *websocket.Conn
	w    http.ResponseWriter
	r    *http.Request
}

func NewClientConn(conn *websocket.Conn, w http.ResponseWriter, r *http.Request) {
	var id = ksuid.New().String()
	hid := clustermanager.GetInstance().Server.GetConf().Addr
	cc := &ClientConn{id: id, host: hid, conn: conn, w: w, r: r}
	cc.disconectChan = make(chan bool)
	cc.startingChan = make(chan bool)

}

func newcon(conn *websocket.Conn, w http.ResponseWriter, r *http.Request) {
	var id = ksuid.New().String()
	hid := clustermanager.GetInstance().Server.GetConf().Addr
	NewClientConn(conn, w, r)
	var open = true
	var user *user2.User
	login := false
	sess, e := sessionmanager.GlobalSessions.GetSession(w, r, false)
	if e != nil {
		log.Warning.Println(e)
		conn.Close()
		return

	} else {
		user, _ = userprovider.GetInstance().GetUserById(sess.Get("uid").(string))
		pk := packets.NewDUserPacket()
		pk.Us = user
		pk.Conn = conn
		pk.Id = id

		processor.Pross.UserIn <- pk

		login = true

		if perr := precensemanager.ProcenceManager.Provider.Add(user.Id, hid+"_"+id); perr != nil {
			log.Error.Printf("ADD %v", perr)
		}

		/*
			precensemanager.ProcenceManager.Provider.Add(user.Id,hid)
			hosts:=precensemanager.ProcenceManager.Provider.Get(user.Id)
			ms:=clustermanager.GetInstance().GetMembersBy(hosts)
			for _,m:= range ms{
				m.GetRpc().
			}*/
		//log.Warning.Println("session ", sess)
	}

	conn.SetCloseHandler(func(code int, text string) error {
		log.Warning.Println("se cerro "+id, code, text)
		open = false
		hid := clustermanager.GetInstance().Server.GetConf().Addr

		if perr := precensemanager.ProcenceManager.Provider.Remove(user.Id, hid+"_"+id); perr != nil {
			log.Error.Printf("REMOVE %v", perr)
		}

		return nil
	})

	c := conn
	fase := 1
	var err bool
	var uid string
	if login {
	}
	for {
		if !open {
			return
		}
		if fase == 1 {
			f1 := Fase1{}
			closeOnError(c.ReadJSON(&f1), c, &err)
			uid = f1.Uid
			if user.Id == f1.Uid && login {
				if f1.Fase == fase {
					f1r := Fase1r{Fase: f1.Fase, Uid: f1.Uid, Pased: true}
					closeOnError(c.WriteJSON(f1r), c, &err)
					log.Warning.Printf("%#v %#v\n", f1, f1r)
					fase++
				} else {
					d := struct {
						Fase  int
						Pased bool
						Msg   string
					}{
						Fase: fase,
						Msg:  "no match fase",
					}
					c.WriteJSON(d)
					c.Close()
				}
			} else {
				d := struct {
					Fase  int
					Pased bool
					Msg   string
				}{
					Fase: fase,
					Msg:  "no session",
				}
				c.WriteJSON(d)
				c.Close()
			}

		}

		if fase == 2 {
			_, _, errr := c.ReadMessage()
			closeOnError(errr, c, &err)
			d := struct {
				Fase  int    `json:"fase"`
				Pased bool   `json:"pased"`
				Msg   string `json:"msg"`
			}{
				Fase:  fase,
				Pased: true,
			}
			closeOnError(c.WriteJSON(d), c, &err)
			fase++
		}
		if fase == 3 {
			d := Fase3{}
			closeOnError(c.ReadJSON(&d), c, &err)
			if uid == d.Uid {
				if hosts, e := precensemanager.ProcenceManager.Provider.Get(user.Id); e == nil {
					log.Warning.Printf("HOST by %s %v", user.Id, hosts)
					ms := clustermanager.GetInstance().GetMembersBy(hosts)
					for _, m := range ms {
						args := new_user_conn.Args{user.Id + ":" + hid + ": " + d.Msg}
						resp := new_user_conn.Response{}
						m.GetRpc().Go("NewConnUser.NewConn", args, &resp, nil)
					}
				}

				log.Warning.Printf("%v %#v \n", fase, d)
				r := struct {
					Fase  int    `json:"fase"`
					Pased bool   `json:"pased"`
					Msg   string `json:"msg"`
				}{
					Fase:  fase,
					Pased: true,
					Msg:   d.Msg,
				}

				closeOnError(c.WriteJSON(r), c, &err)
			} else {
				r := struct {
					Fase  int    `json:"fase"`
					Pased bool   `json:"pased"`
					Msg   string `json:"msg"`
				}{
					Fase:  fase,
					Pased: false,
					Msg:   "uid faild",
				}
				closeOnError(c.WriteJSON(r), c, &err)
				err = true
				c.Close()
			}

		}

	}
}

type SendError struct {
	When time.Time
	What string
}

func (this *SendError) Error() string {

	return this.What

}

type hello struct {
	Msg string `json:"msg"`
}

const tpl string = `{{if .Login}}
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
</head>

<body>


<div>
    {{.User}}
</div>


<div>
    <input type="text" id="to">
</div>

<div>
    <textarea id="msg"></textarea>
</div>

<button onclick="conn.sendMsg()">Enviar</button>
<div id="logged">
    estas logeado
</div>
<div id="by_typing">
	Estan escribiendo ...
</div>
<div>
    <strong>payload</strong>
</div>

<div id="payload">

</div>
</body>
<script>
    class Conn {
        constructor(props) {
            this.props = props;
            this.byTyping = document.getElementById("by_typing");
            this.input = document.getElementById("uid");
            this.logged = document.getElementById("logged");
            this.to = document.getElementById("to");
            this.msg = document.getElementById("msg");
            this.msg.addEventListener("keypress", e => {
                if (e.keyCode == 13) {
                    this.sendMsg()
                }
            });
            this.logged.style.display = "none";
            this.byTyping.style.display = "none";
            this.payload = document.getElementById("payload");
            this.fase = 1;
            this.createConn()
            
            var inter = null
            var typing = false
            
            this.msg.onkeydown = e => {
               
                if (inter != null) {
                    clearInterval(inter)
                    if (!typing){
                        var uid = this.props.uid;
                        var to = this.to.value
                        var payload_init_typing = {
                            event:"init_typing",
                            by:uid,
                            to,
                            id_conversation: "cid",
                        }
                        this.soc.send("e!" + JSON.stringify(payload_init_typing, null, 2));
                        typing = true
                    }
                }

                inter = setInterval(_ => {
                    console.error("dejo de escribir")
                    var uid = this.props.uid;
                    var to = this.to.value
                    var payload_leave_typing = {
                        event:"leave_typing",
                        by:uid,
                        to,
                        id_conversation: "cid",
                    }
                    this.soc.send("e!" + JSON.stringify(payload_leave_typing, null, 2));
                    clearInterval(inter)
                    typing=false
                }, 600)

            }

        }

        logIn() {

            var uid = this.props.uid
            var payload = {fase: this.fase, uid};
            this.soc.send(JSON.stringify(payload, null, 2))

        }

        sendMsg() {
            if (this.fase == 3) {
                var msg = this.msg.value;
                var to = this.to.value
                var uid = this.props.uid;

                var packetMessage = {
                    message: msg,
                    by: uid,
                    to,
                    id_conversation: "cid",

                }

                this.soc.send("m!" + JSON.stringify(packetMessage, null, 2));
                this.msg.value = ""
            } else {
                alert("no te has conectado, " + this.fase)
            }
        }

        createConn() {
            this.soc = new WebSocket("ws://" + this.props.addr + "/ws");
            //ws.onmessage=e=>alert(e.data)
            this.soc.addEventListener("open", e => {
                //this.logIn()
                this.fase = 3
            })
            this.soc.addEventListener("close", e => {
                this.fase = 1;
                this.logged.style.display = "none"
            });
            this.soc.addEventListener("message", e => {
                var data = JSON.parse(e.data);
                
                if (data.fase == 1) {
                    if (data.pased) {
                        this.fase++;
                        var uid = this.props.uid;
                        var payload = {fase: this.fase, uid};
                        this.soc.send(JSON.stringify(payload, null, 2))

                    }
                }
                if (data.fase == 2) {
                    if (data.pased) {
                        this.fase++;
                        this.logged.style.display = "block"

                    }
                }
                if (data.fase == 3) {

                }
                if(data.kind=="event"){
                    if(data.event =="init_typing"){
                        this.byTyping.style.display="block"
                    }else{
                        this.byTyping.style.display="none"
                        
                    }
                }
                if (data.kind == "message"){
                    var child = document.createElement("div");
                    child.textContent = data.message;
                    if (this.payload.firstChild == null) {
                        this.payload.append(child)
                    } else {
                        this.payload.prepend(this.payload.firstChild, child)
                    }
                }
            })

        }
    }

    conn = new Conn({uid: {{ .User.Id }},
    addr:{{.Addr}}
    })
</script>
</html>
{{else}}
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
    login = () => {

        var user = f.elements.username;
        var pass = f.elements.password;

        var xhr = new XMLHttpRequest();
        xhr.withCredentials = true
        var data = new FormData();
        data.append("user", user.value);
        data.append("pass", pass.value);
        xhr.open("post", "http://{{ .RestAddr }}/api/user/login");
        xhr.onload = (e) => {
            var data = JSON.parse(e.target.response);
            if (data.Loging) {
                location.reload()
            } else {
                alert(data.msg)
            }
        };
        xhr.send(data);

        return false
    }
</script>
</html>
{{end}}

`

var tplHome, _ = template.New("nc").Parse(tpl)
