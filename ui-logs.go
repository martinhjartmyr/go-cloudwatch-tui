package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	logsTable        *tview.Table
	logsFlex         *tview.Flex
	logsMenu         *tview.TextView
	logCount         int32
	logLimit         int32
	activeLogGroup   string
	activeLogStreams []string
	isTailing        bool
	stream           *cloudwatchlogs.StartLiveTailEventStream
)

func createLogs() tview.Primitive {
	isTailing = false
	activeLogGroup = ""
	activeLogStreams = []string{}
	logCount = 0
	logLimit = 50
	logsTable = tview.NewTable()
	logsTable.SetBackgroundColor(ColorBg)
	logsTable.SetFixed(1, 2)
	logsTable.SetSelectable(true, false)
	logsTable.SetCell(0, 0, tview.NewTableCell("Time").SetTextColor(ColorFg).SetSelectable(false))
	logsTable.SetCell(0, 1, tview.NewTableCell("Text").SetTextColor(ColorFg).SetExpansion(1).SetSelectable(false))

	logsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case LogsTailKey.KeyRune:
			if isTailing {
				stopTailing()
				drawUILogsMenu()
			} else {
				startTailing()
				drawUILogsMenu()
			}
			return nil
		case LogsChangeKey.KeyRune:
			addConsoleRow("switch to search")
			if !pages.HasPage("search") {
				pages.AddAndSwitchToPage("search", createSearch(), true)
			} else {
				pages.SwitchToPage("search")
			}
			return nil
		case LogsRefreshKey.KeyRune:
			addConsoleRow("Refresh logs")
			currentActiveLogGroup := getActiveLogGroup()
			if currentActiveLogGroup != "None" {
				go fetchLatestLogs(currentActiveLogGroup)
			} else {
				addConsoleRow("No active log group to refresh")
			}
			return nil
		}

		return event
	})

	logsMenu = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft)
	logsMenu.SetBackgroundColor(ColorSecondaryMenuBg)

	drawUILogsMenu()

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(logsMenu, 1, 1, false).
		AddItem(logsTable, 0, 1, true)

	return layout
}

func drawUILogsMenu() {
	var menuList []string
	menuList = append(menuList, fmt.Sprintf("<%s>%s", LogsChangeKey.KeyLabel, fmt.Sprintf("Log[%s]", getActiveLogGroup())))

	if len(activeLogGroup) != 0 {
		menuList = append(menuList, fmt.Sprintf("<%s>%s", LogsRefreshKey.KeyLabel, LogsRefreshKey.KeyDesc))
	}

	if isTailing {
		menuList = append(menuList, fmt.Sprintf("<%s>%s", LogsTailKey.KeyLabel, "Stop tailing"))
		logsMenu.SetBackgroundColor(tcell.ColorRed)
		logsMenu.SetTextColor(ColorBlack)
	} else {
		menuList = append(menuList, fmt.Sprintf("<%s>%s", LogsTailKey.KeyLabel, LogsTailKey.KeyDesc))
		logsMenu.SetBackgroundColor(ColorSecondaryMenuBg)
		logsMenu.SetTextColor(ColorFg)

	}

	menuString := strings.Join(menuList, " ")
	logsMenu.SetText(menuString)
}

func getActiveLogGroup() string {
	if len(activeLogGroup) == 0 {
		return "None"
	} else {
		// return everything after the last colon in the first log group
		return activeLogGroup[strings.LastIndex(activeLogGroup, ":")+1:]
	}
}

func startTailing() {
	isTailing = true
	addConsoleRow("start tailing: " + activeLogGroup)
	activeTailGroups := []string{}
	activeTailGroups = append(activeTailGroups, activeLogGroup)

	request := &cloudwatchlogs.StartLiveTailInput{
		LogGroupIdentifiers: activeTailGroups,
	}

	response, err := client.StartLiveTail(context.TODO(), request)
	if err != nil {
		addConsoleRow(fmt.Sprintf("Error starting live tail: %v", err))
		isTailing = false
		return
	}

	stream = response.GetStream()
	go handleEventStreamAsync(stream)
}

func stopTailing() {
	addConsoleRow("closing tail")
	if stream != nil {
		stream.Close()
	}
	isTailing = false
}

func addLogRow(timestamp int64, logGroup, text string) {
	if logCount >= logLimit {
		return
	}
	logCount++
	rowCount := logsTable.GetRowCount()

	// convert time to human readable format
	t := time.Unix(timestamp/1000, 0)
	timeStr := t.Format("2006-01-02 15:04:05")

	logsTable.SetCell(rowCount, 0, tview.NewTableCell(timeStr).SetTextColor(ColorFg).SetSelectable(true))
	logsTable.SetCell(rowCount, 1, tview.NewTableCell(text).SetTextColor(ColorFg).SetSelectable(true))

	// select row
	logsTable.Select(rowCount, 0)
}

func resetTable() {
	addConsoleRow("Logs table reset")
	logsTable.Clear()
	logsTable.SetCell(0, 0, tview.NewTableCell("Time").SetTextColor(ColorFg).SetSelectable(false))
	logsTable.SetCell(0, 1, tview.NewTableCell("Text").SetTextColor(ColorFg).SetExpansion(1).SetSelectable(false))
	logCount = 1
}

func handleEventStreamAsync(stream *cloudwatchlogs.StartLiveTailEventStream) {
	eventsChan := stream.Events()
	for {
		event := <-eventsChan
		switch e := event.(type) {
		case *types.StartLiveTailResponseStreamMemberSessionStart:
			addConsoleRow("Received SessionStart event")
		case *types.StartLiveTailResponseStreamMemberSessionUpdate:
			for _, logEvent := range e.Value.SessionResults {
				app.QueueUpdateDraw(func() {
					addLogRow(*logEvent.Timestamp, *logEvent.LogStreamName, *logEvent.Message)
				})
			}
		default:
			// Handle on-stream exceptions
			if err := stream.Err(); err != nil {
				log.Fatalf("Error occurred during streaming: %v", err)
			} else if event == nil {
				app.QueueUpdateDraw(func() {
					addConsoleRow("Stream is Closed")
				})
				return
			} else {
				log.Fatalf("Unknown event type: %T", e)
			}
		}
	}
}
