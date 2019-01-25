package restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Davidc2525/messager/log"
	"github.com/gorilla/mux"
	"net/http"
)

var (
	log  = mlog.New()
	s    = mux.NewRouter()
	REST *RestApi
)

type RestApi struct {
	Conf Conf
}
type Conf struct {
	Addr string
}

func New(conf Conf) *RestApi {
	REST = &RestApi{Conf: conf}
	return REST
}

func Register2(pattern string, fn func(*mux.Router)) {
	log.Warning.Println("register pattern", pattern)
	r := s.PathPrefix(pattern).Subrouter()
	fn(r)
}

func Register(pattern string, r http.Handler) {
	log.Warning.Println("api register", pattern)
	s.Handle(pattern, r)

}

func (this *RestApi) Start() {
	fmt.Println("starting restapi")

	s.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

		writer.Write([]byte("Debe espesificar un api"))
	})

	s.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("origin"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			// Do stuff here
			log.Warning.Println(r.RequestURI)
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	})

	go func() {
		err := http.ListenAndServe(this.Conf.Addr, s)
		if err != nil {
			log.Error.Panic(err)
		}
	}()
}

func sendJson(w http.ResponseWriter, data interface{}) {

	b, _ := json.Marshal(data)

	buf := bytes.Buffer{}
	json.Indent(&buf, b, "", "	")
	w.Write(buf.Bytes())
}

func init() {
	fmt.Println("initiantin restapi")

}
