package rule

import (
    "testing"
    "regexp"
    "strings"
    "fmt"
)

func TestRegex(t *testing.T){
    oriurl := "http://172.30.37.53:8060/sdn/other/ask/next/push?ip=$remoteIp&instance=$2&app=zqlive.$1"
    reg, _ := regexp.Compile("(.+)\\.(.+)\\.zqlive.push")
    url := strings.Replace(oriurl, "$remoteIp", "1.1.1.1", -1)
    url = reg.ReplaceAllString("htlive.com.998833.zqlive.push" , url)
    fmt.Println(url)
}