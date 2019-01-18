package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Davidc2525/messager/cluster/clustermanager"
	"github.com/Davidc2525/messager/core/restapi"
	"github.com/Davidc2525/messager/core/sessionmanager"
	"github.com/Davidc2525/messager/log"
	user2 "github.com/Davidc2525/messager/user"
	"github.com/Davidc2525/messager/userprovider"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
)

var (
	log        = mlog.New()
	HostByUser map[string][]string
)

func sendJson(w http.ResponseWriter, data interface{}) {
	b, _ := json.Marshal(data)

	buf := bytes.Buffer{}
	json.Indent(&buf, b, "", "	")
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
func init() {
	HostByUser = make(map[string][]string)
	fmt.Println("iniciando api user")

	restapi.Register2("/apiv2/user", func(user *mux.Router) {

		user.HandleFunc("", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("api user v2"))
		})
	})

	restapi.Register2("/api/user/", func(user *mux.Router) {
		user.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			out := `
	logout
	login
	get
	get/{uid}
	home
			`

			writer.Write([]byte(out))
		})
		user.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			s, e := sessionmanager.GlobalSessions.GetSession(w, r, false)
			if e == nil {
				hid := clustermanager.GetInstance().Server.GetConf().Addr
				uid := s.Get("uid").(string)
				sessionmanager.GlobalSessions.SessionDestroy(w, r)
				//delete(HostByUser,)
				hs := HostByUser[uid]
				hs = remove(hs, pos(hs, hid))
				if len(hs) == 0 {
					delete(HostByUser, uid)
				} else {
					HostByUser[uid] = hs
				}

			}
			fmt.Fprintf(w, "logout <a href='/api/user/get'>user</a>")

		})

		user.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")
			//re, _ := r.MultipartReader()
			r.ParseMultipartForm(1024)

			//f, _ := re.ReadForm(1024)
			usern := r.FormValue("user")
			//pass := r.FormValue("pass")
			u, _ := userprovider.GetInstance().GetUserById(usern)
			hid := clustermanager.GetInstance().Server.GetConf().Addr
			if u != nil {
				s, _ := sessionmanager.GlobalSessions.GetSession(w, r, true)
				s.Set("uid", u.Id)
				//s.Set("pass", pass)

				hs := HostByUser[u.Id]
				hs = append(hs, hid)
				HostByUser[u.Id] = hs

				sendJson(w, struct {
					Name   string `json:"name"`
					Id     string `json:"id"`
					Loging bool
				}{
					Loging: true,
					Name:   u.Name,
					Id:     u.Id,
				})
			} else {

				sendJson(w, struct {
					Loging bool
					Msg    string `json:"msg"`
				}{
					Loging: false,
					Msg:    "no existe usuario ",
				})
			}

			//fmt.Fprintf(w,"login %s, %#v, <a href='/api/user/get'>user</a>",usern,s)

		}).Methods("POST")
		user.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			tplForm.Execute(w, nil)
		}).Methods("GET")

		user.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {

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
				w.Header().Set("content-type", "application/json")
				u, _ := userprovider.GetInstance().GetUserById(s.Get("uid").(string))
				if u == nil {
					sendJson(w, struct {
						Status bool   `json:"status" xml:"status"`
						Msg    string `json:"msg" xml:"msg"`
					}{
						Status: false,
						Msg:    "User no exist",
					})

				} else {
					sendJson(w, struct {
						Status bool       `json:"status" xml:"status"`
						User   user2.User `json:"user" xml:"user"`
					}{
						Status: true,
						User:   *u,
					})

				}

			}

		})

		user.HandleFunc("/get/{user}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")

			u, _ := userprovider.GetInstance().GetUserById(mux.Vars(r)["user"])
			if u == nil {
				w.WriteHeader(http.StatusNotFound)
				sendJson(w, struct {
					Status bool   `json:"status" xml:"status"`
					Msg    string `json:"msg" xml:"msg"`
				}{
					Status: false,
					Msg:    "User no exist",
				})
			} else {
				sendJson(w, struct {
					Status bool       `json:"status" xml:"status"`
					User   user2.User `json:"user" xml:"user"`
				}{
					Status: true,
					User:   *u,
				})

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
