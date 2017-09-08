package main

import (
	"fmt"
	"time"
	"github.com/jasonlvhit/gocron"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"sort"
	"encoding/json"
	"os"
	"net/http"
	"io/ioutil"
)


func check(e error) {
	if e != nil {
		//log.Fatalf("%s:", e)
		panic(e)
	}
}

func downloadWOD() {

	doc, err := goquery.NewDocument("http://www.raedbox.eu/wod/")
	if err != nil {
		log.Fatal(err)
	}


	wods := make(map[string]string)

	keyString := ""
	doc.Find("#black-studio-tinymce-7 .textwidget p").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		wod := s.Text()

		if(len(wod) > 8){
			stringDate := wod[0:8]

			tmp := strings.Split(stringDate, ".")
			if(len(tmp) == 3){
				stringDate = tmp[1] + "-" + tmp[0] + "-" + "20" + tmp[2]  + " 10:00:00"
				t, _ := time.Parse("1-2-2006 15:04:05", stringDate)
				//fmt.Print(t.Day(), t.Month(), t.Year(), "\n")
				//fmt.Println(t.Format("2006-01-02"))

				keyString = t.Format("2006-01-02")

				wods[keyString] = wod
			} else {
				wods[keyString] =wods[keyString] + "\n\n" + wod
			}
			//fmt.Printf(stringDate)
		}
	})

	jsonPath := dataPath + "/json/wods.json"

	// delete file
	err = os.Remove(jsonPath)
	if err != nil {
		panic(err)
	}


	f, err := os.OpenFile(jsonPath, os.O_CREATE | os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	var keys []string
	for k := range wods {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// To perform the opertion you want
	for i, k := range keys {
		//fmt.Println("Key:", k, "Value:", wods[k])

		jsonString, _ := json.Marshal(wods[k])
		fmt.Println(string(jsonString))


		str := strings.Replace( wods[k], "'", "", -1)
		str = strings.Replace( str, `"`, `\"`, -1)
		str = strings.Replace( str, "\n", "\\n", -1)
		str = strings.Replace( str, "]", "", -1)
		str = strings.Replace( str, "[", "", -1)

		firstLine := "["
		if(i > 0){
			firstLine = ""
		}

		lastLine := ","

		if(i == len(keys) - 1){
			lastLine = "]"
		}

		if _, err = f.WriteString(firstLine + "{\"date\":\"" + k + "\"," + "\"wod\":\"" + str + "\"}" + lastLine); err != nil {
			panic(err)
		}
	}



	fmt.Println("i'm bored!!! --> please give me something to do...")
}

var configPath string

type WOD struct {
	Date	string `json:"date"`
	Wod 	string `json:"wod"`
}

func handler(w http.ResponseWriter, r *http.Request) {

	jsonPath := dataPath + "/json/wods.json"
	file, e := ioutil.ReadFile(jsonPath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}


	//m := new(Dispatch)
	//var m interface{}
	var jsontype []WOD
	json.Unmarshal(file, &jsontype)

	json.NewEncoder(w).Encode(jsontype)
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
var dataPath string


func main() {

	dataPath = "./data"
	if (os.Getenv("DOCKER") == "true") {
		dataPath = "/data"
	}

	println("dataPath: ", dataPath)

	downloadWOD();

	gocron.Every(1).Hour().Do(downloadWOD)

	// remove, clear and next_run
	_, NextRunTime := gocron.NextRun()
	fmt.Println("next run:", NextRunTime)

	t := time.Now()
	fmt.Println("Current Time:", t.Format("15:04:05.000"))


	http.HandleFunc("/", handler)
	http.ListenAndServe(":8081", nil)


	// function Start start all the pending jobs
	<-gocron.Start()


}
