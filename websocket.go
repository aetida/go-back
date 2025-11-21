package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var socketMutex = &sync.Mutex{}
var sockets = make(map[string]map[string]*websocket.Conn) // sessionID -> {"client": conn, "phone": conn}

func CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	session := Session{SessionID: uuid.New().String()}
	DB.Create(&session)

	socketMutex.Lock()
	sockets[session.SessionID] = make(map[string]*websocket.Conn)
	socketMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"sessionId":"%s"}`, session.SessionID)))
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	role := r.URL.Query().Get("type")
	if sessionID == "" || (role != "client" && role != "phone") {
		http.Error(w, "invalid session or type", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	socketMutex.Lock()
	if _, ok := sockets[sessionID]; !ok {
		sockets[sessionID] = make(map[string]*websocket.Conn)
	}
	sockets[sessionID][role] = conn
	socketMutex.Unlock()

	fmt.Printf("[%s] %s connected\n", sessionID, role)

	// обновляем DB
	var sess Session
	DB.First(&sess, "session_id = ?", sessionID)
	if role == "client" {
		sess.ClientConnected = true
	} else {
		sess.PhoneConnected = true
	}
	DB.Save(&sess)

	// Если оба подключены — уведомляем
	socketMutex.Lock()
	if sockets[sessionID]["client"] != nil && sockets[sessionID]["phone"] != nil {
		sendSafe(sockets[sessionID]["client"], map[string]string{"event": "paired"})
		sendSafe(sockets[sessionID]["phone"], map[string]string{"event": "paired"})
		fmt.Printf("[%s] PAIRED\n", sessionID)
	}
	socketMutex.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("[%s] %s disconnected\n", sessionID, role)
			break
		}

		socketMutex.Lock()
		target := sockets[sessionID]
		var other *websocket.Conn
		if role == "client" {
			other = target["phone"]
		} else {
			other = target["client"]
		}
		socketMutex.Unlock()

		if other != nil {
			sendSafe(other, msg)
		}
	}

	socketMutex.Lock()
	delete(sockets[sessionID], role)
	socketMutex.Unlock()
}

func sendSafe(ws *websocket.Conn, msg interface{}) {
	if ws == nil {
		return
	}
	var data []byte
	switch v := msg.(type) {
	case []byte:
		data = v
	default:
		data, _ = json.Marshal(v)
	}
	ws.WriteMessage(websocket.TextMessage, data)
}
