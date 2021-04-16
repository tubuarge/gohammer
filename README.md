# GoHammer
GoHammer is a test tool designed to get performance metrics (TPS) of the nodes and operationg system by deploying a smart contract and calling smart contract's methods. <br /> GoHammer provides configurable test profiles created by the user and easy execution of these test profiles. GoHammer inspired by [Chainhammer](https://github.com/drandreaskrueger/chainhammer) and [Quorum Profiling](https://github.com/ConsenSys/quorum-profiling) <br />

GoHammer deploys number of transactions on the given nodes according to configuration file (config.json) then tps-monitoring tool collects TPS and node metrics and visualizes these data on Grafana (if you want to see node metrics gohammer has to be run on that node, if you are testing a remote node you can't get OS related metrics about these node.)

## Requirements
* Docker CLI
* Go Version 1.15.6

## Installation
```bash
git clone https://github.com/tubuarge/gohammer
cd gohammer/gohammer
go build
```

## Usage
Before running `gohammer` make sure every node in the given configuration file is running.

Start TPS-Monitor tool.
```bash 
cd monitoring
docker-compose up -d
cd ../tps-monitor
go build 
./tps-monitor --httpendpoint http://localhost:22000 --consensus raft --influxdb --influxdb.token admin:admin
```
After starting tps-monitor you can access Grafana from your browser with this URL `http://localhost:3000` with `username: admin` and `password: admin`.

```bash
./gohammer --testprofilefile ../config.json --logdir /home/test/logs
```
When the test finished you can find your test's log under file named `<logdirPath>YYYY_MM_DD HH_MM_result.log`.

## List of Command-Line Arguments
| Command        | Description  |
| :-------------: |:-------------:|
| --testprofilefile <config.json> | path of a json file that contains your test config. |
| --logdir <dirPath> | directory path where you want to store your log files (default is the same directory with gohammer executable). |

## Test Profile Config File(config.json)
Config file is a json file that consists of test profiles, this is the section where you can construct different types of test profiles. <br />
```json
{
  "profiles": [
    {
      "name": "Test Profile1",
      "concurrent": false,
      "roundRobin": true,
      "callContractMethod": true,
      "nodes": [
        {
         ...
        }
      ]
    },
    {
      "name": "Test Profile2",
      ...
    },
    ...
  ]
}
```
### Profiles
`profiles` section contains information about how transactios will be run, which nodes will be involved.
| key   | Value | Type |
| :---: | :---: | :---: |
| name  | name of the profile | string |
| concurrent | transaction will be deployed concurrently (NOTE: not implemented yet.) | boolean |
| roundRobin | a transaction will be deployed on the given nodes one after the other | boolean |
| callContractMethod | instead of deploying smart contracts, nodes are going to call method of the smart contract | boolean |
| nodes | nodes where the test profile will be run (for more information check `nodes` section | json array |
<br />

`Important`: If you haven't set any deploy transaction configuration (like `roundRobin` or `concurrent`) your transaction will be deployed according to default configuration which is deploying number of transactions on a node then proceeding to other node.


### Nodes
`nodes` section contains information about nodes where transactions will be deployed.

| key | Value | type|
| :---: | :---: | :---: |
| name | name of the node | string |
| url | url and port of the node | string |
| cipher | cipher text of the node (can be found under qdata/dd{x}/keystore/key file) | string |
| deployCounts | how many transactions will be deployed on the given node | json array |
| deployInterval | how much time test will be stalled after deploying number of transactions ("10s", "1m" etc.) | string |



