package main

import (
    "github.com/miekg/dns"
    "flag"
    "os"
    "sync"
    "time"
    "fmt"
    "bufio"
    "math/rand"
)

var mutex sync.Mutex
var wg sync.WaitGroup

func main(){
    queryTimes := flag.Int("q", 10000, "query times")
    concurrent := flag.Int("c", 100, "concurrent queries")
    addr := flag.String("a", "127.0.0.1:53", "dns server addr")
    domainName := flag.String("t", "", "domain name for test")
    domainFile := flag.String("d", "", "domain file for test")
    remoteIp := flag.String("r", "", "remote ip for request")
    flag.Parse()
    if *domainName == "" && *domainFile == ""{
        flag.Usage()
        os.Exit(1)
    }

    domains := []string{}
    if *domainName != "" {
        domains = append(domains, *domainName)
    }

    if *domainFile != "" {
        file, err := os.Open(*domainFile)
        defer file.Close()
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        r := bufio.NewReader(file)
        for {
            line, _, err := r.ReadLine()
            if err != nil {
                break
            }
            domains = append(domains, string(line))
        }
    }
    domainCount := len(domains)

    fmt.Println("==dns performance test start===")
    fmt.Println("query times:", *queryTimes)
    fmt.Println("domain count:", domainCount)
    fmt.Println("concurrent:", *concurrent)
    if *remoteIp != ""{
        fmt.Println("remoteIp:", *remoteIp)
    }
    succ := 0
    fail := 0
    minRespTime := time.Second*100
    maxRespTime := time.Second*0
    sumRespTime := time.Second*0
    start := time.Now().UnixNano()
    for i:=0;i<*concurrent;i++ {
        wg.Add(1)
        go func(){
            r := rand.New(rand.NewSource(12345))
            for j:=0;j<(*queryTimes)/(*concurrent);j++ {
                ri := r.Intn(domainCount)
                m := &dns.Msg{}
                m.SetQuestion(domains[ri]+".", dns.TypeA)
                if *remoteIp != "" {
                    m.SetQuestion(*remoteIp, dns.EDNS0SUBNET)
                }
                c := &dns.Client{ReadTimeout: time.Second*2}
                _, rtt, err := c.Exchange(m, *addr)
                if err == nil {
                    mutex.Lock()
                    succ += 1
                    if rtt < minRespTime {
                        minRespTime = rtt
                    }
                    if rtt > maxRespTime {
                        maxRespTime = rtt
                    }
                    sumRespTime += rtt
                    mutex.Unlock()
                }else {
                    mutex.Lock()
                    fail += 1
                    fmt.Println(err)
                    mutex.Unlock()
                }
            }
            wg.Done()
        }()
    }
    wg.Wait()
    end := time.Now().UnixNano()
    runningTime := end - start
    fmt.Println("==dns performance test finished==")
    fmt.Println("succ:", succ)
    fmt.Println("fail:", fail)
    fmt.Println("min response time:", minRespTime)
    fmt.Println("max response time:", maxRespTime)
    fmt.Printf("mean response time: %f ms\n", float64(sumRespTime)/float64(succ)/1000/1000)
    fmt.Printf("running time: %d ms\n", int64(runningTime)/1000/1000)
    fmt.Printf("performace: %d req/s\n", int64(succ)*1000*1000*1000/int64(runningTime))
}
