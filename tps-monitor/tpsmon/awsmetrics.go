package tpsmon

import (
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// wrapper for aws cloudwatch to publish metrics
type AwsCloudwatchService struct {
	region   string
	nwName   string
	instance string
	cw       *cloudwatch.CloudWatch
}

func NewCloudwatchService(region string, nwname string, inst string) *AwsCloudwatchService {
	mySession := session.Must(session.NewSession())
	// Create a CloudWatch chainReader with additional configuration
	cw := cloudwatch.New(mySession, aws.NewConfig().WithRegion(region))
	return &AwsCloudwatchService{region, nwname, inst, cw}
}

func (a *AwsCloudwatchService) PutMetrics(mname string, value string, ts time.Time) error {
	var pmd *cloudwatch.PutMetricDataInput
	var mdn *cloudwatch.MetricDatum
	dname := "System"
	dvalue := "TpsMonitor"
	nspace := fmt.Sprintf("%s-%s", a.nwName, a.instance)
	var data float64
	data, _ = strconv.ParseFloat(value, 64)
	dimension := &cloudwatch.Dimension{Name: &dname, Value: &dvalue}
	mdn = &cloudwatch.MetricDatum{
		Dimensions: []*cloudwatch.Dimension{dimension},
		MetricName: &mname,
		Timestamp:  &ts,
		Value:      &data,
	}
	pmd = &cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{mdn},
		Namespace:  &nspace,
	}
	if _, err := a.cw.PutMetricData(pmd); err != nil {
		log.Errorf("aws cloudwatch putmetrics failed for mname:%s value:%s err:%v", mname, value, err)
		return err
	}
	log.Infof("published metric name:%s value:%v to aws cloudwatch", mname, data)
	return nil
}
