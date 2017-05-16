package rule

import (
    "testing"
    "fmt"
)

func TestLoadPools(t *testing.T) {
    err := LoadPools("/Users/chenggong/Documents/shataspace/nsd/etc/pools")
    fmt.Println(err)
    fmt.Println(Pools)
}

func TestFindEntries(t *testing.T) {
    LoadPools("/Users/chenggong/Documents/shataspace/nsd/etc/pools")
    entries := FindEntriesByName("dc1", true, false)
    counter := map[string]int{}
    for i := 0; i < len(entries); i++ {
        counter[entries[i].Name] = 0
    }
    for i := 0; i < 7000; i++ {
        firstEntry := FindEntriesByName("dc1", true, false)[0]
        counter[firstEntry.Name] += 1
    }
    fmt.Println("should be 2:3:1:1, result:", counter)
}