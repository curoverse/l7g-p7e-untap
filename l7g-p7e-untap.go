package main

import "os"
import "fmt"
import "log"
import "io/ioutil"

import "database/sql"
import _ "github.com/mattn/go-sqlite3"

import "github.com/abeconnelly/sloppyjson"

import "reflect"
import "time"

type LPUD struct {
  DB *sql.DB
  HTMLDir string
  JSDir string
  Port int
}

func (lpud *LPUD) Init(sql_fn string) error {
  var err error
  lpud.DB, err = sql.Open("sqlite3", "file:" + sql_fn + "?parseTime=true")
  if err !=nil { return err }
  return nil
}

func (lpud *LPUD) SQLExec(req string) ([][]string, error ) {
  local_debug := false

  rows,err := lpud.DB.Query(req)
  if err!=nil { return nil, err }
  cols,e := rows.Columns() ; _ = cols
  if e!=nil { return nil, e }

  columns := make([]interface{}, len(cols))
  columnPointers := make([]interface{}, len(cols))
  for i:=0; i<len(cols); i++ {
    columnPointers[i] = &columns[i]
  }

  res_str_array := [][]string{}

  ti := time.Time{}
  bt := []byte{}

  res_str_array = append(res_str_array, cols)

  for rows.Next() {
    if err := rows.Scan(columnPointers...); err != nil {
      return nil, err
    }

    strstr := []string{}

    for _,raw := range columns {
      if reflect.TypeOf(raw) == reflect.TypeOf(ti) {
        t := raw.(time.Time)
        s := t.String()
        strstr = append(strstr, s)
      } else if reflect.TypeOf(raw) == reflect.TypeOf(bt) {
        var s = fmt.Sprintf("%s", raw.([]byte))
        strstr = append(strstr, s)
      } else {
        var s = fmt.Sprintf("%v", raw)
        strstr = append(strstr, s)
      }
    }

    res_str_array = append(res_str_array, strstr)
  }


  //DEBUG
  if local_debug { fmt.Printf(">>>>\n%v\n", res_str_array) }

  return res_str_array, nil
}

func (lpud *LPUD) SQLExecS(req string) ([][]string, error ) {
  local_debug := true


  rows,err := lpud.DB.Query(req)
  if err!=nil { return nil, err }
  cols,e := rows.Columns() ; _ = cols
  if e!=nil { return nil, e }

  rawResult := make([][]byte, len(cols))

  res_str_array := [][]string{}

  // add column names to first row
  //
  res_str_array = append(res_str_array, cols)


  dest := make([]interface{}, len(cols))
  for i,_ := range rawResult {
    dest[i] = &rawResult[i]
  }

  for rows.Next() {
    err := rows.Scan(dest...)
    if err!=nil { return nil,err }

    result := make([]string, len(cols))

    for i,raw := range rawResult {
      raw_type := reflect.TypeOf(raw)
      if raw==nil {
        result[i] = "\n"
      } else if el := raw_type.Elem() ; (el.PkgPath() == "time" || el.Name() == "Time" ) {
        result[i] = fmt.Sprintf("%v", raw)
        //ti := raw.(time.Time)
        //result[i] = ti.String()
      } else {
        result[i] = string(raw)

        //DEBUG
        if local_debug {
          fmt.Printf("raw>>>>\n%v\n", string(raw))
        }

      }
    }

    res_str_array = append(res_str_array, result)

  }

  //DEBUG
  if local_debug {
    fmt.Printf(">>>>\n%v\n", res_str_array)
  }

  return res_str_array, nil
}

func main() {
  local_debug := true

  lpud := LPUD{}

  cfg_fn := "./l7g-p7e-config.json"
  if len(os.Args)>1 { cfg_fn = os.Args[1] }

  //config_s,e := ioutil.ReadFile("./l7g-p7e-config.json")
  config_s,e := ioutil.ReadFile(cfg_fn)
  if e!=nil { log.Fatal(e) }
  config_json,e := sloppyjson.Loads(string(config_s))
  if e!=nil { log.Fatal(e) }

  err := lpud.Init(config_json.O["database"].S)
  if err!=nil { log.Fatal(err) }

  lpud.Port = int(config_json.O["port"].P)
  lpud.HTMLDir = config_json.O["html-dir"].S
  lpud.JSDir = config_json.O["js-dir"].S

  if local_debug {
    fmt.Printf(">> starting\n")
  }

  err = lpud.StartSrv()
  if err!=nil { panic(err) }

}
