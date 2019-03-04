package main

import (
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/aws"
)

func main() {
	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-prod"))
	if err != nil {
		log.WithError(err).Fatal("setting up credentials")
		return
	}
	cfg.Region = endpoints.ApSoutheast1RegionID
	svc := cloudwatchlogs.New(cfg)
	from := time.Now().Add(-time.Hour*time.Duration(2)).Unix() * 1000
	log.WithField("from", from).Info("last 2 hours")

	req := svc.FilterLogEventsRequest(&cloudwatchlogs.FilterLogEventsInput{
		FilterPattern: aws.String(`[..., request = *HTTP*, status_code = 5**, , ,]`),
		LogGroupName:  aws.String("bugzilla"),
		StartTime:     &from,
	})

	p := req.Paginate()
	for p.Next() {
		page := p.CurrentPage()
		if len(page.Events) > 0 {
			log.Infof("Events: %#v", len(page.Events))

			previous, err := findPreviousLog(svc, &page.Events[0])
			if err != nil {
				panic(err)
			}
			log.Infof("Previous log: %s", previous)
		}
	}

	if err := p.Err(); err != nil {
		panic(err)
	}

}

func findPreviousLog(client *cloudwatchlogs.CloudWatchLogs, event *cloudwatchlogs.FilteredLogEvent) (string, error) {
	log.WithField("streamName", *event.LogStreamName).Info("streamName")
	req := client.GetLogEventsRequest(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String("bugzilla"),
		LogStreamName: event.LogStreamName,
	})
	p := req.Paginate()
	for p.Next() {
		page := p.CurrentPage()
		if len(page.Events) > 0 {
			log.Infof("Events from findPreviousLog: %#v", len(page.Events))
		}
		for i, e := range page.Events {
			// log.WithFields(log.Fields{
			// 	"e":     *e.Message,
			// 	"event": *event.Message,
			// }).Info("searching")
			if *event.Message == *e.Message {
				log.Info("BINGO")
				return *page.Events[i-1].Message, nil
			}
		}
	}

	return "", p.Err()

}
