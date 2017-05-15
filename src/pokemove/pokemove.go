package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "os"

  "github.com/chromatixau/negroni"
  "github.com/chromatixau/gomiddleware"
)

func main() {
  logfilename := "log/pokemove.log"
  errorLog, err := os.OpenFile( logfilename , os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666 )
  if err != nil {
    log.Fatal( "error writing to log: " + logfilename )
  }
  defer errorLog.Close()

  n := negroni.New()
  l := gomiddleware.NewLoggerWithStream( errorLog )
  r := negroni.NewRecovery()
  r.Logger = l
  r.PrintStack = false
  m := http.NewServeMux()

  handleRoutes( m )

  n.Use(r)
  n.Use(l)
  n.UseHandler( m )

  log.Fatal( http.ListenAndServe( ":18885", n ) )
}

func handleRoutes( m *http.ServeMux ) {

  m.HandleFunc( "/", func( w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
      text := r.URL.Path[1:]
      if text == "" {
        text = r.URL.Query().Get( "text" )
      }
      url := "https://hooks.slack.com/services/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"
      channel := r.URL.Query().Get( "channel_id" )
      if text == "" {
        fmt.Fprintf( w, "%s", about() )
        w.WriteHeader( http.StatusOK )
      } else {
        go sendMoveInfo( &url, &channel, &text )
        w.WriteHeader( http.StatusOK )
      }
    }
  })
}

func about() ( s string ) {
  s = fmt.Sprint( "Just type your move after the /pokemove command" )
  return
}

type Message struct {
  Text string `json:"text"`
  Channel string `json:"channel"`
}

func sendMoveInfo( url *string, channel *string, move_name *string ) {
  s := fmt.Sprintf( "%s", *move_name )
  m := Message{ s, *channel }
  message, err := json.Marshal( m )
  if err != nil {
    panic( "json marshall failed" )
  }
  _ = sendToSlack( url, message )
}

func sendToSlack( url *string, message []byte) (resp *http.Response) {
  client := http.Client{}
  body := bytes.NewBuffer( message )
  req, err := http.NewRequest( "POST", *url, body )
  if err != nil {
    panic( "error in request" )
  }
  req.Header.Set( "Content-Type", "application/json" )
  resp, err2 := client.Do( req )
  if err2 != nil {
    panic( "error in request 2" )
  }
  return
}
