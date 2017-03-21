package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OncallOptions defines possible options for OnCall resolver
type OncallOptions struct {
	SheetID       string
	DefaultOnDuty string
	OffShiftStart string
	OffShiftStop  string
}

// NewOnCall returns OnCall resolver instance
func NewOnCall(opts *options) *OnCall {
	OffShiftStart, err := strconv.Atoi(opts.OffShiftStart)
	if err != nil {
		OffShiftStart = 0
	}
	OffShiftStop, err := strconv.Atoi(opts.OffShiftStop)
	if err != nil {
		OffShiftStop = 7
	}

	onCall := &OnCall{
		SheetsURL:     "https://docs.google.com/spreadsheets/d/",
		SheetID:       opts.SheetID,
		DefaultOnDuty: opts.DefaultOnDuty,
		OffShiftStart: OffShiftStart,
		OffShiftStop:  OffShiftStop,
	}

	return onCall
}

// OnCall resolver
type OnCall struct {
	SheetsURL     string
	SheetID       string
	DefaultOnDuty string
	OffShiftStart int
	OffShiftStop  int
}

// WhoIsOnCall returns the current active duty
func (o *OnCall) WhoIsOnCall() string {
	t := time.Now()
	year, month, day := t.Date()
	t1 := time.Date(year, month, day, o.OffShiftStart, 0, 0, 0, t.Location())
	t2 := time.Date(year, month, day, o.OffShiftStop, 0, 0, 0, t.Location())

	onDuty := o.DefaultOnDuty

	// NightShift == DefaultOnDuty
	if inTimeSpan(t1, t2, t) {
		return onDuty
	}

	res, err := http.Get(o.SheetsURL + o.SheetID + "/export?&exportFormat=csv&&gid=0")
	if err != nil {
		log.Fatalln(err)
		return onDuty
	}
	defer res.Body.Close()

	reader := csv.NewReader(res.Body)
	csvData, err := reader.ReadAll()
	if err != nil {
		log.Fatalln(err)
		return onDuty
	}

	for _, row := range csvData {
		if strings.EqualFold(row[2], "oncall") {
			onDuty = row[1]
		}
	}

	return onDuty
}

// checks whether given `check` time is between two specified times
func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}
