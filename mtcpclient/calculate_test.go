package mtcpclient

import (
	"testing"
	"time"

	"github.com/dachad/tcpgoon/tcpclient"
)

func TestCalculateMetricsReport(t *testing.T) {
	var metricsReportScenariosChecks = []struct {
		scenarioDescription         string
		groupOfConnectionsToReport  *GroupOfConnections
		tcpStatusToReport           tcpclient.ConnectionStatus
		expectedReportWithoutStdDev metricsCollectionStats
	}{
		{
			scenarioDescription:        "Empty group of connections should report 0 as associated metrics",
			groupOfConnectionsToReport: newGroupOfConnections(0),
			tcpStatusToReport:          tcpclient.ConnectionEstablished,
			expectedReportWithoutStdDev: metricsCollectionStats{
				avg:                 0,
				min:                 0,
				max:                 0,
				total:               0,
				numberOfConnections: 0,
			},
		},
		{
			scenarioDescription:        "Single connection should generate a report that describes its associated metric",
			groupOfConnectionsToReport: newSampleSingleConnection(),
			tcpStatusToReport:          tcpclient.ConnectionEstablished,
			expectedReportWithoutStdDev: metricsCollectionStats{
				avg:                 500 * time.Millisecond,
				min:                 500 * time.Millisecond,
				max:                 500 * time.Millisecond,
				total:               500 * time.Millisecond,
				numberOfConnections: 1,
			},
		},
		{
			// TODO: We will need to extend this to cover a mix connections closed + established on closure, when the code supports it
			scenarioDescription:        "Multiple connections with different statuses should generate a report that describes the metrics of the right subset",
			groupOfConnectionsToReport: newSampleMultipleConnections(),
			tcpStatusToReport:          tcpclient.ConnectionError,
			expectedReportWithoutStdDev: metricsCollectionStats{
				avg:                 2 * time.Second,
				min:                 1 * time.Second,
				max:                 3 * time.Second,
				total:               4 * time.Second,
				numberOfConnections: 2,
			},
		},
	}

	for _, test := range metricsReportScenariosChecks {
		resultingReport := test.groupOfConnectionsToReport.calculateMetricsReport()
		test.expectedReportWithoutStdDev.stdDev = test.groupOfConnectionsToReport.calculateStdDev(resultingReport.avg)
		if resultingReport.stdDev != test.expectedReportWithoutStdDev.stdDev {
			t.Error(test.scenarioDescription+", and its", resultingReport)
		}
	}
}

func TestCalculateStdDev(t *testing.T) {
	var stdDevScenariosChecks = []struct {
		scenarioDescription string
		durationsInSecs     []int
		expectedStdDev      int
	}{
		{
			scenarioDescription: "Empty group of connections should report 0 as stats values",
			durationsInSecs:     []int{},
			expectedStdDev:      0,
		},
		{
			scenarioDescription: "Single connection should report a std dev of 0",
			durationsInSecs:     []int{1},
			expectedStdDev:      0,
		},
		{
			scenarioDescription: "Several connections with same durations should report a std dev of 0",
			durationsInSecs:     []int{1, 1, 1, 1, 1},
			expectedStdDev:      0,
		},
		{
			scenarioDescription: "A known set of durations should report a known std dev",
			// https://en.wikipedia.org/wiki/Standard_deviation#Population_standard_deviation_of_grades_of_eight_students
			durationsInSecs: []int{2, 4, 4, 4, 5, 5, 7, 9},
			expectedStdDev:  2,
		},
	}

	for _, test := range stdDevScenariosChecks {

		var gc *GroupOfConnections
		gc = newGroupOfConnections(0)

		var sum int
		var connectionState tcpclient.ConnectionStatus
		for i, connectionDuration := range test.durationsInSecs {
			if i%2 == 0 {
				connectionState = tcpclient.ConnectionEstablished
			} else {
				connectionState = tcpclient.ConnectionClosed
			}
			gc.connections = append(gc.connections, tcpclient.NewConnection(i, connectionState,
				time.Duration(connectionDuration)*time.Second))
			sum += connectionDuration
		}

		var mr *metricsCollectionStats
		mr = newMetricsCollectionStats()

		if len(test.durationsInSecs) != 0 {
			mr.avg = time.Duration(sum/len(test.durationsInSecs)) * time.Second
			mr.numberOfConnections = len(gc.connections)
		}

		stddev := gc.calculateStdDev(mr.avg)

		if stddev != time.Duration(test.expectedStdDev)*time.Second {
			t.Error(test.scenarioDescription+", and its", stddev)
		}
	}
}
