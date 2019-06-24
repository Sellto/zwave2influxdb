package main

import (
    //"encoding/xml"
    "fmt"
    "time"
    "net/http"
    "io/ioutil"
    "log"
    "github.com/clbanning/mxj"
    "github.com/influxdata/influxdb1-client/v2"
    "strconv"
)

//PRE : those values must be valid.
const (
    db_name = "ZWave"
    db_user = "zwave"
    db_pass = "zwave"
    db_addr = "http://influxdb.tick.svc.cluster.local:8086"
    poll_addr = "http://ozwcp.zwave.svc.cluster.local:8090/poll.xml"
)

//Catch Error function
func CatchError(e error) {
  if e != nil {
    fmt.Println("Error:", e.Error())
    return
  }
}

//Access to the poll.xml file from the REST API of ozwcp
func Get() []byte {
      result, err := http.Get(poll_addr)
      if err != nil {
          log.Fatal(err)
      }
      robots, err := ioutil.ReadAll(result.Body)
      defer result.Body.Close()
      return robots
  }

//Create influxdb Client
func influxDBClient(addr,user,pass string) client.Client {
      c, err := client.NewHTTPClient(client.HTTPConfig{
          Addr:     addr,
          Username: user,
          Password: pass,
      })
      CatchError(err)
      return c
  }

//Create entry in the node_data folder with a filter by node ID
func CreateDBEntry(eventTime time.Time, nodeid, field string, value interface{} ) client.BatchPoints{
  tags := map[string]string{
      "node": nodeid,
  }
  fields :=  make(map[string]interface{})
  fields[field] = value
  point, err := client.NewPoint(
      "node_data",
      tags,
      fields,
      eventTime,
  )
  CatchError(err)
  bp, err := client.NewBatchPoints(client.BatchPointsConfig{
      Database:  db_name,
      Precision: "s",
  })
  CatchError(err)
  bp.AddPoint(point)
  return bp
}



func loop(c client.Client) {
  //Get data from poll.xml
  xmldata:=Get()
  //Parse the data as map([]interface{})[]interface{} of xmldata
  m, err := mxj.NewMapXml(xmldata)
  CatchError(err)
  //Get the node path value
  node, err := m.ValuesForPath("poll.node")
  CatchError(err)
  //Check if node is present
  if len(node) != 0 {
    //Extract Data
    //prod,err := m.ValuesForKey("-product")
    id,err := m.ValuesForKey("-id")
    CatchError(err)
    //Get all available data from this node
    values,err := m.ValuesForPath("poll.node.value")
    CatchError(err)
    //Loop to find extract the desired value
    for _, value := range values {
      castvalue, _ := value.(map[string]interface{})
      if stringInSlice(castvalue["-label"].(string), []string{"Temperature","Battery Level","Luminance"}) {
        value_str2float, err := strconv.ParseFloat(castvalue["#text"].(string), 64)
        CatchError(err)
        c.Write(CreateDBEntry(time.Now(),id[0].(string),castvalue["-label"].(string),value_str2float))
        log.Println(castvalue["-label"].(string)+" value of the node "+id[0].(string)+" : "+castvalue["#text"].(string))
      }
    }
  }
  //Wait before a other loop turn
  time.Sleep(500*time.Millisecond)
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

//Superloop Program
func main() {
   for {
     loop(influxDBClient(db_addr,db_user,db_pass))
   }
}
