package main

import (
  "fmt"
  "io/ioutil"
  "github.com/howeyc/fsnotify"
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
var rewatchTimeout = time.Second * 5

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

    // Process events
    go func() {
        for {
            select {
            case ev := <-watcher.Event:
                switch {
                case ev.IsModify():
                    changed <- true
                case ev.IsDelete():
                    // certain texteditors like vim don't modify but delete and replace a file upon save. Thus Watch() has to be called again to continue watching after the delete.
                    start := time.Now()
                    for {
                        err = watcher.Watch(filepath)
                        switch {
                        case time.Now().Sub(start) > rewatchTimeout:
                            log.Fatalf("File deleted and not recreated within timeout: %v", filepath)
                            return
                        case err != nil:
                            continue
                        }
                        break
                    }
                    changed <- true
                }
            case err := <-watcher.Error:
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Watch(filepath)
    if err != nil {
        return changed, err
    }
    return changed, nil
}

func renderPage(filepath string) ([]byte, error) {
    input, err := ioutil.ReadFile(filepath)
    if err != nil {
        log.Printf("error reading file: %v\n", err)
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
        log.Fatalf("error watching file: %v\n", err)
    }

    for {
        _ = <- changed
        err = sendRenderedPage(filepath, connections)
        if err != nil {
            log.Fatalf("error render/sending page: %v\n", err)
        }
    }
    log.Fatal("Exited forwardMessageLoop")
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
            log.Fatalf("error render/sending page: %v\n", err)
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
        fmt.Printf("file %v does not exist\n", filepath)
        os.Exit(1)
    }

    go forwardMessageLoop(filepath)

    fmt.Printf("Server running on http://%v:%v/\n", host, port)
    http.HandleFunc("/", PageHandler)
    http.HandleFunc("/ws", GetWebSocketHandler(filepath))
    log.Fatal(http.ListenAndServe(fmt.Sprintf("%v:%v", host, port), nil))
}
