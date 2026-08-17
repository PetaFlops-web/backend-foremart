package repository

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildDailySalesHistoryFillsMissingDatesOldestToNewest(t *testing.T) {
	startDate, err := time.Parse("2006-01-02", "2026-07-20")
	if err != nil {
		t.Fatalf("failed to parse start date: %v", err)
	}

	rows := []dailySalesRow{
		{Date: startDate.AddDate(0, 0, 0), TotalQty: 3},
		{Date: startDate.AddDate(0, 0, 10), TotalQty: 7},
		{Date: startDate.AddDate(0, 0, 29), TotalQty: 11},
	}

	history := buildDailySalesHistory(rows, 30, startDate)

	if len(history) != 30 {
		t.Fatalf("expected 30 history entries, got %d", len(history))
	}

	expected := make([]float64, 30)
	expected[0] = 3
	expected[10] = 7
	expected[29] = 11

	if !reflect.DeepEqual(history, expected) {
		t.Fatalf("unexpected history:\nwant: %#v\n got: %#v", expected, history)
	}
}
