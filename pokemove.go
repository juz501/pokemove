package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "io/ioutil"
  "strings"
  "strconv"
  "os"

  "github.com/urfave/negroni"
  "github.com/juz501/go_logger_middleware"
)

func main() {
  logfilename := "log/pokemove.log"
  errorLog, err := os.OpenFile( logfilename , os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666 )
  if err != nil {
    log.Fatal( "error writing to log: " + logfilename )
  }
  defer errorLog.Close()

  n := negroni.New()
  l := go_logger_middleware.NewLoggerWithStream( errorLog )
  m := http.NewServeMux()

  handleRoutes( m, l )

  n.Use( l )

  n.UseHandler( m )

  log.Fatal( http.ListenAndServe( ":18885", n ) )
}

func handleRoutes( m *http.ServeMux, l *go_logger_middleware.Logger ) {

  m.HandleFunc( "/", func( w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
      text := r.URL.Query().Get( "text" )
      url := ""
      response_url := r.URL.Query().Get( "response_url" )
      if response_url != "" {
        url = response_url
      } else {
        w.WriteHeader( http.StatusOK )
        return
      }
      if text == "" {
        fmt.Fprintf( w, "%s", about() )
      } else {
        go sendMoveInfo( url, text, w, l )
        fmt.Fprintf( w, "loading move data" )
      }
    default:
      w.WriteHeader( http.StatusOK )
    }
  })

  m.HandleFunc( "/favicon.ico", func( w http.ResponseWriter, r *http.Request) {
      fmt.Fprintf( w, "%s", about() )
  })
}

func about() ( s string ) {
  s = fmt.Sprint( "Just type your move after the /pokemove command" )
  return
}

type Message struct {
  Text string `json:"text"`
  Markdown bool `json:"mrkdwn"`
  ReplaceOriginal bool `json:"replace_original"`
}

func sendMoveInfo( url string, move_name string, w http.ResponseWriter, l *go_logger_middleware.Logger ) {
  s := getMoveResult( move_name )
  l.Println( move_name )

  m := Message{ s, true, true }
  message, err := json.Marshal( m )
  if err != nil {
    panic( "json marshall failed" )
  }
  _ = sendToSlack( url, message )
}

func sendToSlack( url string, message []byte) (resp *http.Response) {
  client := http.Client{}
  body := bytes.NewBuffer( message )
  req, err := http.NewRequest( "POST", url, body )
  if err != nil {
    return
  }
  req.Header.Set( "Content-Type", "application/json" )
  resp, _ = client.Do( req )
  defer resp.Body.Close()
  return
}

type Data struct {
  Accuracy int64 `json:"accuracy"`
  EffectEntries []Effect `json:"effect_entries"`
  EffectChance int64 `json:"effect_chance"`
  Name string `json:"name"`
  PP int64 `json:"pp"`
  Power int64 `json:"power"`
  Target Target `json:"target"`
  Type Type `json:"type"`
}

type Effect struct {
  Language Language `json:"language"`
  Effect string `json:"effect"`
  ShortEffect string `json:"short_effect"`
}

type Language struct {
  Name string `json:"name"`
}

type Target struct {
  Name string `json:"name"`
}

type Type struct {
  Name string `json:"name"`
}

func getMoveResult( move_name string ) ( move_desc string ) {
  res, err := http.Get( "http://pokeapi.co/api/v2/move/" + move_name )
  if err != nil {
    move_desc = ""
    return
  }
  move_data, err := ioutil.ReadAll( res.Body )
  var m Data
  _ = json.Unmarshal( move_data, &m )
  if err != nil {
    move_desc = ""
    return
  }

  if m.Name != "" {
    move_desc = "*" + strings.Replace( strings.ToUpper( m.Name ), "-", " ", -1 ) + ":* \n"
    move_desc += "• Accuracy : " + strconv.FormatInt( m.Accuracy, 10 ) + "\n"
    move_desc += "• PP : " + strconv.FormatInt( m.PP, 10 ) + "\n"
    move_desc += "• Power : " + strconv.FormatInt( m.Power, 10 ) + "\n"
    move_desc += "• Target : " + m.Target.Name + "\n"
    move_desc += "• Type : " + m.Type.Name + "\n"
    effect_chance := strconv.FormatInt( m.EffectChance, 10 )
    move_desc += "• Effect Chance : " + effect_chance + "\n"

    for _, entry := range m.EffectEntries {
      move_desc += "• Language : " + entry.Language.Name + "\n"
      move_desc += "• Effect short : " + strings.Replace( entry.ShortEffect, "$effect_chance", effect_chance, -1 ) + "\n"
      move_desc += "• Effect : \n>>>" + strings.Replace( entry.Effect, "$effect_chance", effect_chance, -1 ) + "\n"
    }
  }

  return
}
