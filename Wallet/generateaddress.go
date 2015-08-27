// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package Wallet

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factoid"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
)

func GenerateAddress(name string) (factoid.IAddress, error) {
	ok := Utility.IsValidKey(name)
	if !ok {
		return nil, fmt.Errorf("Invalid name or address")
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

func GenerateAddressFromPrivateKey(name string, privateKey string) (factoid.IAddress, error) {
	if Utility.IsValidKey(name) == false {
		return nil, fmt.Errorf("Invalid name or address")
	}
	if len(privateKey) != 64 && len(privateKey) != 128 {
		return nil, fmt.Errorf("Invalid private key length")
	}
	if Utility.IsValidHex(privateKey) == false {
		return nil, fmt.Errorf("Invalid private key format")
	}
	priv, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	addr, err := factoidState.GetWallet().GenerateFctAddressFromPrivateKey([]byte(name), privateKey, 1, 1)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func GenerateAddressStringFromPrivateKey(name string, privateKey string) (string, error) {
	addr, err := GenerateAddressFromPrivateKey(name, privateKey)
	if err != nil {
		return "", err
	}
	return factoid.ConvertECAddressToUserStr(addr), nil
}

func GenerateECAddress(name string) (factoid.IAddress, error) {
	ok := Utility.IsValidKey(name)
	if !ok {
		return nil, fmt.Errorf("Invalid name or address")
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

func GenerateECAddressFromPrivateKey(name string, privateKey string) (factoid.IAddress, error) {
	if Utility.IsValidKey(name) == false {
		return nil, fmt.Errorf("Invalid name or address")
	}
	if len(privateKey) != 64 && len(privateKey) != 128 {
		return nil, fmt.Errorf("Invalid private key length")
	}
	if Utility.IsValidHex(privateKey) == false {
		return nil, fmt.Errorf("Invalid private key format")
	}
	priv, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	addr, err := factoidState.GetWallet().GenerateECAddressFromPrivateKey([]byte(name), priv)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func GenerateECAddressStringFromPrivateKey(name string, privateKey string) (string, error) {
	addr, err := GenerateECAddressFromPrivateKey(name, privateKey)
	if err != nil {
		return "", err
	}
	return factoid.ConvertECAddressToUserStr(addr), nil
}
