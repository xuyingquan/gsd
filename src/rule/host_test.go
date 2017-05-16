package rule
import (
    "testing"
    "fmt"
    "encoding/json"
)

func TestLoadHosts(t *testing.T){
    LoadPools("/Users/chenggong/Documents/shataspace/nsd/etc/pools")
    LoadHosts("/Users/chenggong/Documents/shataspace/nsd/etc/hosts")
    fmt.Println(Hosts)
}

func TestJson(t *testing.T){
    data := `{"max":4, "ttl":300, "ipKey": "pub", "policy": "random", "record": "A", "target": [{"pool": "pool1", "weight": 1}]}`
    host := &Host{}
    err := json.Unmarshal([]byte(data), &host)
    fmt.Println(host, err)
}