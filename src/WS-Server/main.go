package main

import (
	"Draftly/WS/internal"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var (
	cfg      = internal.NewConfig()
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	manager Managers // roomID -> *roomManager
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	resp := map[string]interface{}{
		"status": "WS Server is Live",
		"time":   fmt.Sprint(time.Now().Format(time.RFC3339)),
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error generating JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	id := mux.Vars(r)["roomID"]
	userName := r.URL.Query().Get("username")
	if userName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing username"))
		return
	}
	// Upgrade initial GET request to a websocket

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	m := manager.GetRoomManager(id)
	m.initClient(conn, userName)
	defer conn.Close()
	defer m.removeMember(conn)
	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				// Normal closure, just clean up
				m.removeMember(conn)
				return
			}
			log.Println("Read error:", err)
			break
		}
		var inputOperation internal.Operation
		err = json.Unmarshal(message, &inputOperation)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "Invalid operation format", "input": string(message), "error_details": err.Error()})
			continue
		}
		err = inputOperation.Validate()
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "Operation validation failed", "details": err.Error()})
			continue
		}
		log.Printf("Received: %v", inputOperation)
		// Now you perform the operation using OT logic
		// TODO:

		outputOperation := m.Apply(inputOperation)

		// write this out to the postgress database
		ts := time.Now()
		w, err := internal.NewWriteStore()
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "Failed to initialize storage", "details": err.Error()})
			continue
		}
		err = w.WriteOperation(id, inputOperation, ts)

		if err != nil {
			conn.WriteJSON(map[string]string{"error": "Failed to write operation", "details": err.Error()})
			continue
		}
		// process the input and stream it to everyone
		output := map[string]interface{}{
			"type":      "operation",
			"ts":        ts.Format(time.RFC3339),
			"operation": outputOperation,
		}
		fmt.Printf("broadcasting: %v to all connected clients in room %s\n", output, m.roomID)
		m.broadcast(output, conn) // broadcast to other members

	}
}
func routes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", HealthCheckHandler)
	// ?username=richard
	r.HandleFunc("/ws/{roomID}", webSocketHandler)
	return r
}

func main() {
	fmt.Printf("server running on port :%s\n", cfg.WSPort)
	go manager.roomCount()
	if err := http.ListenAndServe(":"+cfg.WSPort, routes()); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

type Managers struct {
	roomMembers sync.Map // roomID -> *roomManager
}

type wsManager struct {
	roomID      string
	roomMembers sync.Map // conn -> bool (active status)
	// use roomMembers to keep track of active connections in each room
	lastUpdate   map[string]time.Time // conn/IP -> last update time
	connUsername map[*websocket.Conn]string
	// Operation transform Management
	Ops     []internal.Operation
	Version int
}

func (ws *wsManager) Apply(op internal.Operation) internal.Operation {
	for i := op.Version; i < ws.Version; i++ {
		newer := ws.Ops[i]

		if newer.Kind != "insert" {
			continue
		}

		//  Only shift if the newer insert is strictly before your op
		if newer.Position < op.Position {
			op.Position += utf8.RuneCountInString(newer.Text)
		}
	}
	return op
}

func (ws *wsManager) initClient(conn *websocket.Conn, userName string) {
	// TODO: read updates from postgress and send to client and then add them to the room and treat them as any other client
	if since, ok := ws.lastUpdate[userName]; ok {
		fmt.Printf("Client %s reconnected, sending updates since %v\n", userName, since)
		w, err := internal.NewWriteStore()
		if err != nil {
			log.Println("Error initializing storage:", err)
			return
		}
		ops, err := w.OperationsSince(ws.roomID, since)
		if err != nil {
			log.Println("Error fetching operations since last update:", err)
			return
		}
		response := map[string]interface{}{
			"type":       "history",
			"since":      since.Format(time.RFC3339),
			"operations": ops,
		}
		conn.WriteJSON(response)
	}
	ws.connUsername[conn] = userName
	ws.addMember(conn, userName)
}

func (ws *wsManager) addMember(conn *websocket.Conn, userName string) {
	ws.roomMembers.Store(conn, true)
	ws.lastUpdate[userName] = time.Now()
}

func (ws *wsManager) removeMember(conn *websocket.Conn) {
	ws.roomMembers.Delete(conn)
	ws.checkEmpty()
}

func (ws *wsManager) checkEmpty() {
	count := 0
	ws.roomMembers.Range(func(k, v interface{}) bool {
		if v.(bool) {
			count++
		}
		return true
	})
	if count == 0 {
		log.Println("Room is empty, performing cleanup")
		_ = closeRoomRequest(ws.roomID)
	}
}

func (ws *wsManager) broadcast(message interface{}, _ *websocket.Conn) {
	ws.roomMembers.Range(func(k, v interface{}) bool {
		conn := k.(*websocket.Conn)
		//if conn != sender { // don't send the message back to the sender

		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("Broadcast error:", err)
			conn.Close()
			ws.removeMember(conn)
		}

		//}
		k1, ok := ws.connUsername[conn]
		if ok {
			ws.lastUpdate[k1] = time.Now()
		}
		return true
	})
}

func (m *Managers) GetRoomManager(roomID string) *wsManager {
	v, ok := m.roomMembers.Load(roomID)
	if ok {
		return v.(*wsManager)
	}
	rm := &wsManager{
		roomID:       roomID,
		roomMembers:  sync.Map{},
		lastUpdate:   make(map[string]time.Time),
		connUsername: make(map[*websocket.Conn]string),
	}
	m.roomMembers.Store(roomID, rm)
	return rm
}

func (m *Managers) roomCount() {
	for {
		m.roomMembers.Range(func(k, v interface{}) bool {
			roomID := k.(string)
			rm := v.(*wsManager)
			count := 0
			rm.roomMembers.Range(func(k, v interface{}) bool {
				if v.(bool) {
					count++
				}
				return true
			})
			log.Printf("Room %s has %d active members", roomID, count)
			return true
		})

		time.Sleep(30 * time.Second)
	}

}

func closeRoomRequest(roomID string) error {
	// TODO: send a request to the compaction service (CRUD) to save the file and then delete the room from memory
	// dont forget to remove from manager
	// delete the roomManager -> cleans up connections to usernames and stuff easier than doing one by one
	return nil
}

/*

all operations go to the server
server sets the timestamp for each operation
if theres a race condition the server always wins
server will create a tranformation and then send it out to the clients

push all the work to the browser

for each change sent by the browser we acknowledge it
now the browser knows to add this to the doc


Server orders the operations
Updates all the clients with the order of operations

*/

/*
TODO :
(1). Read in content from S3  as a string -> DONE
(2). When a client joins a ws/{roomID} create a helper function to return the latest content from in memeory representation of that file -> DONE
(3). add Error checking to make sure the shape of all the request from the front end is an Operation -> DONE
(4). Add metaData about rooms (# of users) so that when theres no more users we can send a message to the compaction service to save the file -> Done
to do this we need to keep track of each users when they join, have a heartbreat to check if their still there and then on checks for heartbreat if no one responds the room is considered closed -> DONE


*/
