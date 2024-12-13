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

//Object to contain receipt JSON
type Receipt struct{
	Retailer string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total string `json:"total"`
	Items []Item `json:"items"`
}

//Object to contain item JSON
type Item struct{
	ShortDescription string `json:"shortDescription"`
	Price string `json:"price"`
}

type Id struct{
	Id string `json:"id"`
}

type Points struct{
	Points int `json:"points"`
}

func main() {
	fmt.Println("Starting Application...")
	server()
	time.Sleep(time.Second)

}

//handle post and get functions
func server(){
	var addressAndPort string = ":9000"

	var receiptPoints map[string]int
	receiptPoints = make(map[string]int)

	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/receipts/{id}/process", func(wrt http.ResponseWriter, req *http.Request){
		params := mux.Vars(req)["id"]
		fmt.Println("params..", params)
		pointsResp, err := getProcess(receiptPoints, wrt, params)
		if err != nil {
	    	log.Println("Issue with Get Process")
	        http.Error(wrt, err.Error(), http.StatusBadRequest)
	    }
		json.NewEncoder(wrt).Encode(pointsResp)
	})

	r.HandleFunc("/receipts/process", func(wrt http.ResponseWriter, req *http.Request){
			idResp, err := postProcess(receiptPoints, wrt , req )
			if err != nil {
		    	log.Println("Issue with Post Process")
		        http.Error(wrt, err.Error(), http.StatusBadRequest)
		    }
			json.NewEncoder(wrt).Encode(idResp)
		})

	http.Handle("/", r)

	fmt.Println("Listening on ",addressAndPort)
	log.Println("Listen")

	err := http.ListenAndServe(addressAndPort, r)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}

}

func postProcess(receiptPoints map[string]int, writer http.ResponseWriter, request *http.Request) (*Id, error){
	idresp := &Id{ }
    receipt := &Receipt{}
    err := json.NewDecoder(request.Body).Decode(receipt)
    log.Println("Decoded receipt")
    if err != nil {
    	log.Println("Issue with receipt")
        http.Error(writer, err.Error(), http.StatusBadRequest)
        return idresp, err
    }

    fmt.Println("got receipt:", receipt)
    writer.WriteHeader(http.StatusCreated)

    //store uuid and points total
    uuid := genUuid()
	receiptPoints[uuid] = processReceipt(receipt); 
 	
 	fmt.Println("Recipt ID, Points:", receiptPoints)

    //build uuid response
    idresp.Id = uuid

    idJSON := new(bytes.Buffer)
    err = json.NewEncoder(idJSON).Encode(idresp)
    if err != nil {
        return idresp, err
    }

    return idresp, err
}

//Generate a random UUID to map to points total
func genUuid() string {
    id := uuid.New()
    fmt.Println(id.String())
    return id.String()
}

//Create a JSON response and return to requestor with points count
func getProcess(receiptPoints map[string]int, writer http.ResponseWriter, id string) (*Points, error){

	response := &Id{
		Id: id,
	}

	totalPoints := &Points{
		Points: receiptPoints[response.Id],
	}

    fmt.Println("got points:", totalPoints)
    writer.WriteHeader(http.StatusCreated)

    totalPointsResponse := new(bytes.Buffer)
    err := json.NewEncoder(totalPointsResponse).Encode(totalPoints)
    if err != nil {
        return totalPoints, err
    }

    return totalPoints ,err
}

// Handles receipt processing and tabulation of total for response.
//Place ID into memory
func processReceipt(receiptJson *Receipt) int{
	fmt.Println("Processing Receipt")

	var totalPoints int = 0;

	var companyName string = receiptJson.Retailer;
	var date string = receiptJson.PurchaseDate;
	var time string = receiptJson.PurchaseTime;
	var purchaseTotal string = receiptJson.Total;

	totalPoints += processName(companyName);
	fmt.Println("Processing Receipt:",totalPoints)
	if(date != "" && time != ""){
		fmt.Println("Time and Date", date, time)
		totalPoints += processDateTime(date, time);
	}
	fmt.Println("Processing Receipt:",totalPoints)
	if(purchaseTotal != ""){
		fmt.Println("Purchase Total", purchaseTotal)
		totalPoints += processTotal(purchaseTotal);
	}
	fmt.Println("Processing Receipt:",totalPoints)

	totalPoints += processLineItems(receiptJson.Items);
	fmt.Println("Processing Receipt:",totalPoints)

	return totalPoints;

}

//One point for each alphanumeric character in name.
func processName(companyName string) int{
	var points int = 0;
	var regexPattern string = "[A-Za-z0-9]";
	pattern := regexp.MustCompile(regexPattern)
	matches := pattern.FindAllString(companyName, -1)

	points = points+len(matches)

	return points;
}

//6 points if the day in the purchase date is odd.
//10 points if the time of purchase is after 2:00pm and before 4:00pm.
func processDateTime(dt string, tm string) int{
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

	fmt.Println("Date time:", numDay, timeParsed)

	if(numDay % 2 != 0){

		points=points+6;
		fmt.Println("Adding six points:", points)
	}

	if(timeParsed.After(lowestTime) && timeParsed.Before(highestTime)){
		points=points+10;
		fmt.Println("Adding ten points:", points)
	}

	return points;
}

//50 points if the total is a round dollar amount with no cents.
//25 points if the total is a multiple of 0.25.
func processTotal(purchaseTotal string) int{
	var points int = 0;

	if(strings.Contains(purchaseTotal, (".00"))){
		points += 50;
	}

	fltTotal, err := strconv.ParseFloat(purchaseTotal, 64) 

	if err != nil {
		fmt.Println("Float Conversion Error:", err)
	} 

	if( math.Mod(fltTotal, .25 ) == 0){
		points += 25;
	}

	return points;
}

//5 points for every two items on the receipt.
//If the trimmed length of the item description is a multiple of 3, 
// multiply the price by 0.2 and round up to the nearest integer. 
// The result is the number of points earned.
func processLineItems(elem []Item) int{
	fmt.Println("Processing Line Items:",elem)
	var points int = (len(elem)/2)*5;
	fmt.Println("Processing Line Points:",points)
	for _, e := range elem{
		fmt.Println("Processing Line Item:",e.ShortDescription, e.Price)
		var desc string = strings.TrimSpace(e.ShortDescription)
		fmt.Println("Processing Line desc:",len(desc))
		if(len(desc)%3 == 0){
			num, err := strconv.ParseFloat(e.Price, 64) 

			if err != nil {
				// Handle the error
				fmt.Println("Error:", err)
			} else {
				increased := num * 0.2
				points += int(math.Ceil(increased))
				fmt.Println("Processing Line total:",points)
			}
			
		}
	}
	fmt.Println("Processing Line Items:",points)
	return points;
}