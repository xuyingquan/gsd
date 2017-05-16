package main

import (
    "dnss"
    "runtime"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "github.com/mgutz/logxi/v1"
    "hserver"
    "rule"
)

const (
    Version = "5.2.1"
)

func main(){
    runtime.GOMAXPROCS(runtime.NumCPU())
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "usage: gsd.bin -l LISTEN_ADDR -c CONF_DIR\n")
        fmt.Fprintf(os.Stderr, "command line switches:\n")
        flag.PrintDefaults()
    }
    listenAddr := flag.String("l", ":53", "dns service listen ip:port")
    httpListenAddr := flag.String("hl", "", "http service listen ip:port")
    confDir := flag.String("c", "etc", "config dir")
    version := flag.Bool("v", false, "show version and exit")
    help := flag.Bool("h", false, "show help and exit")
    flag.Parse()

    if *version {
        fmt.Printf("Global Service Dispatcher v%s build by %s\n", Version, runtime.Version())
        os.Exit(0)
    }

    if *help {
        flag.Usage()
        os.Exit(0)
    }

    go func(){
        sc := make(chan os.Signal)
        for {
            signal.Notify(sc)
            s := <-sc
            if s == syscall.SIGHUP {
                log.Debug("receive HUP signal, reloading...")
                rule.CleanCache()
                dnss.Reload()
            }else if s== syscall.SIGINT {
                log.Info("receive INT signal, exit.")
                os.Exit(0)
            }else if s == syscall.SIGTERM {
                log.Info("receive TERM signal, exit.")
                os.Exit(0)
            }else{
                log.Info("receive unsupported signal, nothing to do", "signal", fmt.Sprintf("%v", s))
            }
        }

    }()
    if httpListenAddr != nil && *httpListenAddr != ""{
        go hserver.ListenAndServe(*httpListenAddr)
    }
    dnss.ListenAndServe(*listenAddr, *confDir)
}