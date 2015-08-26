// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"encoding/hex"
	"github.com/hoisie/web"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/fctwallet/Wallet"
)

var DBHead    []byte
var DBHeadStr string

func getAll() error {
	db, err := factom.GetDBlockHead()
	if err != nil {
		return err
	}
	
	DBHeadStr = db.KeyMR
	
	return nil
}

func refresh() error {
	if DBHead == nil {
		getAll()
	}else{
		db,err := factom.GetDBlockHead()
		if err != nil {
			return err
		}
		if db.KeyMR != DBHeadStr {
			DBHeadStr = db.KeyMR
			DBHead, err = hex.DecodeString(db.KeyMR)
			if err != nil {
				return err
			}
		}
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

