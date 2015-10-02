package main

 import (
 	"fmt"
 	"net/http"
 	"html/template"
    "strings"
    "io/ioutil"
    "github.com/FactomProject/fctwallet/Wallet"
    "github.com/FactomProject/factoid"
    "log"
 )

 var chttp = http.NewServeMux()
 var myState IState
 
 func check(e error, shouldEnd bool) {
    if e != nil {
        if shouldEnd {
            log.Fatal("Produced error: ", e)
        } else {
     	    log.Print("Produced error: ", e)
     	}
    }
 }


 func Home(w http.ResponseWriter, r *http.Request) {
 
    if (strings.Contains(r.URL.Path, ".")) {
        chttp.ServeHTTP(w, r)
    } else {
        t, err := template.ParseFiles("fwallet.html")
        if err != nil {
            fmt.Println("err: ", err)
        }
        t.Execute(w, nil)
    }

 }
 
 
 func receiveAjax(w http.ResponseWriter, r *http.Request) {
 	if r.Method == "POST" {
 		ajax_post_data := r.FormValue("ajax_post_data")
 		call_type := r.FormValue("call_type")
 		//fmt.Println(ajax_post_data)
 		switch call_type {
 		    case "balance":
 		        printBal, err := Wallet.FactoidBalance(ajax_post_data)
 		        check(err, false)
 		        w.Write([]byte("Factoid Address " + ajax_post_data + " Balance: " + strings.Trim(factoid.ConvertDecimal(uint64(printBal)), " ") + " ƒ"))
 		    case "balances":
 		        printBal := GetBalances(myState)
 		        w.Write(printBal)
 		    case "addNewTx":
     		 	err := Wallet.FactoidNewTransaction(ajax_post_data)
     		 	if err != nil {
     		 	    if err.Error()[:13] == "Duplicate key" {
     		 	        w.Write([]byte("Already have TX: " + ajax_post_data))
     		 	    }
     		 	    return
     		 	}
     		 	w.Write([]byte("Created tx " + ajax_post_data))
     	}
 		//©
 	} else {
 	    helpText, err := ioutil.ReadFile("./extra/help.txt")
        check(err, false)
        w.Write([]byte(helpText))
 	}
 }

 func startServer(state IState) {
 	// http.Handler
 	myState = state
 	chttp.Handle("/", http.FileServer(http.Dir("./extra/")))
 	mux := http.NewServeMux()
 	mux.HandleFunc("/", Home)
 	mux.HandleFunc("/receive", receiveAjax)

 	http.ListenAndServe(":2337", mux)
 }
