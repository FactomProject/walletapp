package main

 import (
 	"fmt"
 	"net/http"
 	"html/template"
 	"bytes"
    "strings"
    "strconv"
    "io/ioutil"
    "encoding/json"
    "encoding/hex"
    "github.com/FactomProject/fctwallet/Wallet"
    "github.com/FactomProject/factoid"
    "log"
 )

 var chttp = http.NewServeMux()
 var myState IState
 
 type inputList struct {
    InputSize float64 `json:"inputSize"`
    InputAddress string `json:"inputAddress"`
    }
    
 type outputList struct {
    OutputSize float64 `json:"outputSize"`
    OutputAddress string `json:"outputAddress"`
    OutputType string `json:"outputType"`
    }

 type pseudoTran struct {
		Inputs []inputList
		Outputs []outputList
	}
    
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
 
 func currRate(w http.ResponseWriter, r *http.Request) {
		v, err := GetRate(myState)
		if err != nil {
			fmt.Println(err)
            return
		}
		w.Write([]byte(factoid.ConvertDecimal(uint64(v))))
 }
 
 func reqFee(w http.ResponseWriter, r *http.Request) {
        txKey := r.FormValue("key")
        
        ib := myState.GetFS().GetDB().GetRaw([]byte(factoid.DB_BUILD_TRANS), []byte(txKey))
		trans, ok := ib.(factoid.ITransaction)
		if ib != nil && ok {
			
			
			v, err := GetRate(myState)
			if err != nil {
				fmt.Println(err)
				return
			}
			fee, err := trans.CalculateFee(uint64(v))
			if err != nil {
				fmt.Println(err)
				return
			}
			w.Write([]byte(strings.TrimSpace(factoid.ConvertDecimal(fee))))
	    } else {
	        w.Write([]byte("..."))
	    }
			
 }

 func showFee(txKey string) []byte {
        
        ib := myState.GetFS().GetDB().GetRaw([]byte(factoid.DB_BUILD_TRANS), []byte(txKey))
		trans, ok := ib.(factoid.ITransaction)
		if ib != nil && ok {
			
			
			v, err := GetRate(myState)
			if err != nil {
				fmt.Println(err)
	            return []byte("...")
			}
			fee, err := trans.CalculateFee(uint64(v))
			if err != nil {
				fmt.Println(err)
	            return []byte("...")
			}
			return []byte(strings.TrimSpace(factoid.ConvertDecimal(fee)))
	    } else {
	        return []byte("...")
	    }
			
 }
 
 
 func craftTx(w http.ResponseWriter, r *http.Request) {
   		txKey := r.FormValue("key")
   		actionToDo := r.FormValue("action")
   		
        execStrings := []string{"NewTransaction", txKey}
        newTXErr := myState.Execute(execStrings)
        if newTXErr != nil {
            deleteErr := FactoidDeleteTx(txKey)
            if deleteErr != nil {
                w.Write([]byte(deleteErr.Error()))
                return
            }
        } 
        
        
            myState.Execute(execStrings)

            var buffer bytes.Buffer
            buffer.WriteString("Transaction " + txKey + ":\n\n")
            
            inputStr := r.FormValue("inputs")
            var inRes []inputList
            err := json.Unmarshal([]byte(inputStr), &inRes)
	        if err != nil {
		        w.Write([]byte("Error: " + err.Error()))
		        return
	        }
	        
            
            outputStr := r.FormValue("outputs")
            var outRes []outputList
            json.Unmarshal([]byte(outputStr), &outRes)
            
            var inputFeed []string
            var outputFeed []string
            totalInputs := 0.0
            totalOutputs := 0.0
            
            for _, inputElement := range(inRes) {
                inputFeed = []string{"AddInput", string(txKey), string(inputElement.InputAddress), strconv.FormatFloat(inputElement.InputSize, 'f', -1, 64)}
                totalInputs += inputElement.InputSize
                inputFeedErr := myState.Execute(inputFeed)
                if inputFeedErr != nil {
                    w.Write([]byte(inputFeedErr.Error() + " (INPUTS)"))
                    return
                }
                
                buffer.WriteString("\tInput: " + inputElement.InputAddress + " : " + strconv.FormatFloat(inputElement.InputSize, 'f', -1, 64) + "\n")
            }
            
            /*printTest := []string{"Print", string(txKey)}   
                    printTestErr := myState.Execute(printTest)
                    if printTestErr != nil {
                        w.Write([]byte(printTestErr.Error()))
                    }    
            */
            
            for _, outputElement := range(outRes) {
                totalOutputs += outputElement.OutputSize
                if outputElement.OutputType == "fct" {
                    outputFeed = []string{"AddOutput", string(txKey), string(outputElement.OutputAddress), strconv.FormatFloat(outputElement.OutputSize, 'f', -1, 64)}
                } else {
                    outputFeed = []string{"AddECOutput", string(txKey), string(outputElement.OutputAddress), strconv.FormatFloat(outputElement.OutputSize, 'f', -1, 64)}
                }
                
                outputFeedErr := myState.Execute(outputFeed)
                if outputFeedErr != nil {
                    w.Write([]byte(outputFeedErr.Error() + " (OUTPUTS)"))
                    return
                }   
                
                buffer.WriteString("\tOutput: " + outputElement.OutputAddress + " : " + strconv.FormatFloat(outputElement.OutputSize, 'f', -1, 64) + " (" + outputElement.OutputType + 
                                   ") \n")
            }
      	    currFee := totalInputs - totalOutputs
      	    
      	    buffer.WriteString("\n\tFee: " + strconv.FormatFloat(currFee, 'f', -1, 64))

                                   
      	    switch actionToDo {
      	        case "fee":
                    w.Write(showFee(txKey))
                case "print":
                    printTest := []string{"Print", string(txKey)}   
                    printTestErr := myState.Execute(printTest)
                    if printTestErr != nil {
                        w.Write([]byte("PRINTERR: " + printTestErr.Error()))
                    }
                    w.Write(buffer.Bytes())
                case "save":
                    fileToSaveTo := r.FormValue("fileName")
                    if len(fileToSaveTo) < 1 {
                        w.Write([]byte("Filename cannot be empty!\n\n"))
                        return
                    }
                    
                    signFeedString := []string{"Sign", string(txKey)}    
                    signErr := myState.Execute(signFeedString)
                    if signErr != nil {
                        w.Write([]byte("SIGNERR: " + signErr.Error()))
                    }

                    saveFeedString := []string{"Export", string(txKey), string(fileToSaveTo)}    
                    saveErr := myState.Execute(saveFeedString)
                    if saveErr != nil {
                        fmt.Println(saveErr)
                    }
                    buffer.WriteString("\n\nTransaction ")
                    buffer.WriteString(txKey)
                    buffer.WriteString(" has been saved to file: ")
                    buffer.WriteString(string(fileToSaveTo))
                    w.Write(buffer.Bytes())
                case "sign":
                    signFeedString := []string{"Sign", string(txKey)}    
                    signErr := myState.Execute(signFeedString)
                    if signErr != nil {
                        w.Write([]byte(signErr.Error()))
                        return
                    }
                    
                    buffer.WriteString("\n\nTransaction ")
                    buffer.WriteString(txKey)
                    buffer.WriteString(" has been signed.")
                    w.Write(buffer.Bytes())
                case "send":
                    testPrintTx := []string{"Print", string(txKey)}   

                    printErr := myState.Execute(testPrintTx)
                    if printErr != nil {
                         w.Write([]byte(printErr.Error()))
                         return
                    }      
                    
                    signFeedString := []string{"Sign", string(txKey)}    
                    signErr := myState.Execute(signFeedString)
                    if signErr != nil {
                        w.Write([]byte(signErr.Error()))
                        return
                    }

                    submitFeedString := []string{"Submit", string(txKey)}    
                    submitErr := myState.Execute(submitFeedString)
                    if submitErr != nil {
                        w.Write([]byte(submitErr.Error()))
                        return
                    }
                       
                    buffer.WriteString("\n\nTransaction ")
                    buffer.WriteString(txKey)
                    buffer.WriteString(" has been submitted to Factom.")
                    w.Write(buffer.Bytes())
            }
 }
 
 func FactoidDeleteTx(key string) error {
	// Make sure we have a key
	if len(key) == 0 {
		return fmt.Errorf("Missing transaction key")
	}
	// Wipe out the key
	myState.GetFS().GetDB().DeleteKey([]byte(factoid.DB_BUILD_TRANS), []byte(key))
	return nil
 }
 
 
 func receiveAjax(w http.ResponseWriter, r *http.Request) {
 	if r.Method == "POST" {
 		ajax_post_data := r.FormValue("ajax_post_data")
 		call_type := r.FormValue("call_type")
 		switch call_type {
 		    case "balance":
 		        printBal, err := Wallet.FactoidBalance(ajax_post_data)
 		        check(err, false)
 		        w.Write([]byte("Factoid Address " + ajax_post_data + " Balance: " + strings.Trim(factoid.ConvertDecimal(uint64(printBal)), " ") + " ⨎"))
 		    case "balances":
 		        printBal := GetBalances(myState)
 		        testErr := myState.Execute([]string{"balances"})
 		        if testErr != nil {
                    fmt.Println(testErr.Error())
                    w.Write([]byte(testErr.Error()))
                    return
                }
 		        w.Write(printBal)
  		    case "allTxs":
 		        txNames, _, err := Wallet.GetTransactions()
 		        if err != nil {
 		            fmt.Println(err.Error())
 		            w.Write([]byte(err.Error()))
 		            return
 		        }
 		        if len(txNames) == 0 {
                    w.Write([]byte("No transactions to display."))
 		         	return
 		        }
 		        sliceTxNames := []byte("")
 		        for i:= range txNames {
 		            sliceTxNames = append(sliceTxNames, txNames[i]...)
 		            if i < len(txNames) - 1 {
 		                sliceTxNames = append(sliceTxNames, byte('\n'))
 		            }
 		        }
 		        w.Write(sliceTxNames)
 		    case "addNewAddress":
 		        if len(ajax_post_data) > 0 {
     		        genErr := GenAddress(myState, "fct", ajax_post_data)
     		        if genErr != nil {
 		                w.Write([]byte(genErr.Error()))
     		            return
     		        }
     		        w.Write([]byte(ajax_post_data + " has been added to your wallet successfully."));
                }
 		    case "addNewEC":
 		        if len(ajax_post_data) > 0 {
     		        genErr := GenAddress(myState, "ec", ajax_post_data)
     		        if genErr != nil {
 		                w.Write([]byte(genErr.Error()))
     		            return
     		        }
     		        w.Write([]byte(ajax_post_data + " has been added to your wallet successfully."));
     		    }
     		case "importPrivKey":
     		    addressName := r.FormValue("addressName")
 		        if len(ajax_post_data) > 0 && len(addressName) > 0 {

                    importFeedString := []string{"ImportKey", string(addressName), string(ajax_post_data)}    
                    importErr := myState.Execute(importFeedString)
                    if importErr != nil {
                        w.Write([]byte(importErr.Error()))
                        return
                    }

     		        w.Write([]byte("The contents of the private key have been added to " + addressName + " successfully!"));
     		    } else {
     		        w.Write([]byte("You must include a non-empty private key and name for the address to import it into."));
     		    }
     		case "loadTx":
     		    txName := r.FormValue("txName")
 		        if len(ajax_post_data) > 0 {
                    loadFeedString := []string{"Import", string(txName), string(ajax_post_data)}    
                    loadErr := myState.Execute(loadFeedString)
                    if loadErr != nil {
                        w.Write([]byte(loadErr.Error()))
                        return
                    }
                }
                    
                    ib := myState.GetFS().GetDB().GetRaw([]byte(factoid.DB_BUILD_TRANS), []byte(txName))
                    jib, jerr := json.Marshal(ib)
                    var dat map[string]interface{}

                        if err := json.Unmarshal(jib, &dat); err != nil {
                            panic(err)
                        }
                        fmt.Printf("%+v", dat)
                        
                        inputObjects := dat["Inputs"].([]interface{})
                        myInps := make([]inputList, len(inputObjects))
                        if len(inputObjects) > 0 {
                            currInput := inputObjects[0].(map[string]interface{})
                            for i := range(inputObjects) {
                                currInput = inputObjects[i].(map[string]interface{})
			                    decodeAddr, hexErr := hex.DecodeString(currInput["Address"].(string))
			                    if hexErr != nil {
			                        fmt.Println("Error: " + hexErr.Error())
			                        return
			                    }
                                myInps[i].InputAddress = factoid.ConvertFctAddressToUserStr(factoid.NewAddress(decodeAddr))
                                myInps[i].InputSize = currInput["Amount"].(float64)
                            }
                        }
                        loo := 0
                        loeco := 0
                        var outputObjects []interface{}
                        var outputECObjects []interface{}
                        if dat["Outputs"] != nil {
                            outputObjects = dat["Outputs"].([]interface{})
                            loo = len(outputObjects)
                        }
                        if dat["OutECs"] != nil {
                            outputECObjects = dat["OutECs"].([]interface{})
                            loeco = len(outputECObjects)
                        }
                        myOuts := make([]outputList, (loo + loeco))
                        if outputObjects != nil {
                            if loo > 0 {
                                currOutput := outputObjects[0].(map[string]interface{})
                                for i := range(outputObjects) {
                                    currOutput = outputObjects[i].(map[string]interface{})
			                        decodeAddr, hexErr := hex.DecodeString(currOutput["Address"].(string))
			                        if hexErr != nil {
			                            fmt.Println("Error: " + hexErr.Error())
			                            return
			                        }
                                    myOuts[i].OutputAddress = factoid.ConvertFctAddressToUserStr(factoid.NewAddress(decodeAddr))
                                    myOuts[i].OutputSize = currOutput["Amount"].(float64)
                                    myOuts[i].OutputType = "fct"
                                }
                            }
                        }
                        
                        if outputECObjects != nil {
                            if loeco > 0 {
                                currOutput := outputECObjects[0].(map[string]interface{})
                                for i := range(outputECObjects) {
                                    currOutput = outputECObjects[i].(map[string]interface{})
			                        decodeAddr, hexErr := hex.DecodeString(currOutput["Address"].(string))
			                        if hexErr != nil {
			                            fmt.Println("Error: " + hexErr.Error())
			                            return
			                        }
                                    myOuts[(i+len(outputObjects))].OutputAddress = factoid.ConvertECAddressToUserStr(factoid.NewAddress(decodeAddr))
                                    myOuts[(i+len(outputObjects))].OutputSize = currOutput["Amount"].(float64)
                                    myOuts[(i+len(outputObjects))].OutputType = "ec"
                                }
                            }
                        }
                        
                    returnTran := pseudoTran{
                        Inputs: myInps,
                        Outputs: myOuts,
                    }
                    
                    lastTry, jayErr := json.Marshal(returnTran)
                    if jayErr != nil {
                        w.Write([]byte(jerr.Error()))
                        return
                    }
     		        //w.Write([]byte(jib))    //"The contents of " + ajax_post_data + " have been added as transaction " + txName + " ."));
                        
                    if jerr != nil {
                        w.Write([]byte(jerr.Error()))
                        return
                    }
     		        w.Write([]byte(lastTry))    //"The contents of " + ajax_post_data + " have been added as transaction " + txName + " ."));
 		    /*
 		    case "addNewTx":
 		        execStrings := []string{"NewTransaction", ajax_post_data}
                newErr := myState.Execute(execStrings)
     		 	if newErr != nil {
     		 	    if newErr.Error()[:13] == "Duplicate key" {
     		 	        return //w.Write([]byte("Already have TX: " + ajax_post_data))
     		 	    }
     		 	    return
     		 	}
     		 	//Wallet.FactoidNewTransaction(ajax_post_data)	

             	w.Write([]byte(ajax_post_data))*/
        }
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
 	mux.HandleFunc("/rate", currRate)
 	mux.HandleFunc("/tx", craftTx)
 	mux.HandleFunc("/fee", reqFee)
 	
 	http.ListenAndServe(":2337", mux)
 }
