package main

import (
	"flag"
	"fmt"
	"github.com/billhathaway/volleyLT/common"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"net/rpc"
	"strings"
	"sync"
)

func main() {
	url := flag.String("url", "http://www.google.com/", "url to hit")
	requests := flag.Int("n", 10, "number of requests")
	concurrency := flag.Int("c", 1, "amount of concurrency")
	keepAlive := flag.Bool("k", false, "use HTTP keepalive")
	agentCount := flag.Int("agentCount", 1, "number of agents to use (only needs to be set when using etcd)")
	agents := flag.String("agents", "localhost:9876", "csv of server:port pairs")
	etcdServers := flag.String("etcdServers", "http://localhost:4001", "csv of etcd URLs")
	etcdPath := flag.String("etcdPath", "/volleyAgent", "etcd path for registered agents")
	useEtcd := flag.Bool("etcd", false, "use etcd for agents (otherwise specify with -agents)")
	full := flag.Bool("full", false, "show full results")
	sum := flag.Bool("sum", true, "show summary results")
	flag.Parse()
	var servers []string

	if *useEtcd {
		client := etcd.NewClient(strings.Split(*etcdServers, ","))
		response, err := client.Get(*etcdPath, false, true)
		if err != nil {
			log.Fatal("Error talking to etcd servers=%s path=%d err=%s\n", etcdServers, etcdPath, err.Error())
		}
		for _, node := range response.Node.Nodes {
			servers = append(servers, strings.TrimPrefix(node.Key, *etcdPath+"/"))
		}
	} else {
		servers = strings.Split(*agents, ",")
		if *agentCount == 1 && len(servers) > 1 {
			*agentCount = len(servers)
		}
	}

	if *agentCount > len(servers) {
		log.Fatalf("Error wanted %d agents but only able to find %d (%v)\n", *agentCount, len(servers), servers)
	}

	loadTestRequest := volley.Request{
		Url:               *url,
		Requests:          *requests,
		DisableKeepAlives: !*keepAlive,
		Concurrency:       *concurrency,
	}

	responses := make([]*volley.SessionResponse, *agentCount)
	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(*agentCount)
	for i := 0; i < *agentCount; i++ {
		go func(i int) {
			loadTestResponse := &volley.SessionResponse{}
			client, err := rpc.DialHTTP("tcp", servers[i])
			if err != nil {
				log.Fatal("dialing:", err)
			}
			err = client.Call("Controller.Execute", loadTestRequest, loadTestResponse)
			if err != nil {
				log.Fatal("calling %s returned :", servers[i], err)
			}
			mutex.Lock()
			responses[i] = loadTestResponse
			mutex.Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()

	if *full {
		raw := &rawReporter{}
		fmt.Println(raw.report(responses))
	}
	if *sum {
		stats := &statReporter{}
		fmt.Println(stats.report(responses))
	}

}
