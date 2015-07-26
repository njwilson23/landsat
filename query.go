package query

import (
	"encoding/xml"
	"fmt"
	"github.com/njwilson23/geometry"
	"strings"
	//"html/template"
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
	Bbox          [4]float64
	Dates         [2]time.Time
	Sensor        string // "LANDSAT_8"
	WRSPathStart  int32
	WRSPathEnd    int32
	WRSRowStart   int32
	WRSRowEnd     int32
	CloudCoverMax int32
}

func (q *QueryParameters) ByBbox(lon0 float64, lon1 float64, lat0 float64, lat1 float64) *QueryParameters {
	q.Bbox = [4]float64{lon0, lon1, lat0, lat1}
	return q
}

func (q *QueryParameters) ByDateRange(t0 time.Time, t1 time.Time) *QueryParameters {
	q.Dates = [2]time.Time{t0, t1}
	return q
}

func (q *QueryParameters) BySensor(sens string) *QueryParameters {
	q.Sensor = sens
	return q
}

func (q *QueryParameters) ByWRSPath(path0 int32, path1 int32) *QueryParameters {
	q.WRSPathStart = path0
	q.WRSPathEnd = path1
	return q
}

func (q *QueryParameters) ByWRSRow(row0 int32, row1 int32) *QueryParameters {
	q.WRSRowStart = row0
	q.WRSRowEnd = row1
	return q
}

func (q *QueryParameters) ByCloudCover(cc int32) *QueryParameters {
	q.CloudCoverMax = cc
	return q
}

func NewQuery() *QueryParameters {
	q := QueryParameters{
		[4]float64{-180, 180, -90, 90}, // Initialize with global Bbox
		[2]time.Time{
			time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(3000, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		LandsatAll,
		-1, -1, -1, -1, // WRS fields
		-1}
	return &q
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
//
// Request schema at http://earthexplorer.usgs.gov/EE/metadata.xsd
func Request(q *QueryParameters) ([]byte, error) {
	var result []byte
	var request string

	const datefmt = "2006-01-02"
	//request := "http://earthexplorer.usgs.gov/EE/InventoryStream/"
	if q.WRSPathStart+q.WRSPathEnd+q.WRSRowStart+q.WRSRowEnd == -4 {
		request = fmt.Sprintf(
			`http://earthexplorer.usgs.gov/EE/InventoryStream/latlong?north=%v&south=%v&east=%v&west=%v`,
			q.Bbox[3], q.Bbox[2], q.Bbox[1], q.Bbox[0])
	} else {
		request = fmt.Sprintf(
			`http://earthexplorer.usgs.gov/EE/InventoryStream/pathrow?start_path=%v&end_path=%v&start_row=%v&end_row=%v`,
			q.WRSPathStart, q.WRSPathEnd, q.WRSRowStart, q.WRSRowEnd)
	}

	sensor := fmt.Sprintf(`&sensor=%v`, q.Sensor)

	timespan := fmt.Sprintf(`&start_date=%v&end_date=%v`,
		q.Dates[0].Format(datefmt), q.Dates[1].Format(datefmt))

	request = strings.Join([]string{request, sensor, timespan}, "")

	if q.CloudCoverMax != -1 {
		cc := fmt.Sprintf(`&cc=%v`, q.CloudCoverMax)
		request = strings.Join([]string{request, cc}, "")
	}

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

// Convert an XML byte data into a list of Scene structs
func ParseXMLBytes(data []byte) ([]LandsatScene, error) {
	response := XMLResponse{}
	err := xml.Unmarshal(data, &response)
	if err != nil {
		fmt.Println(err)
	}
	return response.Scenes, err
}

// Display a list of Landsat scenes in an HTML table
/*func DisplayWeb(ls []LandsatScene) (string, error) {
	//t, err := template.New("lsgrid").ParseFiles("templates/lsgrid.html")
	t, err := template.New("lsgrid").Parse("<div class=\"scene\" style=\"background-image: {{BrowseUrl}}\"></div>")
	if err != nil {
		fmt.Println(err)
	}
	t.ExecuteTemplate(out, "lsgrid", []LandsatScene)
	return "", err
}*/
