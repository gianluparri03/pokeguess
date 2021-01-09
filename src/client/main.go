package main

import (
    "fmt"
    "time"
    "github.com/gosuri/uilive"
    "github.com/gorilla/websocket"
)


var Username string
var Conn *websocket.Conn
var Writer *uilive.Writer = uilive.New()


func main() {
    // Prepares the UI
    Writer.Start()
    defer Writer.Stop()

    // Asks for an username
    fmt.Fprintf(Writer, "Insert your username:\n> ")
    fmt.Scanf("%s", &Username)

    // Tires to create a websocket
    var err error
    Conn, _, err = websocket.DefaultDialer.Dial("ws://localhost:8080/connect/" + Username, nil)
    defer Conn.Close()

    // Checks for errors
    if err != nil {
        fmt.Fprintln(Writer, err)
        return
    }

    // Prints a connecting message
    fmt.Fprintln(Writer, "Connecting...")
    time.Sleep(time.Second) // Eheheheeh

    // Waits the server for a reply
    if msgType, msg, err := Conn.ReadMessage(); err != nil {
        fmt.Fprintln(Writer, err)
    } else if msgType == websocket.CloseMessage {
        fmt.Fprintln(Writer, string(msg))
    } else {
        fmt.Fprintln(Writer, string(msg))
    }

    // Closes the connection
    Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""))
}
