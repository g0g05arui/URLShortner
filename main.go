package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/julienschmidt/httprouter"
)

type RedirectInfo struct {
	From string `json:"From"`
	To   string `json:"To"`
}

type HTTPResponse struct {
	Response []string `json:"Response"`
	Code     int      `json:"Code"`
}

type JsonHelper struct {
	Rd []RedirectInfo `json:"List"`
}

var redirects map[string]string = make(map[string]string)
var redirectList JsonHelper
var indexPage []byte

func main() {
	router := httprouter.New()
	router.GET("/", redirectHandler)
	router.GET("/:key", redirectHandler)
	router.POST("/api/add", addHandler)
	SetupCloseHandler()
	log.Fatal(http.ListenAndServe(":8080", router))
}

func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		onExit()
		os.Exit(0)
	}()
}

func onExit() {
	os.Remove("info.json")
	res, _ := json.Marshal(redirectList)
	ioutil.WriteFile("info.json", res, 0644)
}

func init() {
	buf, _ := ioutil.ReadFile("info.json")
	indexPage, _ = ioutil.ReadFile("index.html")
	json.Unmarshal(buf, &redirectList)
	fmt.Println(redirectList.Rd)
	for _, v := range redirectList.Rd {
		redirects[v.From] = v.To
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	query := p.ByName("key")
	newURL, ok := redirects[query]
	fmt.Println(newURL, ok)
	if ok == false {
		fmt.Fprintln(w, string(indexPage))
	} else {
		http.Redirect(w, r, newURL, 302)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var newLn RedirectInfo
	var response []byte
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&newLn)
	if _, ok := redirects[newLn.From]; ok {
		response, _ = json.Marshal(HTTPResponse{[]string{"Already Exists"}, 400})
	} else {
		redirectList.Rd = append(redirectList.Rd, newLn)
		redirects[newLn.From] = newLn.To
		response, _ = json.Marshal(HTTPResponse{[]string{"Added"}, 200})
	}
	fmt.Fprintln(w, string(response))
}
