package main

import (
	"flag"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

func main() {
	hours := flag.Int("hours", 1, "Hours since it happened")
	flag.Parse()

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-prod"))
	if err != nil {
		log.WithError(err).Fatal("setting up credentials")
		return
	}
	cfg.Region = endpoints.ApSoutheast1RegionID
	svc := cloudwatchlogs.New(cfg)
	from := time.Now().Add(-time.Hour*time.Duration(*hours)).Unix() * 1000
	log.WithField("from", from).Infof("last %d hours", *hours)

	req := svc.FilterLogEventsRequest(&cloudwatchlogs.FilterLogEventsInput{
		FilterPattern: aws.String(`[..., request = *HTTP*, status_code = 5**, , ,]`),
		LogGroupName:  aws.String("bugzilla"),
		StartTime:     &from,
	})

	p := req.Paginate()
	for p.Next() {
		page := p.CurrentPage()
		for _, event := range page.Events {
			previous, err := findPreviousLog(svc, &event)
			if err != nil {
				panic(err)
			}

			log.WithFields(log.Fields{
				"5xx":  *event.Message,
				"prev": previous,
			}).Info("previous log")
		}
	}

	if err := p.Err(); err != nil {
		panic(err)
	}

}

func findPreviousLog(client *cloudwatchlogs.CloudWatchLogs, event *cloudwatchlogs.FilteredLogEvent) (string, error) {
	// log.WithFields(log.Fields{
	// 	"streamName": *event.LogStreamName,
	// 	"event":      *event.Message,
	// }).Info("pulling up previous log")
	req := client.GetLogEventsRequest(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String("bugzilla"),
		LogStreamName: event.LogStreamName,
	})
	p := req.Paginate()
	for p.Next() {
		page := p.CurrentPage()
		for i, e := range page.Events {
			// log.WithFields(log.Fields{
			// 	"e":     *e.Message,
			// 	"event": *event.Message,
			// }).Info("searching")
			if *event.Message == *e.Message {
				return *page.Events[i-1].Message, nil
			}
		}
	}

	return "", p.Err()

}
