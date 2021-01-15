#!/bin/bash

ABI_GEN_PATH=""
GO_ETHEREUM_PATH=""

CONTRACT_PATH=""
CONTRACT_FILE_ABI=""
CONTRACT_FILE_BIN=""
CONTRACT_FILE_PATH=""

CONTRACT_STORE_FILE_PATH=""

TPS_MONITOR_PATH=""
GO_HAMMER_PATH=""
DOCKER_COMPUSE_PATH=""

QUORUM_ENDPOINT=""
CONSENSUS=""



function usage () {
  echo ""
  echo "Usage:"
  echo "    $0 --testProfile '<jmeter test profile>' --consensus <raft|ibft> --endpoint <quorum RPC endpoint> --basedir <repo base dir>"
  echo ""
  echo "Where:"
  echo "    contract - path of the smart contract. default Store.sol"
  echo "    testProfile - name of jmeter test profile. eg: 4node/deploy-contract-public)"
  echo "    consensus - name of consensus - raft or ibft. eg: raft)"
  echo "    endpoint - quorum rpc endpoint. eg: http://localhost:22000)"
  echo "    basedir - base dir of repo. eg: /home/bob/quorum-profiling)"
  echo ""
  exit -1
}


while (( $# )); do
	case "$1" in
		--contract)
			CONTRACT_FILE_PATH=$2
			shift 2
			;;
		 --consensus)
			CONSENSUS=$2
			if [ $CONSENSUS != 'raft' ] && [ $CONSENSUS != 'ibft' ]
			then
				echo "consensus must be raft or ibft"
				usage
			fi
			shift 2
			;;
		--endpoint)
			QUORUM_ENDPOINT=$2
			shift 2
			;;
		--help)
			shift
			usage
			;;
		*)
			echo "Error: Unsportted command line parameter $1"
			usage
			;;
	esac
done

function install_solc () {
	sudo add-apt-repository ppa:ethereum/ethereum
	sudo apt-get update
	sudo apt-get install solc
}

function build_abigen () {
	# check if go-ethereum exists
	if [[ -f $GO_ETHEREUM_PATH ]]; then
		echo "debug: goethereum found $GO_ETHEREUM_PATH ."
	else
		echo "debug: not found go-ethereum"
		echo "getting go-ethereum"
		go get -d github.com/ethereum/go-ethereum
		GO_ETHEREUM_PATH=${GOPATH}/src/github.com/ethereum/go-ethereum
		ABI_GEN_PATH=${GOPATH}/src/github.com/ethereum/go-ethereum/cmd/abigen
	fi
}

function generate_abi_bin () {
	CONTRACT_PATH=$( cd "$( dirname "${CONTRACT_FILE_PATH}" )" >/dev/null 2>&1 && pwd )
	if [[ -f $CONTRACT_FILE_PATH ]]; then
		echo "contract path is $CONTRACT_FILE_PATH "
	else
		echo "contract not found."
		echo "using default contract."
		CONTRACT_PATH=$CONTRACT_PATH/contract
		CONTRACT_FILE_PATH=$CONTRACT_PATH/Store.sol
	fi

	solc --abi $CONTRACT_FILE_PATH | awk 'NR>3' > $CONTRACT_PATH/out.abi
	CONTRACT_FILE_ABI=$CONTRACT_PATH/out.abi

	solc --bin $CONTRACT_FILE_PATH | awk 'NR>3' > $CONTRACT_PATH/out.bin
	CONTRACT_FILE_BIN=$CONTRACT_PATH/out.bin
}

function generate_go_modules () {
	abigen --bin=$CONTRACT_FILE_BIN --abi=$CONTRACT_FILE_ABI --pkg=main --out=store.go

	CONTRACT_STORE_FILE_PATH="$( dirname $0 )"/store.go
}

function build_tps_monitor () {
	echo "Starting to build tps-monitor."
	cd tps-monitor ; make

	# check if tps-monitor binary exists
	if [[ ! -f tps-monitor ]]; then
		echo "Couldn't build tps-monitor."
		exit 1
	else
		echo "tps-monitor is installed."
	fi

	TPS_MONITOR_PATH="$( dirname $0 )"/tps-monitor
	cd ..
}

function build_gohammer () {
	go get -u -v ; go build

	# check if gohammer binary exists
	if [[ ! -f gohammer ]]; then
		echo "Couldn't build gohammer."
		exit 1
	else
		echo "gohammer is installed."
	fi
	GO_HAMMER_PATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"/gohammer
}

function install_telegraf () {
	if ! [ -x "$(command -v telegraf)" ]; then
		echo "debug: telegraf is not installed."
		echo "debug: installing telegraf."
		curl -sL https://repos.influxdata.com/influxdb.key | sudo apt-key add -
		source /etc/lsb-release
		echo "deb https://repos.influxdata.com/${DISTRIB_ID,,} ${DISTRIB_CODENAME} stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
		sudo apt-get update && sudo apt-get install telegraf
	else
		echo "debug: telegraf is installed."
	fi

	# copy telegraf.conf
	sudo cp "$( dirname $0 )"/monitoring/telegraf.conf /etc/telegraf/telegraf.conf
}

function start_telegraf () {
	# TODO check whether service is running or not.
	systemctl is-active --quiet telegraf
	if [[ "$(systemctl is-active telegraf)" = "active" ]]; then
		echo "Telegraf service is running."
	else
		echo "Starting telegraf service."
		sudo systemctl start telegraf
	fi
}

function start_docker_containers () {
	# check monitoring network is exists.
	if [[ ! "$(docker network ls | grep -w \"monitoring\")" ]]; then
		echo "Debug: monitoring docker network already exists."
	else
		docker network create monitoring
	fi

	# check volumes are exist or not.
	if [[ ! "$(docker volume ls | grep -w \"grafana-volume\")" ]]; then
		docker volume create grafana-volume
	else
		echo "Debug: grafana-volume already exists."
	fi

	if [[ ! "$(docker volume ls | grep -w \"influxdb-volume\")" ]]; then
		docker volume create influxdb-volume
	else
		echo "Debug: docker-volume already exists."
	fi

	echo "Starting Grafana, Influxdb on Docker"
	DOCKER_COMPUSE_PATH="$( dirname $0 )"/monitoring
	# TODO: before running docker-compose check if the containers already running or not.
	: 'if [[ "$(docker ps -a | grep -w \"grafana\")" ]]; then
		echo "Debug: Grafana is running."
		docker stop grafana
		docker rm grafana
	fi

	if [[ "$(docker ps -a | grep -w \"influxdb\")" ]]; then
		echo "Debug: Influxdb is running."
		docker stop influxdb
		docker rm influxdb
	fi
'

	cd "${DOCKER_COMPUSE_PATH}"
	docker-compose up -d
	cd ..
}

function start_tps_monitor () {
	echo "Starting tps-monitor"
	echo "{$CONSENSUS}"
	"${TPS_MONITOR_PATH}"/tps-monitor --httpendpoint $QUORUM_ENDPOINT --consensus=${CONSENSUS} --influxdb --influxdb.token admin:admin
}

# check GOPATH is set.
if [[ -z "${GOPATH}" ]]; then
	echo "debug: not found gopath"
	echo "Check your GOPATH variable"
	exit 1
else
	echo "debug: found gopath ${GOPATH} "
fi

# check if solc exe is exits.
if [[ -z "$(command -v solc)" ]]; then
	echo "debug: solc is not installed."
	echo "debug: installing solc."
	install_solc
fi

# check if abigen path exists
if [[ -z "$ABI_GEN_PATH" ]]; then
	echo "debug: abigen not found."
else
	ABI_GEN_PATH=$GOPATH/src/github.com/ethereum/go-ethereum/cmd/abigen
	GO_ETHEREUM_PATH=$GOPATH/src/github.com/ethereum/go-ethereum
fi

# check if abigen exe is exists
if ! [ -x "$(command -v abigen)" ]; then
	echo "debug: abigen is not installed."
	echo "debug: installing abigen."
	go build {ABI_GEN_PATH}/main.go
else
	echo "debug: abigen is installed."
fi

#generate_abi_bin
#generate_go_modules
#build_tps_monitor
#build_gohammer
start_docker_containers
#install_telegraf
#start_telegraf
#start_tps_monitor


