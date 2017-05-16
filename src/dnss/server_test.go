package dnss
import (
    "testing"
    "time"
)

func TestPrintAccessLog(t *testing.T){
    aLog := accesslog{
        ReqTime: time.Now(),
        RespTime: time.Now(),
        IsEDns: true,
        IsFound: true,
        Domain: "test.shatacloud.com",
        RemoteIp: "192.168.1.1",
        RecordType: "A",
        IpList: []string{"202.96.209.5", "202.96.209.6"},
        Labels: map[string]string{"channel": "testchn"},
    }
    printAccessLog(aLog)
}

func BenchmarkPrintAccessLog(b *testing.B){
    aLog := accesslog{
        ReqTime: time.Now(),
        RespTime: time.Now(),
        IsEDns: true,
        IsFound: true,
        Domain: "test.shatacloud.com",
        RemoteIp: "192.168.1.1",
        RecordType: "A",
        IpList: []string{"202.96.209.5", "202.96.209.6"},
        Labels: map[string]string{"channel": "testchn"},
    }
    printAccessLog(aLog)
}
