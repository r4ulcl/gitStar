package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
    "github.com/microcosm-cc/bluemonday"

	"gopkg.in/russross/blackfriday.v2"
)

// Owner repository owner
type Owner struct {
	Login string
}

// Item repository struct
type Item struct {
	ID              int
	FullName        string `json:"full_name"`
	Description     string
	StargazersCount int    `json:"stargazers_count"`
	HTMLURL         string `json:"html_url"`
	Language        string
}

// JSONData contains array
type JSONData struct {
	Array []Item
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		log.Println("USAGE: ", os.Args[0], "username")
		log.Fatal("Parametros incorrectos")
	}
	user := os.Args[1]

	folder := "output"
	var arrAll []Item

	starred := 1 //for loop
	page := 1
	for starred != 0 {
		fmt.Println("https://api.github.com/users/" + user + "/starred?page=" + strconv.Itoa(page))
		res, err := http.Get("https://api.github.com/users/" + user + "/starred?page=" + strconv.Itoa(page))
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		if res.StatusCode == 403 {
			log.Fatal("API rate limit exceeded ", res.StatusCode)
		}
		if res.StatusCode != http.StatusOK {
			log.Fatal("Unexpected status code ", res.StatusCode)
		}
		arr := JSONData{}
		err = json.Unmarshal([]byte(body), &arr.Array)
		if err != nil {
			log.Printf("Unmarshaled: %v, error: %v", arr.Array, err)
		}

		arrAll = append(arrAll, arr.Array...) //join all body (all pages)
		starred = len(arr.Array)
		page++
	}

	toc := processTOC(arrAll)
	data := processData(arrAll)

	//fmt.Println(toc)
	//fmt.Println(data)

	print2MD(folder, user, toc+data)
	print2HTLM(folder, user, toc+data)

}

func getFile(repo string, file string) string {

	//fmt.Println("https://api.github.com/repos/" + repo + "/readme/")
	//res, err := http.Get("https://api.github.com/repos/" + repo + "/readme/") //using api, but 60 max request per hour
	fmt.Println("https://raw.githubusercontent.com/" + repo + "/master/"+file)
	res, err := http.Get("https://raw.githubusercontent.com/" + repo + "/master/"+file)

	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Println("Unexpected status code", res.StatusCode)
		return ""
	}

	return string(body)
}


func processTOC(data []Item) string {
	output := ""
	output += ("# Table of Contents\n\n")
	for e, i := range data {
		output += (strconv.Itoa(e+1) + ". [" + i.FullName + "](#" + i.FullName + ")\n\n")
	}

	return output
}

func processData(data []Item) string {
	output := ""
	for _, i := range data {
		output += ("\n\n# " + i.FullName)

		output += ("\n\n**Link**: ")
		output += ("[" + i.HTMLURL + "](" + i.HTMLURL + ")")

		//output += ("\n\n**Owner**: ")
		//output += (i.Owner.Login)

		output += ("\n\n**Language**: ")
		output += (i.Language)

		output += ("\n\n**StargazersCount**: ")
		output += strconv.Itoa(i.StargazersCount)

		output += ("\n\n**Description**: ")
		output += (i.Description)

		
		readme := getFile(i.FullName, "README.md")
		if readme == ""{
			readme = getFile(i.FullName, "README")
			if readme == ""{
				readme = getFile(i.FullName, "readme.md")
			}
		}
		if readme!=""{
			output += ("\n\n**README.md**: \n")
			output += ("\n-------------------\n")
			output += readme
			output += ("\n-------------------\n")
		}
		
	}

	return output
}

func print2MD(folder string, name string, text string) error {
	t := time.Now()
	f, err := os.Create(folder + "/" + name + "_" + t.Format("01-02-06") + "_Star.md")
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.WriteString(text)
	if err != nil {
		return err
	}
	return nil
}

func print2HTLM(folder string, name string, text string) error {
	//html
	md := []byte(text)
	//html := bf.Run(md)
	unsafe := blackfriday.Run(md)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	//fmt.Println(string(html))

	htmlString := "<html>" + string(html) + "</html>"

	//aux

	t := time.Now()
	f, err := os.Create(folder + "/" + name + "_" + t.Format("01-02-06") + "_Star.html")
	defer f.Close()
	if err != nil {
		fmt.Println(err)

		return err
	}
	_, err = f.WriteString(htmlString)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

/*
TODO
- file md to pdf
*/
