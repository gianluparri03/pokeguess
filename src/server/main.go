package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)


type Party struct {
    Player   []Player
}

type Player struct {
    Username string
    Conn     *websocket.Conn
    Party    Party
}


var Players map[string]*Player
var Upgrader websocket.Upgrader


func Connect(w http.ResponseWriter, r *http.Request) {
    // Gets the player's username
    username := mux.Vars(r)["username"]

    // Upgrades the connection to a websocket (and checks for errors)
    conn, err := Upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("> ERROR during connection: %s", err.Error())
    }

    // Checks if the username is used
    if _, found := Players[username]; found {
        message := websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "username already used")
        conn.WriteMessage(websocket.CloseMessage, message)
        conn.Close()
        return
    }

    // Saves the player
    player := &Player{Conn: conn, Username: username}
    Players[username] = player

    // Welcomes the new player
    log.Printf("> %s just connected", username)
    conn.WriteMessage(websocket.TextMessage, []byte("connected"))

    // Ensures the player will be delete
    defer conn.Close()
    defer delete(Players, username)


    // Endlessly listens to the player's messages
    for {
        _, msg, err := player.Conn.ReadMessage()

        if websocket.IsUnexpectedCloseError(err) {
            log.Printf("> %s just disconnected", username)
            break
        } else if err != nil {
            log.Printf("> ERROR reading a message from %s: %s", username, err.Error())
            break
        } else {
            log.Printf("> new message from %s: %s", username, msg)
        }
    }
}


func main() {
    Players = make(map[string]*Player)

    // Creates a new router
    router := mux.NewRouter()
    router.HandleFunc("/connect/{username}", Connect)

    // Starts the server
    log.Println("Started PokéGuess server")
    http.ListenAndServe(":8080", router)
}
