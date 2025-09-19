package main

import (
	"Draftly/WS/internal"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

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
	Managers sync.Map // roomID -> *roomManager
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	id := mux.Vars(r)["roomID"]
	log.Printf("Client connected to room: %s", id)
	_, err = internal.GetDocument(id)
	if err != nil {
		log.Println("Error getting document:", err)
		conn.WriteJSON(map[string]string{"error": "Error getting document", "details": err.Error()})
		conn.Close()
		return
	}
	// Upgrade initial GET request to a websocket
	_ = GetRoomManager(id)
	defer conn.Close()
	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		var inputOperation Operation
		err = json.Unmarshal(message, &inputOperation)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "Invalid operation format", "input": string(message), "error_details": err.Error()})
			continue
		}
		log.Printf("Received: %v", inputOperation)
		// Now you perform the operation using OT logic
		// TODO:
		// once complete write this out to the postgress database
		// TODO:

		// return ack to the client
		ack := map[string]string{
			"status":  "acknowledged",
			"message": fmt.Sprintf("Operation %s at position %f with text '%s' processed", inputOperation.Kind, inputOperation.Position, inputOperation.Text),
		}
		err = conn.WriteJSON(ack)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
func routes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", HealthCheckHandler)
	r.HandleFunc("/ws/{roomID}", webSocketHandler)
	return r
}

func main() {
	fmt.Printf("server running on port :%s\n", cfg.WSPort)
	if err := http.ListenAndServe(":"+cfg.WSPort, routes()); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

type Operation struct {
	Kind     string  `json:"kind"` // cant use type as field name because its a reserved word
	Position float64 `json:"position"`
	Text     string  `json:"text"`
}

type roomManager struct {
	roomMembers sync.Map // roomID -> map[conn]bool
	// use roomMembers to keep track of active connections in each room
}

func (rm *roomManager) addMember(roomID string, conn *websocket.Conn) {
	rm.roomMembers.Store(conn, true)
}

func (rm *roomManager) removeMember(roomID string, conn *websocket.Conn) {
	rm.roomMembers.Delete(conn)
}

func (rm *roomManager) roomCount(roomID string) int {
	count := 0
	rm.roomMembers.Range(func(k, v interface{}) bool {
		if v.(bool) {
			count++
		}
		return true
	})
	return count
}
func (rm *roomManager) pingClients(roomID string) {
	// write to clients, if an error or no response set their status to false

}
func GetRoomManager(roomID string) *roomManager {
	v, ok := Managers.Load(roomID)
	if ok {
		return v.(*roomManager)
	}
	rm := &roomManager{}
	Managers.Store(roomID, rm)
	return rm
}
func RoomStats() {
	Managers.Range(func(k, v interface{}) bool {
		fmt.Printf("RoomID: %s, Manager: %v\n", k, v)
		return true
	})
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
(4). Add metaData about rooms (# of users) so that when theres no more users we can send a message to the compaction service to save the file ->
to do this we need to keep track of each users when they join, have a heartbreat to check if their still there and then on checks for heartbreat if no one responds the room is considered closed -> DONE


*/
