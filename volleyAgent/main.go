// volleyLT project main.go
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
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
		Error      error
	}
)

func NewController() *Controller {
	return &Controller{log.New(os.Stdout, "volleyAgent ", log.LstdFlags)}
}

func (c *Controller) worker(id int, wg *sync.WaitGroup, tr *http.Transport, urlChan chan string, statusChan chan Response) {
	var vResponse Response
	var httpResponse *http.Response
	var err error
	var content []byte
	client := &http.Client{Transport: tr}
	c.log.Printf("event=workerStart id=%d\n", id)
	for url := range urlChan {
		vResponse.Url = url
		vResponse.StartTime = time.Now()
		httpResponse, err = client.Get(url)
		if err != nil {
			vResponse.Error = err
			vResponse.Duration = time.Since(vResponse.StartTime)
			statusChan <- vResponse
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

func main() {
	port := flag.String("port", "9876", "rpc listen port")
	flag.Parse()
	controller := NewController()
	rpc.Register(controller)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)
}
