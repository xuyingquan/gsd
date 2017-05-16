package hserver

import (
    "net/http"
    "fmt"
    "rule"
    "github.com/mgutz/logxi/v1"
    "net"
)

func GsdServer(w http.ResponseWriter, req *http.Request) {
    params := req.URL.Query()
    remoteIp := params.Get("rip")
    domain := params.Get("domain")
    ip := net.ParseIP(remoteIp)
    if ip != nil && domain != "" {
        records, _, labels := rule.FindRecords(domain + ".", ip, true)
        log.Info("request by http", "domain", domain, "records", records, "labels", labels)
        if records != nil {
            result := ""
            for _, record := range *records {
                switch record.RecordType {
                    case rule.TypeA:
                    ip := record.Ip
                    result += ip.String() + "\n"
                    case rule.TypeCNAME:
                    result += record.CName
                }
            }
            fmt.Fprintf(w, "%s", result)
        }else{
            http.Error(w, domain + " Not Found", 404)
        }
    }else{
        http.Error(w, "Wrong Parameters\nremote ip: " + remoteIp + "\ndomain:" + domain + "\n", 403)
    }
}

func ListenAndServe(listenAddr string) {
    defer func(){
        if x := recover(); x != nil{
            fmt.Println("panic:", x)
        }
    }()
    log.Info("http service started", "listen", listenAddr)
    http.HandleFunc("/", GsdServer)
    err := http.ListenAndServe(listenAddr, nil)
    if err != nil {
        log.Fatal("start http service failed", "err", fmt.Sprintf("%v", err))
    }
}