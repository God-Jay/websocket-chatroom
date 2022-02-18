// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	//rooms map[string]*Room
	rooms sync.Map

	startAt time.Time
}

func newHub() *Hub {
	return &Hub{startAt: time.Now()}
}

func (h *Hub) getRoom(roomId string) *Room {
	if room, ok := h.rooms.Load(roomId); ok {
		return room.(*Room)
	}
	room := h.newRoom(roomId)
	h.rooms.Store(roomId, room)
	return room
}

func (h *Hub) newRoom(roomId string) *Room {
	info := &RoomInfo{
		Id:      roomId,
		Name:    "room" + roomId,
		StartAt: time.Now(),
	}
	room := &Room{
		hub:       h,
		info:      info,
		clients:   make(map[*Client]struct{}),
		broadcast: make(chan []byte),
		enter:     make(chan *Client),
		leave:     make(chan *Client),
	}
	go room.run()
	return room
}

func (h *Hub) deleteRoom(roomId string) {
	h.rooms.Delete(roomId)
}

// clientEnter handles websocket requests from the peer.
func (h *Hub) clientEnter(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	vars := mux.Vars(r)
	room := h.getRoom(vars["roomId"])

	client := &Client{room: room, conn: conn, send: make(chan []byte, 256)}
	room.enter <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

type ReportRoomInfo struct {
	RoomName  string `json:"room_name"`
	ClientNum uint   `json:"client_num"`
	MsgNum    uint   `json:"msg_num"`
	StartAt   string `json:"start_at"`
}

func (h *Hub) getRoomInfo() []ReportRoomInfo {
	var roomInfo []ReportRoomInfo
	h.rooms.Range(func(key, value interface{}) bool {
		room := value.(*Room)
		roomInfo = append(roomInfo, ReportRoomInfo{
			RoomName:  room.info.Name,
			ClientNum: room.info.ClientNum,
			MsgNum:    room.info.MsgNum,
			StartAt:   room.info.StartAt.Format("2006-01-02 15:04:05"),
		})
		return true
	})
	return roomInfo
}
