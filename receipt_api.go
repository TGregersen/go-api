package main

import (
	"bytes"
	"encoding/json"
	"log"
	"fmt"
	"net/http"
	"strings"
    "github.com/google/uuid"
)

//Object to contain receipt JSON
type Receipt struct{
	Retailer string 'json:"retailer"'
	PurchaseDate string 'json:"purchaseDate"'
	PurchaseTime string 'json:"purchaseTime"'
	Total string 'json:"title"'
	Items []Item 'json:"items"'
}

//Object to contain item JSON
type Item struct{
	ShortDescription string 'json:"ShortDescription"'
	Price string 'json:"price"' 
}

type Id struct{
	Id string 'json:"id"'
}

type Points struct{
	Points string 'json:"points"'
}

func main() {

	var receiptPoints map[string]string

	receiptPoints = make(map[string]string)

	go server()
	time.Sleep(time.Second)

}

//handle post and get functions
func server(){
	if(*http.Request.Method == http.MethodPost){
		http.HandleFunc("/receipts/process", postProcess(http.ResponseWriter, *http.Request))

	} if (*http.Request.Method == http.MethodPost) {

		idInput := &Id{}		    
		err := json.NewDecoder(*http.Request.Body).Decode(id)
	    if err != nil {
	        http.Error(writer, err.Error(), http.StatusBadRequest)
	        return
	    }

		http.HandleFunc("/recipts/"+idInput.id+"/process", 
			getProcess(http.ResponseWriter, idInput))

	} else {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
        return
	}
}

func postProcess(writer http.ResponseWriter, response *http.Request){

    receipt := &Receipt{}
    err := json.NewDecoder(response.Body).Decode(receipt)
    if err != nil {
        http.Error(writer, err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Println("got receipt:", receipt)
    writer.WriteHeader(http.StatusCreated)

    //store uuid and points total
    uuid = genUuid()
	receiptPoints[uuid] = processReceipt(receipt); 

    //build uuid response
    id := &Id{
    	id: uuid
    }

    idJSON := new(bytes.Buffer)
    err := json.NewEncoder(idJSON).Encode(user)
    if err != nil {
        return err
    }

    if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
        panic(err)
    }
}

//Generate a random UUID to map to points total
func genUuid() {
    id := uuid.New()
    fmt.Println(id.String())
    return id.String()
}

//Create a JSON response and return to requestor with points count
func getProcess(writer http.ResponseWriter, response Id){

	totalPoints := &Points{
		points: receiptPoints[response.id]
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
}

// Handles receipt processing and tabulation of total for response.
//Place ID into memory
func processReceipt(Receipt receiptJson){

	int16 totalPoints = 0;

	string companyName = receiptJson(Data, &retailier);
	string date = receiptJson(Data, &purchaseDate);
	string time = receiptJson(Data, &purchaseTime);
	string purchaseTotal = receiptJson(Data, &total);

	totalPoints += processName(companyName);
	totalPoints += processDateTime(date, time);
	totalPoints += processTotal(purchaseTotal);

	totalPoints += processLineItems(receiptJson(Data,&items));

	return totalPoints;

}

//One point for each alphanumeric character in name.
func processName(companyName){
	int points = 0;
	string regexPattern = "[A-Za-z0-9]";

	for(char c : companyName){
		if(c.regexPattern.match()){
			points++;
		}
	}

	return points;
}

//6 points if the day in the purchase date is odd.
//10 points if the time of purchase is after 2:00pm and before 4:00pm.
func processDateTime(date, time){
	int points = 0;
	if(date.substring(5,7).toInt() % 2 == 0){
		points+=6;
	}
	if(time > 14:00 && time < 16:00){
		points+=10;
	}
	return points;
}

//50 points if the total is a round dollar amount with no cents.
//25 points if the total is a multiple of 0.25.
func processTotal(purchaseTotal){
	int points = 0;
	if(purchaseTotal.contains(".00")){
		points += 50;
	}
	if(purchaseTotal % .25 == 0){
		points += 25;
	}
	return points;
}

//5 points for every two items on the receipt.
//If the trimmed length of the item description is a multiple of 3, 
// multiply the price by 0.2 and round up to the nearest integer. 
// The result is the number of points earned.
func processLineItems(elem[]){
	int points = elem.length/2;

	for(e : elem){
		string desc = e.getitemDescription().trim();
		if(desc.length()%3==0){
			points += math.round(e.getItemPrice()*0.2);
		}
	}
	return points;
}