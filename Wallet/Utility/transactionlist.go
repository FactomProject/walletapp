// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package Utility


import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factoid/block"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/FactomCode/common"
	"encoding/json"
)

/************************************************
 * Transaction listing code
 ***********************************************/

// Older blocks smaller indexes.  All the Factoid Directory blocks
var DirectoryBlocks  = make([]*common.DirectoryBlock,0,100)
var FactoidBlocks    = make([]block.IFBlock,0,100)
var DBHead    []byte = common.ZERO_HASH
var DBHeadStr string = ""
var DBHeadLast []byte = common.ZERO_HASH	
	
// Refresh the Directory Block Head.  If it has changed, return true.
// Otherwise return false.
func getDBHead() bool {
	db, err := factom.GetDBlockHead()
	
	if err != nil {
		panic(err.Error())
	}
	
	if db.KeyMR != DBHeadStr {
		DBHeadStr = db.KeyMR
		DBHead,err = hex.DecodeString(db.KeyMR)
		if err != nil {
			panic(err.Error())
		}
		
		return true
	}
	return false
}

func getAll() error {
	dbs := make([] *common.DirectoryBlock,0,100)
	next := DBHeadStr
	
	for {
		blk,err := factom.GetRaw(next)
		if err != nil {
			panic(err.Error())
		}
		db := new(common.DirectoryBlock)
		err = db.UnmarshalBinary(blk)
		if err != nil {
			panic(err.Error())
		}
		dbs = append(dbs,db)
		if bytes.Equal(db.Header.PrevKeyMR.Bytes(),DBHeadLast) {
			break
		}
		next = hex.EncodeToString(db.Header.PrevKeyMR.Bytes())
	}
	
	DBHeadLast = DBHead
		
	for i:= len(dbs)-1;i>=0; i-- {
		DirectoryBlocks = append(DirectoryBlocks,dbs[i])
		fb := new(block.FBlock)
		var fcnt int
		for _,dbe := range dbs[i].DBEntries {
			if bytes.Equal(dbe.ChainID.Bytes(),common.FACTOID_CHAINID) {
				fcnt++
				hashstr := hex.EncodeToString(dbe.KeyMR.Bytes())
				fdata,err := factom.GetRaw(hashstr)
				if err != nil {
					panic(err.Error())
				}
				err = fb.UnmarshalBinary(fdata)
				if err != nil {
					panic(err.Error())
				}
				FactoidBlocks = append(FactoidBlocks,fb)
				break
			}
		}
		if fb == nil {
			panic("Missing Factoid Block from a directory block")
		}
		if fcnt > 1 {
			panic("More than one Factom Block found in a directory block.")
		}
		if err := ProcessFB(fb); err != nil {
			return err
		}
	}
	return nil
}

func refresh() error {

	if getDBHead() {
		if err := getAll(); err != nil {
			return err
		}
	}
	return nil
}

func filtertransaction(trans fct.ITransaction, addresses [][]byte) bool {
	if addresses == nil || len(addresses)==0 {
		return true
	}
	if len(trans.GetInputs()) == 0 &&
	   len(trans.GetOutputs())== 0 { 
		   return false
	}
	
	for _,adr := range addresses {
		for _,in := range trans.GetInputs() {
			if bytes.Equal(adr,in.GetAddress().Bytes()) {
				return true
			}
		}
		for _,out := range trans.GetOutputs() {
			if bytes.Equal(adr,out.GetAddress().Bytes()) {
				return true
			}
		}
		for _,ec := range trans.GetECOutputs() {
			if bytes.Equal(adr,ec.GetAddress().Bytes()) {
				return true
			}
		}
	}
	return false
}

func DumpTransactionsJSON(addresses [][]byte) ([]byte, error) {
	if err := refresh(); err != nil {
		return nil, err
	}

	var transactions []fct.ITransaction
	
	for i,fb := range FactoidBlocks {
		for _, t := range fb.GetTransactions() {
			t.SetBlockHeight(i)
			t.GetSigHash()
			prtTrans := filtertransaction(t,addresses)
			if prtTrans {
				transactions = append(transactions, t)
			}
		}
	}
	
	ret,err := json.Marshal(transactions)
	
	return ret,err
}


func DumpTransactions(addresses [][]byte) ([]byte, error) {
	var ret bytes.Buffer
	if err := refresh(); err != nil {
		return nil, err
	}
	usertranscnt := 0
	firstemptyblock := 0
	coinbasetranscnt := 0
	skippedblk := false
	
	for i,fb := range FactoidBlocks {
		var out bytes.Buffer
		
		blkempty := true
		out.WriteString(fmt.Sprintf("Block Height %d total transactions %d\n",i,len(fb.GetTransactions())))
		for j, t := range fb.GetTransactions() {
			
			prtTrans := filtertransaction(t,addresses)
			
			if j != 0 {
				usertranscnt++
				if prtTrans {
					out.WriteString(fmt.Sprintf("Transaction %d Block Height %d\n",usertranscnt,i))
					blkempty = false
				}
			}else{
				coinbasetranscnt++
			}
			if prtTrans {
				if j==0 && len(t.GetOutputs()) == 0 {
					out.WriteString("\nEmpty Coinbase Transaction\n\n")
				}else if j==0 {
					out.WriteString("\nCoinbase Transaction\n")
					out.WriteString(fmt.Sprintf("%s\n",t.String()))
				}else{
					out.WriteString(fmt.Sprintf("%s\n",t.String()))
				}
			}
		}
		if !skippedblk && blkempty {
			skippedblk = true
			firstemptyblock = i
		}
		if blkempty {
			skippedblk = true
		}
		if !blkempty && skippedblk {
			if i-1 == firstemptyblock {
				ret.WriteString(fmt.Sprintf("Skipped block %d\n\n",firstemptyblock))
			}else{
				ret.WriteString(fmt.Sprintf("Skipped blocks %d-%d\n\n",firstemptyblock,i-1))
			}
			skippedblk = false
		}
		if !blkempty {
			ret.WriteString(out.String())
		}
	}
	i := len(FactoidBlocks)-1
	if skippedblk {
		if i == firstemptyblock {
			ret.WriteString(fmt.Sprintf("Skipped block %d\n\n",firstemptyblock))
		}else{
			ret.WriteString(fmt.Sprintf("Skipped blocks %d-%d\n\n",firstemptyblock,i))
		}
	}
	return ret.Bytes(), nil
}

// At some point we will need to be smarter... Process Blocks and transactions here!
func ProcessFB(fb block.IFBlock) error {
	return nil
}