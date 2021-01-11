package main

import (
    "github.com/rivo/tview"
    "github.com/gorilla/websocket"
)


const server = "localhost:8080/"
var conn *websocket.Conn
var app *tview.Application = tview.NewApplication()


func ShowMessage(message string) {
    // Creates the modal
    modal := tview.NewModal()
    modal.SetText(message)
    modal.AddButtons([]string{"Ok"})
    modal.SetDoneFunc(func (i int, l string) { app.Stop() })

    // Shows it
    app.SetRoot(modal, false)
}

func setUsername(username string) {
    // Tries to create a connection
    var err error
    conn, _, err = websocket.DefaultDialer.Dial("ws://" + server + "connect/" + username, nil)

    // Shows eventual errors
    if err != nil {
        ShowMessage(err.Error())
        return
    }

    // Ensure that the connection will be closed
    defer conn.Close()

    // Shows the server's response
    if _, msg, err := conn.ReadMessage(); err != nil {
        ShowMessage(err.Error())
    } else {
        ShowMessage(string(msg))
    }

    // Close the connection gently
    conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""))
}

func main() {
    // Creates the form
    form := tview.NewForm()
    form.SetBorder(true)
    form.SetTitle("Insert an username")
    form.AddInputField("Username:", "", 20, nil, nil)
    form.AddButton("Connect", func () { setUsername(form.GetFormItem(0).(*tview.InputField).GetText()) })
    form.AddButton("Quit", app.Stop)

    // Shows it
    app.SetRoot(form, true).Run()
}
