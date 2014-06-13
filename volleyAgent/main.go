// volleyLT project main.go
package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	etcdTTLsecs = 10
)

type (
	Controller struct {
		log *log.Logger
	}

	VolleyRequest struct {
		Concurrency       int
		MaxErrors         int //not used yet
		Url               string
		Requests          int
		DisableKeepAlives bool
	}

	VolleyResponse struct {
		StartTime  time.Time
		Duration   time.Duration
		Error      error
		ErrorCount int
		Responses  []Response
	}

	Response struct {
		Url        string
		StartTime  time.Time
		Duration   time.Duration
		Bytes      int
		StatusCode int
		Error      string
	}
)

func NewController(port string) *Controller {
	return &Controller{log.New(os.Stdout, fmt.Sprintf("volleyAgent:port=%s ", port), log.LstdFlags)}
}

func (c *Controller) worker(id int, wg *sync.WaitGroup, tr *http.Transport, urlChan chan string, statusChan chan Response) {
	var vResponse Response
	var content []byte
	client := &http.Client{Transport: tr}
	c.log.Printf("event=workerStart id=%d\n", id)
	for url := range urlChan {
		vResponse.Url = url
		vResponse.StartTime = time.Now()
		httpResponse, err := client.Get(url)
		if err != nil {
			vResponse.Error = err.Error()
			vResponse.Duration = time.Since(vResponse.StartTime)
			statusChan <- vResponse
			c.log.Printf("Received error response: %s\n", err.Error())
			continue
		}
		content, _ = ioutil.ReadAll(httpResponse.Body)
		httpResponse.Body.Close()
		vResponse.Bytes = len(content)
		vResponse.Duration = time.Since(vResponse.StartTime)
		vResponse.StatusCode = httpResponse.StatusCode
		statusChan <- vResponse
	}
	c.log.Printf("event=workerEnd id=%d\n", id)
	wg.Done()
}

func (c *Controller) Execute(vRequest VolleyRequest, vResponse *VolleyResponse) error {
	tr := &http.Transport{DisableKeepAlives: vRequest.DisableKeepAlives}
	vResponse.Responses = make([]Response, vRequest.Requests)
	c.log.Printf("event=start url=%s requests=%d concurrency=%d\n", vRequest.Url, vRequest.Requests, vRequest.Concurrency)
	startTime := time.Now()
	vResponse.StartTime = startTime
	urlChan := make(chan string)
	statusChan := make(chan Response)
	wg := &sync.WaitGroup{}
	wg.Add(vRequest.Concurrency)
	for i := 0; i < vRequest.Concurrency; i++ {
		go c.worker(i, wg, tr, urlChan, statusChan)
	}

	go func() {
		for i := 0; i < vRequest.Requests; i++ {
			vResponse.Responses[i] = <-statusChan
		}
	}()
	for i := 0; i < vRequest.Requests; i++ {
		urlChan <- vRequest.Url
	}
	close(urlChan)
	wg.Wait()
	vResponse.Duration = time.Since(startTime)
	c.log.Printf("event=finish duration=%s\n", time.Since(startTime))
	return nil
}

func updateEtcd(etcdPath string, etcdServers string, port string) {
	var ipAddr string
	addresses, err := net.InterfaceAddrs()
	// on the client side, net.Dial wants IPv6 addresses to be enclosed in []
	for _, addr := range addresses {
		if !strings.HasPrefix(addr.String(), "127.") && !strings.Contains(addr.String(), "::") {
			ipAddr = strings.Split(addr.String(), "/")[0]
			break
		}
	}

	// localhost as last resort
	if ipAddr == "" {
		ipAddr = "127.0.0.1"
	}

	if err != nil {
		log.Fatal("Error cannot get hostname - %s\n", err.Error())
	}

	client := etcd.NewClient(strings.Split(etcdServers, ","))
	_, err = client.Set(fmt.Sprintf("%s/%s:%s", etcdPath, ipAddr, port), port, etcdTTLsecs+1)
	if err != nil {
		log.Fatalf("Error updating etcd %s", err.Error())
	}
	ticker := time.NewTicker(etcdTTLsecs * time.Second)
	for {
		<-ticker.C
		_, err = client.Set(fmt.Sprintf("%s/%s:%s", etcdPath, ipAddr, port), port, etcdTTLsecs+1)
		if err != nil {
			log.Fatalf("Error updating etcd %s", err.Error())
		}
	}
}

func main() {
	port := flag.String("port", "9876", "rpc listen port")
	etcdServers := flag.String("etcdServers", "http://localhost:4001", "csv of etcd URLs")
	etcdPath := flag.String("etcdPath", "/volleyAgent", "etcd path for to register on")
	useEtcd := flag.Bool("etcd", false, "register with etcd")
	flag.Parse()
	if *useEtcd {
		go updateEtcd(*etcdPath, *etcdServers, *port)
	}

	controller := NewController(*port)
	rpc.Register(controller)

	l, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	for {
		conn, _ := l.Accept()
		go rpc.ServeConn(conn)
	}
}
