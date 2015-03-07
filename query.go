package query

import (
	"encoding/xml"
	"fmt"
	"github.com/njwilson23/geometry"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	Landsat8     = "LANDSAT_8"
	Landsat7     = "LANDSAT_ETM"
	Landsat7SLC  = "LANDSAT_ETM_SLC_OFF"
	Landsat45TM  = "LANDSAT_TM"
	Landsat45MSS = "LANDSAT_MSS2"
	Landsat13MSS = "LANDSAT_MSS1"
	LandsatAll   = "LANDSAT_COMBINED"
)

type QueryParameters struct {
	Bbox   [4]float64
	Dates  [2]time.Time
	Sensor string // "LANDSAT_8"
}

type LandsatScene struct {
	SceneID             string  `xml:"sceneID"`
	BrowseURL           string  `xml:"browseURL"`
	CloudCover          float32 `xml:"cloudCoverFull"`
	DayOrNight          string  `xml:"dayOrNight"`
	StartTime           string  `xml:"sceneStartTime"`
	EndTime             string  `xml:"sceneEndTime"`
	UpperLeftLatitude   float64 `xml:"upperLeftCornerLatitude"`
	UpperLeftLongitude  float64 `xml:"upperLeftCornerLongitude"`
	LowerLeftLatitude   float64 `xml:"lowerLeftCornerLatitude"`
	LowerLeftLongitude  float64 `xml:"lowerLeftCornerLongitude"`
	UpperRightLatitude  float64 `xml:"upperRightCornerLatitude"`
	UpperRightLongitude float64 `xml:"upperRightCornerLongitude"`
	LowerRightLatitude  float64 `xml:"lowerRightCornerLatitude"`
	LowerRightLongitude float64 `xml:"lowerRightCornerLongitude"`
}

func (s LandsatScene) Poly() geometry.Polygon {
	res := geometry.Polygon{geometry.MultiPoint{
		[]float64{s.LowerLeftLongitude, s.LowerRightLongitude,
			s.UpperRightLongitude, s.UpperLeftLongitude},
		[]float64{s.LowerLeftLatitude, s.LowerRightLatitude,
			s.UpperRightLatitude, s.UpperLeftLatitude}}}
	return res
}

func (s LandsatScene) String() string {
	p := s.Poly()
	bbox := p.Bbox()
	return fmt.Sprintf("Scene: %v\nDate: %v\nURL: %v\nBbox: %v %v %v %v",
		s.SceneID, s.StartTime, s.BrowseURL,
		bbox[0], bbox[1], bbox[2], bbox[3])
}

type XMLReturnStatus struct {
	Status string `xml:"value, attr"`
}

type XMLResponse struct {
	XMLName xml.Name        `xml:"searchResponse"`
	Scenes  []LandsatScene  `xml:"metaData"`
	Status  XMLReturnStatus `xml:"returnStatus"`
}

// Send a request to the Landsat Bulk Metadata server, and return the XML
// response as bytes
func Request(q QueryParameters) ([]byte, error) {
	var result []byte

	const datefmt = "2006-01-02"
	request := fmt.Sprintf(
		`http://earthexplorer.usgs.gov/EE/InventoryStream/latlong?north=%v&south=%v&east=%v&west=%v&sensor=%v&start_date=%v&end_date=%v`,
		q.Bbox[3], q.Bbox[2], q.Bbox[1], q.Bbox[0],
		q.Sensor,
		q.Dates[0].Format(datefmt), q.Dates[1].Format(datefmt))

	resp, err := http.Get(request)
	if err != nil {
		fmt.Println(err)
		return result, err
	}

	defer resp.Body.Close()
	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return result, err
}

// Convert an XML response into a list of Scene structs
func ParseXMLBytes(r []byte) ([]LandsatScene, error) {
	response := XMLResponse{}
	err := xml.Unmarshal(r, &response)
	if err != nil {
		fmt.Println(err)
	}
	return response.Scenes, err
}
