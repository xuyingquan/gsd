package rule

import (
    "testing"
    "fmt"
    "time"
    "net"
    "net/http"
    "io"
)

func TestFindRecordsAll(t *testing.T){
    LoadConf("/Users/chenggong/Documents/shataspace/nsd/etc")
    start := time.Now()
    counter := map[string]int{}
    for i := 0 ; i < 10000 ;i ++ {
        records, _, _ := FindRecords("all.shata.com.", net.ParseIP("127.0.0.1"), true)
        counter[fmt.Sprintf("%v", (*records)[0].Ip)] += 1
    }
    end := time.Now()
    fmt.Printf("hash count %v, time %dms\n", counter, (end.UnixNano() - start.UnixNano())/1000/1000)
}

func TestFindRecordsRandom(t *testing.T){
    LoadConf("/Users/chenggong/Documents/shataspace/nsd/etc")
    start := time.Now()
    counter := map[string]int{}
    for i := 0 ; i < 10000 ;i ++ {
        records, _, _ := FindRecords("random.shata.com.", net.ParseIP("127.0.0.1"), true)
        counter[fmt.Sprintf("%v", (*records)[0].Ip)] += 1
    }
    end := time.Now()
    fmt.Printf("hash count %v, time %dms\n", counter, (end.UnixNano() - start.UnixNano())/1000/1000)
}

func mockSdnAPi(w http.ResponseWriter, req *http.Request){
    fmt.Println(req.FormValue("app"), req.FormValue("instance"), req.FormValue("ip"))
    io.WriteString(w, "10.2.3.4")
}

func TestFindRecordsBySdnApi(t *testing.T){
    http.HandleFunc("/push", mockSdnAPi)
    go http.ListenAndServe(":10001", nil)
    time.Sleep(time.Second * 1)
    LoadConf("/Users/chenggong/Documents/shataspace/nsd/etc")
    records, _, _ := FindRecords("test.sdn.", net.ParseIP("192.168.1.1"), true)
    fmt.Println(records)
}