package query

import (
	"fmt"
	"testing"
	"time"
)

func TestRequest(t *testing.T) {
	q := QueryParameters{[4]float64{-71, -70, 41, 42},
		[2]time.Time{time.Date(2015, time.February, 25, 0, 0, 0, 0, time.UTC),
			time.Date(2015, time.March, 7, 0, 0, 0, 0, time.UTC)},
		Landsat8}
	req, err := Request(q)
	if err != nil {
		fmt.Println(err)
	}
	if len(req) == 0 {
		t.Fail()
	}
}

func TestParseRequest(t *testing.T) {
	q := QueryParameters{[4]float64{-71, -70, 41, 42},
		[2]time.Time{time.Date(2015, time.February, 25, 0, 0, 0, 0, time.UTC),
			time.Date(2015, time.March, 7, 0, 0, 0, 0, time.UTC)},
		Landsat8}

	req, err := Request(q)
	if err != nil {
		fmt.Println(err)
	}

	scenes, _ := ParseXMLBytes(req)
	for i, s := range scenes {
		if s.CloudCover < 1 {
			fmt.Println(i, s)
		}
	}
}
