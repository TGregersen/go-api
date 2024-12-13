package main

import (
	"bytes"
	"fmt"
	"encoding/json"
	"net/http"
	"strings"
	"regexp"
	"log"
	"math"
	"strconv"
    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "time"
    "errors"
)

// Object to contain receipt JSON
type Receipt struct {
	Retailer string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total string `json:"total"`
	Items []Item `json:"items"`
}

// Object to contain item JSON
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price string `json:"price"`
}

// Object to contain Id Response JSON
type Id struct {
	Id string `json:"id"`
}

// Object to contain Points Response JSON
type Points struct {
	Points int `json:"points"`
}

// Driver
func main() {
	fmt.Println("Starting Application...")
	server()
	time.Sleep(time.Second)

}

// Server call to handle router for post and get functions
func server() {
	var addressAndPort string = ":9000"

	var receiptPoints map[string]int
	receiptPoints = make(map[string]int)

	r := mux.NewRouter().StrictSlash(true)

	idregex := "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	getPath := "/receipts/{id:" + idregex + "}/process"
	r.HandleFunc(getPath, func(wrt http.ResponseWriter, req *http.Request){
		params := mux.Vars(req)["id"]
		pointsResp, err := getProcess(receiptPoints, wrt, params)
		if err != nil {
	    	log.Println("Issue with Get Process.")
	        http.Error(wrt, err.Error(), http.StatusBadRequest)
	    }
		json.NewEncoder(wrt).Encode(pointsResp)
	})

	postPath := "/receipts/process"
	r.HandleFunc(postPath, func(wrt http.ResponseWriter, req *http.Request){
			idResp, err := postProcess(receiptPoints, wrt , req )
			if err != nil {
		    	log.Println("Issue with Post Process.")
		        http.Error(wrt, err.Error(), http.StatusBadRequest)
		    }
			json.NewEncoder(wrt).Encode(idResp)
		})

	http.Handle("/", r)

	fmt.Println("Listening on: ",addressAndPort)

	err := http.ListenAndServe(addressAndPort, r)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}

}

// Function to handle the POST call and creates and ID response for 
// requester
func postProcess(receiptPoints map[string]int, writer http.ResponseWriter, 
	request *http.Request) (*Id, error) {
	
	idresp := &Id{ }
    receipt := &Receipt{}
    err := json.NewDecoder(request.Body).Decode(receipt)
    
    if err != nil {
    	log.Println("The receipt is invalid.")
        http.Error(writer, err.Error(), http.StatusBadRequest)
        return idresp, err
    }

    if !verifyReceipt(receipt){
    	log.Println("The receipt is invalid.")
        http.Error(writer, err.Error(), http.StatusBadRequest)
        return idresp, err
    }

    fmt.Println("Processing receipt")
    writer.WriteHeader(http.StatusCreated)

    //store uuid and points total
    uuid := genUuid()
	receiptPoints[uuid] = processReceipt(receipt); 
 	
    //build uuid response
    idresp.Id = uuid

    idJSON := new(bytes.Buffer)
    err = json.NewEncoder(idJSON).Encode(idresp)
    if err != nil {
        return idresp, err
    }

    return idresp, err
}

// Verify Receipt to ensure it is valid.
func verifyReceipt(receiptJson *Receipt) bool {
	fmt.Println("Verifying Receipt")

	const companyPattern := "^[\\w\\s\\-&]+$"
	const timePattern = "15:04"
	const datePattern := "2024-12-08"
	const totalPattern := "^\\d+\\.\\d{2}$"

	var companyName string = receiptJson.Retailer
	var date string = receiptJson.PurchaseDate
	var time string = receiptJson.PurchaseTime
	var purchaseTotal string = receiptJson.Total


	res1, err := regexp.MatchString(companyPattern, companyName)
	if err != nil {
	    fmt.Println(err)
	    return false
	}

	dateP, err := time.Parse(datePattern, date)
	if err != nil {
	    fmt.Println(err)
	    return false
	}

	timeP, err := time.Parse(timePattern, time)
	if err != nil {
	    fmt.Println(err)
	    return false
	}

	totalP, err := time.Parse(totalPattern, purchaseTotal)
	if err != nil {
	    fmt.Println(err)
	    return false
	}

	if !verifyLineItems(receiptJson.Items) {
	    return false
	}

	return true
}

// Verify each line item to ensure they are valid.
func verifyLineItems(elem []Item) bool {
	fmt.Println("Verifying Line Items")

	descriptionPattern := "^[\\w\\s\\-]+$"
	totalPattern := "^\\d+\\.\\d{2}$"

	for _, e := range elem{

		res1, err :=regexp.MatchString(descriptionPattern, e.ShortDescription)
		if err != nil {
		    fmt.Println(err)
		    return false
		}

		res2, err :=regexp.MatchString(totalPattern, e.Price)
		if err != nil {
		    fmt.Println(err)
		    return false
		}
			
	}
	return true;
}

// Generate a random UUID to map to points total
func genUuid() string {
    id := uuid.New()
    fmt.Println(id.String())
    return id.String()
}

// Handles the GET call and creates a JSON response and 
// return to requestor with points count
func getProcess(receiptPoints map[string]int, 
	writer http.ResponseWriter, id string) (*Points, error){

	response := &Id{
		Id: id,
	}

	totalPoints := &Points{
		Points: receiptPoints[response.Id],
	}

	val, exists := receiptPoints[response.Id]
	if !exists {
		log.Println("No receipt found for that ID.")
        http.Error(writer, err.Error(), http.StatusBadRequest)
        return totalPoints, err
	}

    writer.WriteHeader(http.StatusCreated)

    totalPointsResponse := new(bytes.Buffer)
    err := json.NewEncoder(totalPointsResponse).Encode(totalPoints)
    if err != nil {
        return totalPoints, err
    }

    return totalPoints ,err
}

// Handles receipt processing and tabulation of total for response.
// Place ID into memory
func processReceipt(receiptJson *Receipt) int {
	var totalPoints int = 0;

	var companyName string = receiptJson.Retailer
	var date string = receiptJson.PurchaseDate
	var time string = receiptJson.PurchaseTime
	var purchaseTotal string = receiptJson.Total

	totalPoints += processName(companyName)

	if(date != "" && time != ""){
		totalPoints += processDateTime(date, time)
	}

	if(purchaseTotal != ""){
		totalPoints += processTotal(purchaseTotal)
	}

	totalPoints += processLineItems(receiptJson.Items)

	return totalPoints

}

// Tabulates points for Name 
// One point for each alphanumeric character in name.
func processName(companyName string) int{
	fmt.Println("Processing Company Name")

	var points int = 0
	var regexPattern string = "[A-Za-z0-9]"
	pattern := regexp.MustCompile(regexPattern)
	matches := pattern.FindAllString(companyName, -1)

	points += len(matches)

	return points
}

// Tabulates points for Time and Date
// 6 points if the day in the purchase date is odd.
// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
func processDateTime(dt string, tm string) int{
	fmt.Println("Processing Date and Time")

	var points int = 0;
	const layout = "15:04"

	timeParsed, err := time.Parse(layout, tm)
	if err != nil {
		fmt.Println("Time Parse Error:", err)
	} 

	lowestTime, err := time.Parse(layout,"14:00")
	if err != nil {
		fmt.Println("Time Parse Error:", err)
	} 
	
	highestTime, err := time.Parse(layout, "16:00")
	if err != nil {
		fmt.Println("Time Parse Error:", err)
	} 
	
	numDay, err := strconv.Atoi(dt[4:7])
	if err != nil {
		fmt.Println("Time Parse Error:", err)
	} 

	if(numDay % 2 != 0){

		points += 6;
		fmt.Println("Adding six points:", points)
	}

	if(timeParsed.After(lowestTime) && timeParsed.Before(highestTime)){
		points += 10;
		fmt.Println("Adding ten points:", points)
	}

	return points;
}

// Tabulates points for Total cost
// 50 points if the total is a round dollar amount with no cents.
// 25 points if the total is a multiple of 0.25.
func processTotal(purchaseTotal string) int{
	fmt.Println("Processing Total")

	var points int = 0

	if(strings.Contains(purchaseTotal, (".00"))){
		points += 50
	}

	fltTotal, err := strconv.ParseFloat(purchaseTotal, 64) 

	if err != nil {
		fmt.Println("Float Conversion Error:", err)
	} 

	if( math.Mod(fltTotal, .25 ) == 0){
		points += 25
	}

	return points;
}

// Tabulates points for each line item
// 5 points for every two items on the receipt.
// If the trimmed length of the item description is a multiple of 3, 
// multiply the price by 0.2 and round up to the nearest integer. 
// The result is the number of points earned.
func processLineItems(elem []Item) int{
	fmt.Println("Processing Line Items")

	var points int = (len(elem)/2)*5;

	for _, e := range elem{

		var desc string = strings.TrimSpace(e.ShortDescription)

		if(len(desc)%3 == 0){
			num, err := strconv.ParseFloat(e.Price, 64) 

			if err != nil {
				// Handle the error
				fmt.Println("Error:", err)
			} else {
				increased := num * 0.2
				points += int(math.Ceil(increased))
			}
			
		}
	}
	return points
}