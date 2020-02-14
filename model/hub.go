// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package model

import (
	"encoding/json"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan webSocketRequest

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// NewHub generate a new hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan webSocketRequest),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

//Run means new hub start work
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:

			for client := range h.clients {

				messageBytes, err := json.Marshal(message.Data)
				handleError(err)

				select {
				case client.send <- messageBytes:
				default:
					close(client.send)
					delete(h.clients, client)
				}

				if message.Data["channel"] == "a" {
					select {
					case client.send <- messageBytes:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}

			}
		}
	}
}
