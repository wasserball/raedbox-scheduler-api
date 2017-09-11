package main

import (
	"fmt"
	"time"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"sort"
	"encoding/json"
	"os"
	"net/http"
	"io/ioutil"
	"path/filepath"
	"io"
)


func check(e error) {
	if e != nil {
		log.Fatalf("%s:", e)
		panic(e)
	}
}

var dataPath string
var wodsPath string
var stundenplanPath string



func downloadStundenplan(){
	doc, err := goquery.NewDocument("http://www.raedbox.eu/stundenplan")
	check(err)

	doc.Find(".post-entry a").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		imgUrl, _ := s.Attr("href")
		var extension = filepath.Ext(imgUrl)
		if(len(extension) > 0 && strings.Contains(imgUrl, "Stundenplan")){
			println(imgUrl)
			response, err := http.Get(imgUrl)
			check(err)

			tmpImagePath := stundenplanPath + ".jpg"

			// save file
			file, err := os.Create(tmpImagePath)
			check(err)
			_, err = io.Copy(file, response.Body)
			file.Close()

			// rename
			err =  os.Rename(tmpImagePath, stundenplanPath)
			check(err)

			fmt.Println("Image saved!")

		}
	})
}

func downloadWOD() {
	println("start downloadWOD ...")

	doc, err := goquery.NewDocument("http://www.raedbox.eu/wod/")
	check(err)
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
				keyString = t.Format("2006-01-02")
				wods[keyString] = wod
			} else {
				wods[keyString] =wods[keyString] + "\n\n" + wod
			}
		}
	})

	t := time.Now()
	jsonPath := dataPath + "/json/"+t.Format("20060102150405")+"-wods.json"

	// delete file
	//err = os.Remove(jsonPath)
	//check(err)

	f, err := os.OpenFile(jsonPath, os.O_CREATE | os.O_WRONLY, 0600)
	check(err)

	defer f.Close()

	var keys []string
	for k := range wods {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// To perform the opertion you want
	for i, k := range keys {

		//jsonString, _ := json.Marshal(wods[k])
		//fmt.Println(string(jsonString))

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
		} else {

			file, err := ioutil.ReadFile(jsonPath)
			// file exists
			if err == nil {
				// test Unmarshal
				var jsonToSend []WOD
				err = json.Unmarshal(file, &jsonToSend)
				// array is OK and min one wod
				if err == nil && len(jsonToSend) > 0{
					println("jsonToSend length", len(jsonToSend))
					// rename file
					err =  os.Rename(jsonPath, wodsPath)
					check(err)
					println("file rename OK", jsonPath, wodsPath)

				}
			}
		}
	}
}

type WOD struct {
	Date	string `json:"date"`
	Wod 	string `json:"wod"`
}

func wods(w http.ResponseWriter, r *http.Request) {

	file, err := ioutil.ReadFile(wodsPath)
	check(err)

	var jsonToSend []WOD
	err = json.Unmarshal(file, &jsonToSend)

	if err != nil {
		emptyArray := make([]string, 0)
		println("json.Unmarshal error", err)
		json.NewEncoder(w).Encode(emptyArray)
	} else {
		json.NewEncoder(w).Encode(jsonToSend)
	}
}

func stundenplan(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, stundenplanPath)
}


func startPolling() {
	// run @ start
	downloadWOD()
	downloadStundenplan()
	for {
		// run every 1 Hours
		time.Sleep(1 * time.Hour)
		go downloadWOD()
		go downloadStundenplan()
	}
}

func main() {
	dataPath = "./data"
	if (os.Getenv("DOCKER") == "true") {
		dataPath = "/data"
	}
	wodsPath = dataPath + "/json/wods.json"

	stundenplanPath = dataPath + "/img/Stundenplan.jpg"

	println("dataPath: ", dataPath, "wodsPath:", wodsPath)

	go startPolling()

	t := time.Now()
	fmt.Println("Current Time:", t.Format("15:04:05.000"))

	http.HandleFunc("/raedbox/wods", wods)
	http.HandleFunc("/raedbox/stundenplan", stundenplan)
	http.ListenAndServe(":8081", nil)

}
