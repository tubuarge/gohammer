#!/bin/bash

ABIGENPATH=""
GOETHEREUMPATH=""

CONTRACTPATH=""
CONTRACTFILEABI=""
CONTRACTFILEBIN=""
CONTRACTFILEPATH=""

CONTRACTSTOREFILEPATH=""

TPSMONITORPATH=""
GOHAMMERPATH=""

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
			CONTRACTFILEPATH=$2
			shift2
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
	if [[ -f $GOETHEREUMPATH ]]; then
		echo "debug: goethereum found $GOETHEREUMPATH ."
	else
		echo "debug: not found go-ethereum"
		echo "getting go-ethereum"
		go get -d github.com/ethereum/go-ethereum
		GOETHEREUMPATH=${GOPATH}/src/github.com/ethereum/go-ethereum
		ABIGENPATH=${GOPATH}/src/github.com/ethereum/go-ethereum/cmd/abigen
	fi
}

function generate_abi_bin () {
	CONTRACTPATH=$( cd "$( dirname "${CONTRACTFILEPATH}" )" >/dev/null 2>&1 && pwd )
	if [[ -f $CONTRACTFILEPATH ]]; then
		echo "contract path is $CONTRACTFILEPATH "
	else
		echo "contract not found."
		echo "using default contract."
		CONTRACTPATH=$CONTRACTPATH/contract
		CONTRACTFILEPATH=$CONTRACTPATH/Store.sol
	fi

	solc --abi $CONTRACTFILEPATH | awk 'NR>3' > $CONTRACTPATH/out.abi
	CONTRACTFILEABI=$CONTRACTPATH/out.abi

	solc --bin $CONTRACTFILEPATH | awk 'NR>3' > $CONTRACTPATH/out.bin
	CONTRACTFILEBIN=$CONTRACTPATH/out.bin
}

function generate_go_modules () {
	abigen --bin=$CONTRACTFILEBIN --abi=$CONTRACTFILEABI --pkg=main --out=store.go

	CONTRACTSTOREFILEPATH="$( dirname $0 )"/store.go
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

	TPSMONITORPATH="$( dirname $0)"/tps-monitor/tps-monitor
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
	GOHAMMERPATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"/gohammer
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
if [[ -z "$ABIGENPATH" ]]; then
	echo "debug: abigen not found."
else
	ABIGENPATH=$GOPATH/src/github.com/ethereum/go-ethereum/cmd/abigen
	GOETHEREUMPATH=$GOPATH/src/github.com/ethereum/go-ethereum
fi

# check if abigen exe is exists
if ! [ -x "$(command -v abigen)" ]; then
	echo "debug: abigen is not installed."
	echo "debug: installing abigen."
	go build {ABIGENPATH}/main.go
else
	echo "debug: abigen is installed."
fi

generate_abi_bin
generate_go_modules
build_tps_monitor
build_gohammer

