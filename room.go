package main

import "time"

type RoomInfo struct {
	Id        string
	Name      string
	ClientNum uint
	MsgNum    uint

	StartAt time.Time
}
type Room struct {
	hub *Hub

	info    *RoomInfo
	clients map[*Client]struct{}

	broadcast chan []byte
	enter     chan *Client
	leave     chan *Client
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.enter:
			r.clients[client] = struct{}{}
			r.info.ClientNum++
		case client := <-r.leave:
			//TODO check client exists?
			delete(r.clients, client)
			close(client.send)
			r.info.ClientNum--
			if r.info.ClientNum == 0 {
				r.hub.deleteRoom(r.info.Id)
				return
			}
		case msg := <-r.broadcast:
			r.info.MsgNum++
			for client := range r.clients {
				select {
				case client.send <- msg:
				default:
					delete(r.clients, client)
					close(client.send)
					r.info.ClientNum--
					if r.info.ClientNum == 0 {
						r.hub.deleteRoom(r.info.Id)
						return
					}
				}
			}
		}
	}
}
