package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	connections []*websocket.Conn
)

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("err ws %v\n", err)
	}
	connections = append(connections, conn)

	for {
		log.Printf("connections: %v", len(connections))
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("err msg %v\n", err)
			return
		}
		log.Printf("type: %v, msg: %v\n", msgType, string(msg))
		for i, c := range connections {
			if err := c.WriteMessage(msgType, msg); err != nil {
				log.Println("connection closed")
				connections = append(connections[:i], connections[i+1:]...)
			}
		}
	}
}

func auth(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Request from : %v", req.RemoteAddr)
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		u, p, ok := req.BasicAuth()
		if !ok {
			log.Println("Error parsing basic auth")
			w.WriteHeader(401)
			return
		}
		if u == "user" && p == "password" {
			h(w, req)
		} else {
			log.Println("wrong basic auth")
			w.WriteHeader(401)
			return
		}
	}
}

func main() {
	serv := http.FileServer(http.Dir("static"))
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/", auth(serv.ServeHTTP))

	if err := http.ListenAndServeTLS(":1337", "/path/to/fullchain.pem", "/path/to/privkey.pem", nil); err != nil {
		log.Fatalln("Coud not start server")
	}

}
