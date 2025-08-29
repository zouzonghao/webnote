package websocket

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	clients    map[*Client]bool
	Broadcast  chan *BroadcastMsg
	register   chan *Client
	unregister chan *Client
	// Group clients by path
	paths map[string]map[*Client]bool
}

type BroadcastMsg struct {
	Path    string
	Content []byte
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *BroadcastMsg),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		paths:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if h.paths[client.path] == nil {
				h.paths[client.path] = make(map[*Client]bool)
			}
			h.paths[client.path][client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if h.paths[client.path] != nil {
					delete(h.paths[client.path], client)
					if len(h.paths[client.path]) == 0 {
						delete(h.paths, client.path)
					}
				}
			}
		case message := <-h.Broadcast:
			if clients, ok := h.paths[message.Path]; ok {
				for client := range clients {
					select {
					case client.send <- message.Content:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(h.paths[message.Path], client)
					}
				}
			}
		}
	}
}
