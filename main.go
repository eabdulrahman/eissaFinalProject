package main

// import fyne
import (
	"database/sql"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/codingsince1985/geo-golang"
	"github.com/codingsince1985/geo-golang/openstreetmap"
	simpleMap "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	_ "github.com/mattn/go-sqlite3" //import for side effects
	g "github.com/serpapi/google-search-results-golang"
	"image/color"
	"log"
	"math/rand"
)

var LocationCache map[string]*geo.Location
var JobCounts map[string]int

type Entries struct {
	jobTitle     string
	companyName  string
	location     string
	via          string
	description  string
	highlights   string
	extensions   string
	relatedLinks string
}

func apiCall(jobTitle string, jobLocation string) []Entries {
	parameter := map[string]string{
		"q":        jobTitle,
		"location": jobLocation,
	}
	var myData []Entries
	apiKey := "70071f223a2b63722d99cc77bb06a627fb2379ab1706729007944100befcbd53"
	engine := "google_jobs"
	search := g.NewSearch(engine, parameter, apiKey)
	data, err := search.GetJSON()
	if err != nil {
		fmt.Println(err)
	}

	results := data["jobs_results"].([]interface{})
	for i := range results {
		first_result := results[i].(map[string]interface{})
		var descr string
		var highls string
		var rltdLinks string
		var loc string
		highls = fmt.Sprintf("%v", first_result["job_highlights"])
		descr = fmt.Sprintf("%v", first_result["description"])
		rltdLinks = fmt.Sprintf("%v", first_result["related_links"])
		loc = fmt.Sprintf("%v", first_result["location"])
		myData2 := Entries{
			jobTitle:     fmt.Sprintf("%v", first_result["title"]),
			companyName:  fmt.Sprintf("%v", first_result["company_name"]),
			location:     loc[2:],
			description:  descr[:200],
			highlights:   highls[12:200],
			extensions:   fmt.Sprintf("%v", first_result["detected_extensions"]),
			relatedLinks: rltdLinks[10:100],
		}
		myData = append(myData, myData2)
	}
	return myData
}

func main() {
	//a variable to specify a row number
	LocationCache = make(map[string]*geo.Location)
	JobCounts = make(map[string]int)

	myDatabase := OpenDataBase("./jobposts.db")
	defer myDatabase.Close()
	createDBTable(myDatabase)

	// new app
	a := app.New()
	// new title and window
	w := a.NewWindow("Jobs Portal")
	// resize window
	w.Resize(fyne.NewSize(1920, 1080))

	var myData []Entries

	//mapCenter := findLocation("Columbus, OH")
	//showMap(rows, mapCenter)
	//addDatatoDB(myDatabase, myData)

	list := widget.NewList(
		// first argument is item count/length
		func() int { return len(myData) },
		// 2nd argument is for widget choice. I want to use label
		func() fyne.CanvasObject { return widget.NewLabel("") },
		// 3rd argument is to update widget with our new data
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(myData[lii].jobTitle + ", " + myData[lii].location)
		},
	)

	searchJobTitle := widget.NewEntry()
	searchJobTitle.SetPlaceHolder("Enter Job Title here")
	searchLocation := widget.NewEntry()
	searchLocation.SetPlaceHolder("Enter Location Here")
	searchButton := widget.NewButton("Search", func() {

		myData = apiCall(searchJobTitle.Text, searchLocation.Text)
		fmt.Println(myData)
		list.Refresh()
		//insertDataIntoDB(myDatabase, myData)

	})

	label1 := widget.NewLabel("")
	label2 := widget.NewLabel("")
	label3 := widget.NewLabel("")
	label4 := widget.NewLabel("")
	label5 := widget.NewLabel("")
	label6 := widget.NewLabel("")
	label7 := widget.NewLabel("")

	//image := canvas.NewImageFromResource(theme.FyneLogo())
	image := canvas.NewImageFromFile("map.png")
	image.SetMinSize(fyne.NewSize(100, 150))
	//image.FillMode = canvas.ImageFillOriginal

	//when an item is selected on the window, display its values
	list.OnSelected = func(id widget.ListItemID) {

		label1.Text = "Job Title \n" + myData[id].jobTitle
		label2.Text = "Company Name \n" + myData[id].companyName
		label3.Text = "Location \n" + myData[id].location
		label4.Text = "Job Description \n" + myData[id].description
		label5.Text = "Highlights \n" + myData[id].highlights
		label6.Text = "Extension \n" + myData[id].extensions
		label7.Text = "Related Links \n" + myData[id].relatedLinks

		city := findLocation(myData[id].location)
		showMap(city)

		label1.Refresh()
		label2.Refresh()
		label3.Refresh()
		label4.Refresh()
		label5.Refresh()
		label6.Refresh()
		label7.Refresh()
		image.Refresh()

	}

	// show and run
	w.SetContent(container.NewHSplit(list,
		container.NewVBox(searchJobTitle, searchLocation, searchButton, label1, label2, label3, label4, label5, label6, label7, image),
	),
	)
	w.ShowAndRun()
}

//The code of the below function was copied from the previous lecture
func showMap(loc *geo.Location) {
	//processData(data, loc)
	context := simpleMap.NewContext()
	context.SetSize(1200, 1200)
	context.SetZoom(6)

	/*
		for city, numJobs := range JobCounts {
			cityLoc := LocationCache[city]
			pinColor := getColor(numJobs)
			context.AddObject(simpleMap.NewMarker(
				s2.LatLngFromDegrees(cityLoc.Lat, cityLoc.Lng),
				pinColor,
				16,
			),
			)
		}

	*/

	pinColor := getColor(1)
	context.AddObject(simpleMap.NewMarker(
		s2.LatLngFromDegrees(loc.Lat, loc.Lng),
		pinColor,
		16,
	),
	)

	context.SetCenter(s2.LatLngFromDegrees(loc.Lat, loc.Lng))
	image, err := context.Render()
	if err != nil {
		log.Fatalln(err)
	}

	if err := gg.SavePNG("map.png", image); err != nil {
		fmt.Println(err)
	}
}

//The code of the below function was copied from the previous lecture
func findLocation(city string) *geo.Location {
	geoLookup := openstreetmap.Geocoder()
	locationData, err := geoLookup.Geocode(city)
	if err != nil {
		log.Println("Error looking up location:", city, err)
	}

	return locationData
}

//getColor code is copied from the previous lecture
func getColor(numberOfjobs int) color.RGBA {
	if numberOfjobs > 75 {
		return color.RGBA{G: 0xff, A: 0xff}
	}
	if numberOfjobs > 50 {
		return color.RGBA{0, 0xff, 0xff, 0xff}
	}

	if numberOfjobs > 25 {
		return color.RGBA{0, 0, 0xff, 0xff}
	}
	if numberOfjobs > 10 {
		return color.RGBA{0xff, 0, 0xff, 0xff}
	}
	if numberOfjobs > 5 {
		return color.RGBA{0xff, 0, 0, 0xff}
	}
	if numberOfjobs > 1 {
		return color.RGBA{0x10, 0x10, 0x10, 0xff}
	}

	return color.RGBA{}
}

//getColor code is copied from the previous lecture
func processData(data [][]string, defaultLocation *geo.Location) {

	for rowNumber, job := range data {
		if rowNumber < 1 {
			continue
		}
		cityName := job[4]
		_, ok := LocationCache[cityName]
		if ok {
			JobCounts[cityName]++
		} else {
			loc := findLocation(cityName)
			if loc == nil {
				loc = defaultLocation
			}
			LocationCache[cityName] = loc
			JobCounts[cityName] = 1
		}
	}

}

// from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
var letterRunes = []rune("abcdef0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//end from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go

func OpenDataBase(dbfile string) *sql.DB {
	database, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	return database
}

func createDBTable(database *sql.DB) {
	createStatement2 := "CREATE TABLE IF NOT EXISTS Job_Posts( " +
		"Job_ID TEXT PRIMARY KEY," +
		"Job_Title TEXT DEFAULT NA," +
		"Company_Name TEXT DEFAULT NA," +
		"Location TEXT DEFAULT NA," +
		"Description TEXT DEFAULT NA," +
		"Highlights TEXT DEFAULT NA," +
		"Extension TEXT DEFAULT NA," +
		"Related_Links TEXT DEFAULT NA);"
	database.Exec(createStatement2)
}

func insertDataIntoDB(database *sql.DB, myData []Entries) {
	statement := "INSERT INTO Job_Posts (Job_ID, Job_Title, Company_Name, Location, Description, Highlights, Extension, Related_Links)" +
		"VALUES (%s, %s, %s, %s, %s,%s, %s, %s);"
	jobId := RandStringRunes
	for id, _ := range myData {
		prepped_statement, err := database.Prepare(statement)
		if err != nil {
			//cowardly bail out since this is academia
			log.Fatal(err)
		}

		filled_statement := fmt.Sprintf(statement, jobId, myData[id].jobTitle, myData[id].companyName, myData[id].location, myData[id].description, myData[id].highlights, myData[id].extensions, myData[id].relatedLinks)

		prepped_statement.Exec(filled_statement)
	}
}
