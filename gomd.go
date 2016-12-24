package main

import (
  "fmt"
  "io/ioutil"
  "github.com/fsnotify/fsnotify"
  "github.com/gorilla/websocket"
  "github.com/russross/blackfriday"
  "github.com/dron22/gomd/html"
  "log"
  "net/http"
  "os"
  "strings"
  "time"
)

var host = "127.0.0.1"
var port = 6419
var renderInterval = time.Millisecond * 100

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024,
}
var connections []*websocket.Conn

func watchFile(filepath string) (chan bool, error) {
    changed := make(chan bool)
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return changed, err
    }

    go func(changed chan bool) {
        for {
            select {
            case event := <-watcher.Events:
                switch {
                case event.Op&fsnotify.Write == fsnotify.Write:
                    changed <- true
                case event.Op&fsnotify.Remove == fsnotify.Remove:
                    log.Println("file removed:", event.Name)
                    return
                case event.Op&fsnotify.Rename == fsnotify.Rename:
                    log.Println("file renamed:", event.Name)
                    return
                }
            case err := <-watcher.Errors:
                log.Println("error:", err)
            }
        }
    }(changed)

    err = watcher.Add(filepath)
    if err != nil {
        return changed, err
    }
    return changed, nil
}

func renderPage(filepath string) ([]byte, error) {
    input, err := ioutil.ReadFile(filepath)
    if err != nil {
        fmt.Printf("error reading file", err)
        return nil, err
    }

    output := blackfriday.MarkdownBasic(input)
    return output, nil
}

func sendRenderedPage(filepath string, connections []*websocket.Conn) error {
    html, err := renderPage(filepath)
    if err != nil {
        return err
    }
    for _, conn := range connections {
      if err := conn.WriteMessage(1, html); err != nil {
        return err
      }
    }
    return nil
}
func forwardMessageLoop(filepath string) {
    var err error
    changed, err := watchFile(filepath)
    if err != nil {
        log.Fatalf("error watching file: %v", err)
    }

    for {
        _ = <- changed
        err = sendRenderedPage(filepath, connections)
        if err != nil {
            log.Fatalf("error render/sending page: %v", err)
        }
    }
}

func GetWebSocketHandler(filepath string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
          log.Println(err)
          return
        }
        err = sendRenderedPage(filepath, []*websocket.Conn{conn})
        if err != nil {
            log.Fatalf("error render/sending page: %v", err)
        }
        connections = append(connections, conn)
    }
}

func PageHandler(w http.ResponseWriter, r *http.Request) {
    html := html.GetHTML(host, port)
    fmt.Fprintln(w, html)
}

func main() {
    args := os.Args
    splitted := strings.Split(args[0], "/")
    pgm := splitted[len(splitted)-1]
    if len(args) != 2 {
        fmt.Printf("Usage: %v /path/to/file.md\n", pgm)
        os.Exit(1)
    }

    filepath := args[1]
    if _, err := os.Stat(filepath); err != nil {
        fmt.Printf("file %v does not exist", filepath)
        os.Exit(1)
    }

    go forwardMessageLoop(filepath)

    fmt.Printf("Server running on http://%v:%v/\n", host, port)
    http.HandleFunc("/", PageHandler)
    http.HandleFunc("/ws", GetWebSocketHandler(filepath))
    http.ListenAndServe(fmt.Sprintf("%v:%v", host, port), nil)
}
