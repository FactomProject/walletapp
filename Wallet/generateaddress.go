// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package Wallet

import (
	"github.com/FactomProject/factoid"
)

func GenerateAddress(name string) (factoid.IAddress, error) {
	err := ValidateKey(name)
	if err != nil {
		return nil, err
	}
	addr, err := factoidState.GetWallet().GenerateFctAddress([]byte(name), 1, 1)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func GenerateAddressString(name string) (string, error) {
	addr, err := GenerateAddress(name)
	if err != nil {
		return "", err
	}
	return factoid.ConvertECAddressToUserStr(addr), nil
}

func GenerateECAddress(name string) (factoid.IAddress, error) {
	err := ValidateKey(name)
	if err != nil {
		return nil, err
	}
	addr, err := factoidState.GetWallet().GenerateECAddress([]byte(name))
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func GenerateECAddressString(name string) (string, error) {
	addr, err := GenerateECAddress(name)
	if err != nil {
		return "", err
	}
	return factoid.ConvertECAddressToUserStr(addr), nil
}
