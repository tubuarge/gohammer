package tpsmon

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type TPSServer struct {
	tm   *TPSMonitor
	port int
}

func NewTPSServer(tm *TPSMonitor, port int) TPSServer {
	s := TPSServer{
		tm:   tm,
		port: port,
	}
	go s.Start()
	return s
}

func (s TPSServer) Start() {
	http.HandleFunc("/tpsdata", s.PrintTPSData)
	log.Infof("started tps monitor server at port %d", s.port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(s.port), nil))
}

func (s TPSServer) PrintTPSData(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "localTime,refTime,TPS,TxnCount,BlockCount\n")
	for _, v := range s.tm.tpsRecs {
		fmt.Fprintf(w, "%s", v.ReportString())
	}
}
