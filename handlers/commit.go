// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"github.com/hoisie/web"
	"io/ioutil"
	"fmt"

    "github.com/FactomProject/fctwallet/Wallet"
)

func HandleCommitChain(ctx *web.Context, name string) {
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Println("Could not read from http request:", err)
		ctx.WriteHeader(httpBad)
		return
	}

	err=Wallet.CommitChain(name, data)
	if err != nil {
		fmt.Println(err)
		ctx.WriteHeader(httpBad)
		return
	}
}

func HandleCommitEntry(ctx *web.Context, name string) {
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Println("Could not read from http request:", err)
		ctx.WriteHeader(httpBad)
		return
	}

	err=Wallet.CommitEntry(name, data)
	if err != nil {
		fmt.Println(err)
		ctx.WriteHeader(httpBad)
		return
	}
}
