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
	Bbox    [4]float64
	Dates   [2]time.Time
	Sensor  string // "LANDSAT_8"
	WRSPath int32
	WRSRow  int32
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

func (q *QueryParameters) ByWRSPath(path int32) *QueryParameters {
	q.WRSPath = path
	return q
}

func (q *QueryParameters) ByWRSRow(row int32) *QueryParameters {
	q.WRSRow = row
	return q
}

func NewQuery() *QueryParameters {
	q := QueryParameters{
		[4]float64{-180, 180, -90, 90},
		[2]time.Time{
			time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(3000, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		LandsatAll,
		-1,
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
func Request(q *QueryParameters) ([]byte, error) {
	var result []byte

	const datefmt = "2006-01-02"
	//request := "http://earthexplorer.usgs.gov/EE/InventoryStream/"
	request := fmt.Sprintf(
		`http://earthexplorer.usgs.gov/EE/InventoryStream/latlong?north=%v&south=%v&east=%v&west=%v&sensor=%v&start_date=%v&end_date=%v`,
		q.Bbox[3], q.Bbox[2], q.Bbox[1], q.Bbox[0],
		q.Sensor,
		q.Dates[0].Format(datefmt), q.Dates[1].Format(datefmt))

	if q.WRSPath != -1 {
		pathstr := fmt.Sprintf(`&wrspath=%v`, q.WRSPath)
		request = strings.Join([]string{request, pathstr}, "")
	}

	if q.WRSRow != -1 {
		rowstr := fmt.Sprintf(`&wrsrow=%v`, q.WRSRow)
		request = strings.Join([]string{request, rowstr}, "")
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
