package query

import (
	"fmt"
	"testing"
	"time"
)

func TestRequest(t *testing.T) {

	q := NewQuery()
	q.ByDateRange(
		time.Date(2015, time.February, 25, 0, 0, 0, 0, time.UTC),
		time.Date(2015, time.March, 7, 0, 0, 0, 0, time.UTC))
	q.ByBbox(-71, -70, 41, 42)
	q.BySensor(Landsat8)

	req, err := Request(q)

	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if len(req) == 0 {
		t.Fail()
	}

}

func TestParseRequest(t *testing.T) {

	q := NewQuery()
	q.ByDateRange(
		time.Date(2015, time.February, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2015, time.March, 7, 0, 0, 0, 0, time.UTC))
	q.ByBbox(-72, -69, 41, 42)
	q.BySensor(Landsat8)

	result, err := Request(q)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	scenes, _ := ParseXMLBytes(result)
	if len(scenes) != 2 {
		t.Fail()
	}
	/*for i, s := range scenes {
		fmt.Println(i, s)
	}*/
}

func TestParseRequestPathRow(t *testing.T) {

	q := NewQuery()
	q.ByDateRange(
		time.Date(2014, time.February, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2015, time.March, 7, 0, 0, 0, 0, time.UTC))
	q.ByWRSPath(8, 8).ByWRSRow(3, 3)
	q.BySensor(Landsat8)

	result, err := Request(q)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	scenes, _ := ParseXMLBytes(result)
	if len(scenes) != 10 {
		t.Fail()
	}
	/*for i, s := range scenes {
		fmt.Println(i, s)
	}*/
}
