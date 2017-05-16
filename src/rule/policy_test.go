package rule

import (
    "testing"
    "fmt"
    "math/rand"
    "time"
    "encoding/json"
)

func TestLoadPolicies(t *testing.T){
    err := LoadPolicies("/Users/chenggong/Documents/shataspace/nsd/etc/policies")
    fmt.Println(err)
    fmt.Println(Policies)
}

func TestFindEntriesByRegion(t *testing.T){
    Policies = map[string]Policy{
        "test": Policy {
            DispatchMap: map[string][]Target{
                "中国/河北省@中国联通": []Target{
                        Target{Pool: "河北联通", Weight:1},
                        Target{Pool: "河北联通2", Weight:3},
                        Target{Pool: "河北联通3", Weight:1},
                },
            },
        },
    }
    Pools = map[string]Pool{
            "河北联通": Pool{
                Entries: []Entry{
                    Entry{
                        Name: "e1",
                    },
                    Entry{
                        Name: "e2",
                    },
                },
            },
            "河北联通2": Pool{
                Entries: []Entry{
                    Entry{
                        Name: "e3",
                        Weight: 1,
                    },
                    Entry{
                        Name: "e4",
                        Weight: 2,
                    },
                },
            },
            "河北联通3": Pool{
                Entries: []Entry{
                    Entry{
                        Name: "e5",
                    },
                    Entry{
                        Name: "e6",
                    },
                },
            },
        }
    rand.Seed(time.Now().UnixNano())
    counter := map[string]int{}
    for i := 0;i < 5000; i++  {
        entries, _ := findEntriesByRegion("test", Region{Region:"中国/河北省/石家庄市/其他县", Isp:"中国联通"}, false)
        if len(entries) != 6 {
            t.Fail()
        }
        if i == 0 {
            fmt.Println(entries)
        }
        entry := entries[0].Name
        _, ok := counter[entry]
        if !ok{
            counter[entry] = 1
        }else{
            counter[entry] += 1
        }
    }
    fmt.Println("should be 1:1:2:4:1:1 ")
    for k, v := range counter {
        fmt.Printf("%s:%d\n", k, v)
    }
}

func TestFindEntriesByRegionFromRedis(t *testing.T){
    InitRedis("localhost:6379", "")
    entries, _ := findEntriesByRegion("test", Region{Region:"中国/河北省/石家庄市/其他县", Isp:"中国联通"}, true)
    fmt.Println(entries)
}

func TestPolicyJson(t *testing.T){
    data := `{"dispatch": {"default@default": [{"pool": "pool1", "weight": 1, "priority": 2}, {"pool":"pool2", "weight": 1, "priority": 1}]}}`
    policy := Policy{}
    err := json.Unmarshal([]byte(data), &policy)
    fmt.Println(policy, err)
}