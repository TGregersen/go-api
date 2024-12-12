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
)

//Object to contain receipt JSON
type Receipt struct{
	Retailer string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total string `json:"title"`
	Items []Item `json:"items"`
}

//Object to contain item JSON
type Item struct{
	ShortDescription string `json:"ShortDescription"`
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
	go server()
	time.Sleep(time.Second)

}

//handle post and get functions
func server(){

	var receiptPoints map[string]int
	receiptPoints = make(map[string]int)

	r := mux.NewRouter()

	r.HandleFunc("/receipts/process", func(w http.ResponseWriter, r *http.Request){
			postProcess(receiptPoints, w , r )
		})

	r.HandleFunc("/recipts/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/process", 
		func(wrt http.ResponseWriter, req *http.Request){
			params := mux.Vars(req)["id"]
			getProcess(receiptPoints, wrt, params)
		})

	http.Handle("/", r)

	fmt.Println("Listening on :9001...")
	log.Print("Listen")

	log.Println(http.ListenAndServe(":9000" , nil))

}

func postProcess(receiptPoints map[string]int, writer http.ResponseWriter, request *http.Request) (*Id, error){
	idresp := &Id{ }
    receipt := &Receipt{}
    err := json.NewDecoder(request.Body).Decode(receipt)
    if err != nil {
        http.Error(writer, err.Error(), http.StatusBadRequest)
        return idresp, err
    }

    fmt.Println("got receipt:", receipt)
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

    if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
        panic(err)
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
func getProcess(receiptPoints map[string]int, writer http.ResponseWriter, id string) error{

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
        return err
    }

    if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
        panic(err)
    }

    return err
}

// Handles receipt processing and tabulation of total for response.
//Place ID into memory
func processReceipt(receiptJson *Receipt) int{

	var totalPoints int = 0;

	var companyName string = receiptJson.Retailer;
	var date string = receiptJson.PurchaseDate;
	var time string = receiptJson.PurchaseTime;
	var purchaseTotal string = receiptJson.Total;

	totalPoints += processName(companyName);
	totalPoints += processDateTime(date, time);
	totalPoints += processTotal(purchaseTotal);

	totalPoints += processLineItems(receiptJson.Items);

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
	const layout = "03:04"

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
	
	numDay, err := strconv.Atoi(dt[5:7])
	if err != nil {
		fmt.Println("Time Parse Error:", err)
	} 
	if(numDay % 2 == 0){
		points+=6;
	}

	if(timeParsed.After(lowestTime) && timeParsed.Before(highestTime)){
		points+=10;
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
	var points int = len(elem)/2;

	for _, e := range elem{
		var desc string = strings.TrimSpace(e.ShortDescription)
		
		if(len(desc)%3 == 0){
			num, err := strconv.ParseFloat(e.Price, 64) 

			if err != nil {
				// Handle the error
				fmt.Println("Error:", err)
			} else {
				increased := num * 0.2
				points += int(increased+0.5)
			}
			
		}
	}
	return points;
}