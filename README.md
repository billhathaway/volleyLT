volleyLT - a distributed load tester
==
### Still early in development, please use a tool like* JMeter or Tsung if you want to get real load testing work done for now

This is(will be) a distributed load testing tool.  

The plan is that the client will talk to etcd to get a list of available load testing agents and then send commands for the agents to execute.
The agents will return the results which the client will summarize.

Currently the cli only talks to a single agent 

Running volleyLT
--
1. go get github.com/billhathaway/volleyLT/volleyAgent 
2. go get github.com/billhathaway/volleyLT/volleyCli
3. volleyAgent &
4. volleyCli
5. volleyCli -h  # to see options

