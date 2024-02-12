//This file is only used to speedup the docker build
//We use this to download and compile the go module dependencies before adding our own source code.
//See Dockerfile for more details

package main

import (
	_ "flag"
	_ "net/http"
	_ "os"
	_ "os/signal"
	_ "sync"
	_ "syscall"
	_ "time"

	_ "github.com/golang/glog"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "k8s.io/client-go/informers"
	_ "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/rest"
	_ "k8s.io/client-go/tools/clientcmd"
)

func main() {

}
