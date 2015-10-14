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
    "github.com/FactomProject/factoid/wallet"
    fct "github.com/FactomProject/factoid"
    "log"
    "os"
 )

 var chttp = http.NewServeMux()
 var myState IState
 var configDir string
 var staticDir string
 
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
        t, err := template.ParseFiles(staticDir + "fwallet.html")
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
		w.Write([]byte(fct.ConvertDecimal(uint64(v))))
 }
 
 func reqFee(w http.ResponseWriter, r *http.Request) {
        txKey := r.FormValue("key")
        
        ib := myState.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(txKey))
		trans, ok := ib.(fct.ITransaction)
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
			w.Write([]byte(strings.TrimSpace(fct.ConvertDecimal(fee))))
	    } else {
	        w.Write([]byte("..."))
	    }
			
 }

 func showFee(txKey string) []byte {
        
        ib := myState.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(txKey))
		trans, ok := ib.(fct.ITransaction)
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
			return []byte(strings.TrimSpace(fct.ConvertDecimal(fee)))
	    } else {
	        return []byte("...")
	    }
			
 }
 
 
 func craftTx(w http.ResponseWriter, r *http.Request) {
   		txKey := r.FormValue("key")
   		actionToDo := r.FormValue("action")

	    // Make sure we don't already have a transaction in process with this key
	    t := myState.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(txKey))
	    if t != nil {
            deleteErr := FactoidDeleteTx(txKey)
            if deleteErr != nil {
                w.Write([]byte(deleteErr.Error()))
                return
            }
	    }
	    // Create a transaction
	    t = myState.GetFS().GetWallet().CreateTransaction(myState.GetFS().GetTimeMilli())
	    // Save it with the key
	    myState.GetFS().GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(txKey), t)

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
            
            totalInputs := 0.0
            totalOutputs := 0.0
            
            for _, inputElement := range(inRes) {
                totalInputs += inputElement.InputSize
                inputFeedErr := SilentAddInput(string(txKey), string(inputElement.InputAddress), strconv.FormatFloat(inputElement.InputSize, 'f', -1, 64))
                if inputFeedErr != nil {
                    w.Write([]byte(inputFeedErr.Error() + " (INPUTS)"))
                    return
                }
                
                buffer.WriteString("\tInput: " + inputElement.InputAddress + " : " + strconv.FormatFloat(inputElement.InputSize, 'f', -1, 64) + "\n")
            }
            
            var outputFeedErr error
            
            for _, outputElement := range(outRes) {
                totalOutputs += outputElement.OutputSize
                if outputElement.OutputType == "fct" {
                    outputFeedErr = SilentAddOutput(string(txKey), string(outputElement.OutputAddress), strconv.FormatFloat(outputElement.OutputSize, 'f', -1, 64))
                } else {
                    outputFeedErr = SilentAddECOutput(string(txKey), string(outputElement.OutputAddress), strconv.FormatFloat(outputElement.OutputSize, 'f', -1, 64))
                }
                
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
                    
                    if _, err := os.Stat(fileToSaveTo); os.IsNotExist(err) {
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
                    } else {
                        w.Write([]byte(string(fileToSaveTo) + " already exists, please choose another filename to save to."))
                    }
                    return
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
 
 func SilentAddInput(txKey string, inputAddress string, inputSize string) error {
 	ib := myState.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(txKey))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction: " + txKey)
	}
		
	var addr fct.IAddress
	if !fct.ValidateFUserStr(inputAddress) {
		if len(inputAddress) != 64 {
			if len(inputAddress) > 32 {
				return fmt.Errorf("Invalid Name. Check the address or name for proper entry.", len(inputAddress))
			}

			we := myState.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(inputAddress))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				inputAddress = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			badr,err := hex.DecodeString(inputAddress)
			if err != nil {
				return fmt.Errorf("Looks like an Invalid hex address.  Check that you entered it correctly.")
			}
			addr = fct.NewAddress(badr)
		}
	} else {
		//fmt.Printf("adr: %x\n",adr)
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(inputAddress))
	}
	amount, _ := fct.ConvertFixedPoint(inputSize)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := myState.GetFS().GetWallet().AddInput(trans, addr, uint64(bamount))

	if err != nil {
		return err
	}
	return nil
 }

 func SilentAddOutput(txKey string, outputAddress string, outputSize string) error {
 	ib := myState.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(txKey))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}

	var addr fct.IAddress
	if !fct.ValidateFUserStr(outputAddress) {
		if len(outputAddress) != 64 {
			if len(outputAddress) > 32 {
				return fmt.Errorf("Invalid Address or Name.  Check that you entered it correctly.")
			}

			we := myState.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(outputAddress))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				outputAddress = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			if badHexChar.FindStringIndex(outputAddress) != nil {
				return fmt.Errorf("Looks like an invalid Hex Address.  Check that you entered it correctly.")
			}
		}
	} else {
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(outputAddress))
	}
	amount, _ := fct.ConvertFixedPoint(outputSize)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := myState.GetFS().GetWallet().AddOutput(trans, addr, uint64(bamount))
	if err != nil {
		return err
	}


	return nil
 }

 func SilentAddECOutput(txKey string, outputAddress string, outputSize string) error {

	ib := myState.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(txKey))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}

	var addr fct.IAddress
	if !fct.ValidateECUserStr(outputAddress) {
		if len(outputAddress) != 64 {
			if len(outputAddress) > 32 {
				return fmt.Errorf("Invalid Address or Name.  Check that you entered it correctly.")
			}

			we := myState.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(outputAddress))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				outputAddress = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			if badHexChar.FindStringIndex(outputAddress) != nil {
				return fmt.Errorf("Looks like an invalid hex address. Check that you entered it correctly.")
			}
		}
	} else {
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(outputAddress))
	}
	amount, _ := fct.ConvertFixedPoint(outputSize)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := myState.GetFS().GetWallet().AddECOutput(trans, addr, uint64(bamount))
	if err != nil {
		return err
	}

	return nil
}
 
 func FactoidDeleteTx(key string) error {
	// Make sure we have a key
	if len(key) == 0 {
		return fmt.Errorf("Missing transaction key")
	}
	// Wipe out the key
	myState.GetFS().GetDB().DeleteKey([]byte(fct.DB_BUILD_TRANS), []byte(key))
	return nil
 }

  func GetTransactions() ([][]byte, error) {
	    // Get the transactions in flight.
	    keys, values := myState.GetFS().GetDB().GetKeysValues([]byte(fct.DB_BUILD_TRANS))

	    for i := 0; i < len(keys)-1; i++ {
		    for j := 0; j < len(keys)-i-1; j++ {
			    if bytes.Compare(keys[j], keys[j+1]) > 0 {
				    t := keys[j]
				    keys[j] = keys[j+1]
				    keys[j+1] = t
				    t2 := values[j]
				    values[j] = values[j+1]
				    values[j+1] = t2
			    }
		    }
	    }
	    theKeys := [][]byte{}
	

	
	    for i, _ := range values {
		    if values[i] == nil {
			    continue
		    }
		    theKeys = append(theKeys,keys[i])
	    }

	    return theKeys, nil
  }
 
 
 func receiveAjax(w http.ResponseWriter, r *http.Request) {
 	if r.Method == "POST" {
 		ajax_post_data := r.FormValue("ajax_post_data")
 		call_type := r.FormValue("call_type")
 		switch call_type {
 		    case "balance":
 		        printBal, err := FctBalance(myState, ajax_post_data)
 		        check(err, false)
 		        w.Write([]byte("Factoid Address " + ajax_post_data + " Balance: " + strings.Trim(fct.ConvertDecimal(uint64(printBal)), " ") + " ⨎"))
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
 		        txNames, err := GetTransactions()
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
                    
                    ib := myState.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(txName))
                    jib, jerr := json.Marshal(ib)
                    var dat map[string]interface{}

                        if err := json.Unmarshal(jib, &dat); err != nil {
                            panic(err)
                        }
                        //fmt.Printf("%+v", dat)
                        if dat["Inputs"] != nil {
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
                                        myInps[i].InputAddress = fct.ConvertFctAddressToUserStr(fct.NewAddress(decodeAddr))
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
                                            myOuts[i].OutputAddress = fct.ConvertFctAddressToUserStr(fct.NewAddress(decodeAddr))
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
                                            myOuts[(i+len(outputObjects))].OutputAddress = fct.ConvertECAddressToUserStr(fct.NewAddress(decodeAddr))
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
                                
                            if jerr != nil {
                                w.Write([]byte(jerr.Error()))
                                return
                            }
             		        w.Write([]byte(lastTry))    
             		        
             		   }
        }
 	} else {
 	    helpText, err := ioutil.ReadFile(staticDir + "help.txt")
        check(err, false)
        w.Write([]byte(helpText))
 	}
 }

 func startServer(state IState, configDir string) {
 	// http.Handler
 	myState = state
 	staticDir = configDir + "static/"

 	chttp.Handle("/", http.FileServer(http.Dir(staticDir)))
 	mux := http.NewServeMux()
 	mux.HandleFunc("/", Home)
 	mux.HandleFunc("/receive", receiveAjax)
 	mux.HandleFunc("/rate", currRate)
 	mux.HandleFunc("/tx", craftTx)
 	mux.HandleFunc("/fee", reqFee)

 	http.ListenAndServe(":2337", mux)
 }
