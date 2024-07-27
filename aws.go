package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"gopkg.in/ini.v1"
)

type Profiles struct {
	Error string   `json:"error"`
	Data  []string `json:"data"`
}

var client *cloudwatchlogs.Client

func listAwsProfiles() ([]string, error) {
	configPath := config.DefaultSharedCredentialsFilename()
	f, err := ini.Load(configPath)
	profiles := []string{}
	if err != nil {
		return nil, err
	} else {
		for _, v := range f.Sections() {
			if len(v.Keys()) != 0 {
				profiles = append(profiles, v.Name())
			}
		}
	}
	return profiles, nil
}

func loadAWSConfig(profile string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile))
	if err == nil {
		client = cloudwatchlogs.NewFromConfig(cfg)
	}

	return cfg, err
}

func listCloudWatchGroups(pattern string) []logGroup {
	params := &cloudwatchlogs.DescribeLogGroupsInput{LogGroupNamePattern: &pattern}
	resp, err := client.DescribeLogGroups(context.TODO(), params)
	if err != nil {
		addConsoleRow(fmt.Sprintf("Error listing log groups:) %v", err))
		log.Fatal(err)
	}

	logGroups := []logGroup{}
	for _, group := range resp.LogGroups {
		addConsoleRow(*group.LogGroupName)
		logGroups = append(logGroups, logGroup{Name: *group.LogGroupName, Arn: *group.Arn, Profile: activeAwsProfile})
	}

	return logGroups
}

func listCloudWatchStreams(group string, limit int32) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
	descending := true
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		Descending:   &descending,
		LogGroupName: &group,
		Limit:        &limit,
	}
	resp, err := client.DescribeLogStreams(context.TODO(), params)
	if err != nil {
		log.Fatal(err)
		return nil, nil
	}
	// reverse the order of streams
	for i, j := 0, len(resp.LogStreams)-1; i < j; i, j = i+1, j-1 {
		resp.LogStreams[i], resp.LogStreams[j] = resp.LogStreams[j], resp.LogStreams[i]
	}

	return resp, nil
}

func fetchLogs(logStream string, group string, limit int32) (*cloudwatchlogs.GetLogEventsOutput, error) {
	params := &cloudwatchlogs.GetLogEventsInput{
		LogGroupIdentifier: &group,
		LogStreamName:      &logStream,
		Limit:              &limit,
	}

	resp, err := client.GetLogEvents(context.TODO(), params)
	if err != nil {
		log.Fatal(err)
	}

	// sort events based on timestamp
	sort.Slice(resp.Events, func(i, j int) bool {
		return *resp.Events[i].Timestamp < *resp.Events[j].Timestamp
	})

	for _, event := range resp.Events {
		app.QueueUpdateDraw(func() {
			addLogRow(*event.Timestamp, group, *event.Message)
		})
	}
	return resp, nil
}

func fetchLatestLogs(logGroup string) {
	app.QueueUpdateDraw(func() {
		resetTable()
		addConsoleRow("Fetching logs for group: " + logGroup + " limit: " + fmt.Sprint(logLimit))
	})

	logStreams, err := listCloudWatchStreams(logGroup, 5)
	if err != nil {
		panic(err)
	}
	for _, stream := range logStreams.LogStreams {
		if logCount <= logLimit {
			app.QueueUpdateDraw(func() {
				addConsoleRow("Fetching logs for stream: " + *stream.LogStreamName)
			})
			fetchLogs(*stream.LogStreamName, logGroup, logLimit)
		}
	}
}
