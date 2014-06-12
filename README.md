volleyLT - a distributed load tester
==
### Still early in development, please use a tool like* JMeter or Tsung if you want to get real load testing work done for now

This is(will be) a distributed load testing tool.

The plan is that the client will talk to etcd to get a list of available load testing agents and then send commands for the agents to execute.
The agents will return the results which the client will summarize.


Running volleyLT with a single agent
--
1. go get github.com/billhathaway/volleyLT/volleyAgent
2. go get github.com/billhathaway/volleyLT/volleyCli
3. volleyAgent & # start agent on default port
4. volleyCli     # run a mini load test
5. volleyCli -h  # to see options

Running volleyLT with multiple agents (but not etcd)
--
1. volleyAgent -port 9875 &
2. volleyAgent -port 9876 &
3. volleyCli -agents localhost:9875,localhost:9876

Running volleyLT with multiple agents registered in etcd (assumes local etcd)
--
1. volleyAgent -port 9875 -etcd &
2. volleyAgent -port 9876 -etcd &
3. volleyAgent -port 9877 -etcd &
4. volleyCli -agentCount 2 -etcd & # picks 2 out of the 3 agents registered with etcd





