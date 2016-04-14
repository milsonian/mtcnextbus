package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//Set global API endpoint
var apiEndpoint = "http://svc.metrotransit.org/NexTrip"

//Helper for get requests to backend API
func apiGetBody(apiPath string) ([]byte, error) {
	request, err := http.NewRequest("GET", apiEndpoint+apiPath, nil)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	response, err := http.DefaultClient.Do(request)
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("bad response code from API: %s", response.Status)
	}
	contentType := response.Header.Get("Content-Type")
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if strings.Contains(contentType, "application/json") == false { //API error responses contain XML, deal with this later..
		return nil, fmt.Errorf("unexpected response content-type from API: %s", contentType)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting body from API: %s", err)
	}
	return body, err
}

//Function to return routeID from string description
func getRouteID(routeDesc string) (string, error) {
	body, err := apiGetBody("/Routes")
	if err != nil {
		return "", fmt.Errorf("error getting routes: %s", err)
	}
	type Route struct {
		Description string
		ProviderID  string
		Route       string
	}

	var routes []Route
	uerr := json.Unmarshal(body, &routes)
	if uerr != nil {
		return "", fmt.Errorf("error processing response body: %s", uerr)
	}

	filteredRoutes := []Route{}
	var routeID string
	for _, r := range routes {
		if strings.HasSuffix(r.Description, routeDesc) { //Handle shortened route strings (no route ID prepend)
			filteredRoutes = append(filteredRoutes, r)
		}
	}

	switch len(filteredRoutes) { //Deal with multiple route responses
	case 1:
		{
			routeID = filteredRoutes[0].Route
		}
	case 0:
		return "", fmt.Errorf("No routes found matching your description")
	default:
		return "", fmt.Errorf("%d matching routes found, please be more specific", len(filteredRoutes))
	}
	return routeID, err
}

//Function to return direction ID from routeID and direction string.  Also returns valid directions for route.
func getDirectionID(routeID string, directionDesc string) (string, error) {
	body, err := apiGetBody(fmt.Sprintf("/Directions/%s", routeID))
	if err != nil {
		return "", fmt.Errorf("error getting directions: %s", err)
	}
	type Direction struct {
		Text  string
		Value string
	}

	var directions []Direction
	uerr := json.Unmarshal(body, &directions)
	if uerr != nil {
		return "", fmt.Errorf("error processing response body: %s", uerr)
	}

	var directionID string
	var validDirections []string
	for _, r := range directions {
		validDirections = append(validDirections, r.Text)
		if strings.HasPrefix(r.Text, strings.ToUpper(directionDesc)) { //Handle shortened direction strings (e.g. no 'bound' suffix)
			directionID = r.Value
		}
	}

	if directionID == "" {
		return "", fmt.Errorf("Requested direction not found for route.\nAvailable directions: %s", strings.Join(validDirections, ", "))

	}
	return directionID, err
}

//Function to return stop ID from routeID, directionID, and stop string
func getStopID(routeID string, directionID string, stopDesc string) (string, error) {
	body, err := apiGetBody(fmt.Sprintf("/Stops/%s/%s", routeID, directionID))
	if err != nil {
		return "", fmt.Errorf("error getting stops: %s", err)
	}
	type Stop struct {
		Text  string
		Value string
	}

	var stops []Stop
	uerr := json.Unmarshal(body, &stops)
	if uerr != nil {
		return "", fmt.Errorf("error processing response body: %s", uerr)
	}

	filteredStops := []Stop{}
	var stopID string
	for _, s := range stops {
		if strings.Contains(s.Text, stopDesc) {
			filteredStops = append(filteredStops, s)
		}
	}

	switch len(filteredStops) { //Deal with multiple stop responses
	case 1:
		{
			stopID = filteredStops[0].Value
		}
	case 0:
		return "", fmt.Errorf("No stops found matching your description")
	default:
		return "", fmt.Errorf("%d matching stops found, please be more specific", len(filteredStops))
	}
	return stopID, err

}

//Function to return time to next departure by routeID, directionID, stopID
func getDeparture(routeID string, directionID string, stopID string) (string, error) {
	body, err := apiGetBody(fmt.Sprintf("/%s/%s/%s", routeID, directionID, stopID))
	if err != nil {
		return "", fmt.Errorf("error getting departures: %s", err)
	}
	type Departure struct {
		Actual        bool
		DepartureText string
		DepartureTime string
	}

	var departures []Departure
	uerr := json.Unmarshal(body, &departures)
	if uerr != nil {
		return "", fmt.Errorf("error processing response body: %s", uerr)
	}

	if len(departures) == 0 {
		return "", fmt.Errorf("No upcoming departures found for today.")
	}
	nextDeparture := departures[0]
	if nextDeparture.Actual == true { //If actual time is provided, return it as trusted
		return fmt.Sprintf("%s (Actual vehicle report)", nextDeparture.DepartureText), err
	}

	//Get the MS datetime from response field
	timeSplit := strings.FieldsFunc(nextDeparture.DepartureTime, func(r rune) bool {
		return r == '(' || r == '-' //Split to the core value, drop timezone offset
	})
	msTimeDate := timeSplit[1]
	convertedTime, err := msToTime(msTimeDate) //..and convert it
	if err != nil {
		return "", fmt.Errorf("error converting time: %s", err)
	}

	timeNow := time.Now()
	timeDuration := convertedTime.Sub(timeNow)
	timeToDeparture := fmt.Sprintf("%.0f Min (per schedule)", timeDuration.Minutes())
	return timeToDeparture, err
}

//Conversion function for MS datetime
func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

//Main function - take input and call supporting functions
func main() {
	//Define input flags and exit if not provided
	routeFlag := flag.String("route", "", fmt.Sprintf("Route description, e.g. %q", "Express - Target - Hwy 252 and 73rd Av P&R - Mpls"))
	stopFlag := flag.String("stop", "", fmt.Sprintf("Stop description, e.g. %q", "Target North Campus Building F"))
	directionFlag := flag.String("direction", "", fmt.Sprintf("Cardinal direction, e.g. %q", "north"))
	flag.Parse()

	if *routeFlag == "" || *stopFlag == "" || *directionFlag == "" {
		fmt.Println("Usage:  mtcnextbus -route=\"\" -stop=\"\" -direction=\"\"  (see mtcnextbus -help for more)")
		os.Exit(0)
	}

	var routeID string
	routeID, err := getRouteID(*routeFlag)
	//fmt.Println("routeID: ", routeID)
	if err != nil {
		fmt.Printf("error getting routeID: %s\n", err)
	}

	var directionID string
	if routeID != "" {
		directionID, err = getDirectionID(routeID, *directionFlag)
		//fmt.Println("directionID: ", directionID)
		if err != nil {
			fmt.Printf("error getting directionID: %s\n", err)
		}
	}

	var stopID string
	if directionID != "" {
		stopID, err = getStopID(routeID, directionID, *stopFlag)
		//fmt.Println("stopID: ", stopID)
		if err != nil {
			fmt.Printf("error getting stopID: %s\n", err)
		}
	}

	var nextDeparture string
	if stopID != "" {
		nextDeparture, err = getDeparture(routeID, directionID, stopID)
		if err != nil {
			fmt.Printf("error getting departure: %s\n", err)
		}
		if nextDeparture != "" {
			fmt.Printf("next departure: %s", nextDeparture)
		}
	}
}
