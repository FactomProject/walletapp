// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"bytes"
	"encoding/hex"
	"github.com/hoisie/web"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factoid/block"
	"github.com/FactomProject/fctwallet/Wallet"
	"github.com/FactomProject/FactomCode/common"
)

// Older blocks smaller indexes.  All the Factoid Directory blocks
var DirectoryBlocks  = make([]*common.DirectoryBlock,0,100)
var FactoidBlocks    = make([]block.IFBlock,0,100)
var DBHead    []byte
var DBHeadStr string = ""

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
		if bytes.Equal(db.Header.PrevKeyMR.Bytes(),common.ZERO_HASH[:]) {
			break
		}
		next = hex.EncodeToString(db.Header.PrevKeyMR.Bytes())
	}
	
	for i:= len(dbs)-1;i>=0; i-- {
		DirectoryBlocks = append(DirectoryBlocks,dbs[i])
		fb := new(block.FBlock)
		for _,dbe := range dbs[i].DBEntries {
			if bytes.Equal(dbe.ChainID.Bytes(),common.FACTOID_CHAINID) {
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
			fmt.Println("Missing Factoid Block")
		}
	}
	return nil
}

func refresh() error {
	if DBHead == nil {
		getDBHead()
		getAll()
	}
	if getDBHead() {
			
	}
	return nil
}

func FctBalance(adr string) (int64, error) {
	err := refresh()
	if err != nil {
		return 0, err
	}
	return Wallet.FactoidBalance(adr)
}

func ECBalance(adr string) (int64, error) {
	return Wallet.ECBalance(adr)
}

func HandleEntryCreditBalance(ctx *web.Context, adr string) {
	v, err := ECBalance(adr)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}
	str := fmt.Sprintf("%d", v)
	reportResults(ctx, str, true)
}

func HandleFactoidBalance(ctx *web.Context, adr string) {
	v, err := FctBalance(adr)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	str := fmt.Sprintf("%d", v)
	reportResults(ctx, str, true)
}

