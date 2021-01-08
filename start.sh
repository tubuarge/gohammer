#!/bin/bash

#function usage() {
#	echo""
#}

ABIGENPATH=""
GOETHEREUMPATH=""
CONTRACT=""

function buildAbigen() {
	# check if go-ethereum exists
	if [[ -f "$GOETHEREUMPATH" ]]; then
		echo "debug: goethereum found $GOETHEREUMPATH ."
	else
		echo "debug: not found go-ethereum"
		echo "getting go-ethereum"
		go get -d github.com/ethereum/go-ethereum
		GOETHEREUMPATH="${GOPATH}/src/github.com/ethereum/go-ethereum"
		ABIGENPATH="${GOPATH}/src/github.com/ethereum/go-ethereum/cmd/abigen"
	fi
}

function usage() {
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
			CONTRACT=$2
			shift2
			;;
		*)
			echo "Error: Unsportted command line parameter $1"
			usage
			;;
	esac
done


# check GOPATH is set.
if [[ -z "${GOPATH}"]]; then
	echo "debug: not found gopath"
	echo "Check your GOPATH variable"
	exit 1
else
	echo "debug: found gopath ${GOPATH} "
fi

# check if abigen path exists
if [[ -z "$ABIGENPATH"]]; then
	echo "debug: abigen not found."
else
	ABIGENPATH="${GOPATH}/src/github.com/ethereum/go-ethereum/cmd/abigen"
	GOETHEREUMPATH="${GOPATH}/src/github.com/ethereum/go-ethereum"
fi

# check if abigen exe is exists
if ! [ -x "$(command -v abigen)" ]; then
	echo "debug: abigen is not installed."
	echo "debug: installing abigen."
	go build {ABIGENPATH}/main.go
else
	echo "debug: abigen is installed."
fi
