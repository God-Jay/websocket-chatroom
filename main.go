// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8081", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()

	hub := newHub()

	r := mux.NewRouter()

	r.HandleFunc("/", serveHome)

	r.HandleFunc("/ws/{roomId}", func(w http.ResponseWriter, r *http.Request) {
		hub.clientEnter(w, r)
	})

	r.HandleFunc("/roomInfo", func(w http.ResponseWriter, r *http.Request) {
		roomInfo := hub.getRoomInfo()
		marshal, _ := json.Marshal(roomInfo)
		w.Write(marshal)
	})

	http.Handle("/", r)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
