package tpsmon

import (
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type InfluxdbMetricsService struct {
	endpoint string
	token    string
	org      string
	bucket   string
	point    string
	tags     map[string]string
	client   influxdb2.Client
	writeApi api.WriteApiBlocking
}

func NewInfluxdbService(ep string, token string, org string, bucket string, point string, tags string) (*InfluxdbMetricsService, error) {
	if _, err := url.Parse(ep); err != nil {
		return nil, err
	}
	tagsMap := stringToTags(tags)
	log.Debugf("Influxdb: endpoint:%s token:%s org:%s bucket:%s point:%s tags:%v\n", ep, token, org, bucket, point, tagsMap)
	is := &InfluxdbMetricsService{endpoint: ep, token: token, org: org, bucket: bucket, point: point, tags: tagsMap}
	is.makeClient()
	return is, nil
}

func (id *InfluxdbMetricsService) makeClient() {
	id.client = influxdb2.NewClient(id.endpoint, id.token)
	id.writeApi = id.client.WriteApiBlocking(id.org, id.bucket)
	log.Info("influxdb client created")
}

func (id *InfluxdbMetricsService) PushMetrics(tm time.Time, tps uint64, txns uint64, blocks uint64) {
	p := influxdb2.NewPoint(id.point,
		id.tags,
		map[string]interface{}{
			"tps":          float64(tps),
			"transactions": float64(txns),
			"blocks":       float64(blocks),
		},
		tm)
	if err := id.writeApi.WritePoint(context.Background(), p); err != nil {
		log.Errorf("influxdb write failed error: %v", err)
		id.makeClient()
		return
	}
	log.Info("pushed tps metrics to influxdb.")
}

func stringToTags(tagsFlag string) map[string]string {
	tags := strings.Split(tagsFlag, ",")
	tagsMap := map[string]string{}

	for _, t := range tags {
		if t != "" {
			kv := strings.Split(t, "=")

			if len(kv) == 2 {
				tagsMap[kv[0]] = kv[1]
			}
		}
	}

	return tagsMap
}
