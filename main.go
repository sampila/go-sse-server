// main.go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"
)

func formatSSE(event string, data string) []byte {
    eventPayload := "event: " + event + "\n"
    dataLines := strings.Split(data, "\n")
    for _, line := range dataLines {
        eventPayload = eventPayload + "data: " + line + "\n"
    }
    return []byte(eventPayload + "\n")
}

var messageChannels = make(map[chan []byte]bool)

type SayRequest struct{
  Data map[string]string
}

func sayHandler(w http.ResponseWriter, r *http.Request) {
    // Declare a new Person struct.
    var p SayRequest

    // Try to decode the request body into the struct. If there is an error,
    // respond to the client with the error message and a 400 status code.
    err := json.NewDecoder(r.Body).Decode(&p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    jsonStructure, _ := json.Marshal(p.Data)

    go func() {
        for messageChannel := range messageChannels {
            messageChannel <- []byte(jsonStructure)
            log.Println(string(jsonStructure))
        }
    }()

    w.Write([]byte("ok."))
}

func listenHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    _messageChannel := make(chan []byte)
    messageChannels[_messageChannel] = true

    for {
        select {
        case _msg := <-_messageChannel:
            w.Write(formatSSE("message", string(_msg)))
            w.(http.Flusher).Flush()
            log.Println(string(_msg))
        case <-r.Context().Done():
            delete(messageChannels, _messageChannel)
            return
        }
    }
}

func main() {
    http.HandleFunc("/say", sayHandler)
    http.HandleFunc("/listen", listenHandler)

    log.Println("Running at :9011")
    log.Fatal(http.ListenAndServe(":9011", nil))
}
