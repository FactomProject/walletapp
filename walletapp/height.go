// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main


import (
	"fmt"
	"github.com/FactomProject/factom"
)

/************************************************************
 * Height
 ************************************************************/

type Height struct {
	
}

var _ ICommand = (*Height)(nil)

// Height transactions <address list> 
func (Height) Execute(state IState, args []string) (err error) {
	h,err := factom.GetDBlockHeight()
	if err != nil {
		return fmt.Errorf("Failed to contact the Factom Network")
	}else{
		fmt.Printf("DirectoryBlockHeight=%d\n",h)
	}
	return nil
}
	

	func (Height) Name() string {
	return "Height"
}

func (Height) ShortHelp() string {
	return "Height -- Returns the number of completed Directory Blocks in Factom."
	
}

func (Height) LongHelp() string {
	return `
Height                              Returns the number of completed Directory Blocks in Factom. 
cd`
}



