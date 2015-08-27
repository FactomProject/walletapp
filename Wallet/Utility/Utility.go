// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package Utility

import (
	"github.com/FactomProject/factoid"
	"regexp"
	"strings"
)

var badChar, _ = regexp.Compile("[^A-Za-z0-9_-]")                                                      //alphanumeric plus _-
var badHexChar, _ = regexp.Compile("[^A-Fa-f0-9]")                                                     //hexadecimal
var badBase58Char, _ = regexp.Compile("[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]") //Base58 alphabet

var HUMAN_ADDRESS_LENGTH int = 52
var NICKNAME_LENGTH int = 64

func IsValidAddress(address string) bool {

	if len(address) != HUMAN_ADDRESS_LENGTH {
		return false
	}

	if badBase58Char.FindStringIndex(address) != nil {
		return false
	}

	if !strings.HasPrefix(address, "FA") &&
		!strings.HasPrefix(address, "EC") {
		return false
	}

	return true
}

func IsValidHex(h string) bool {
	if badHexChar.FindStringIndex(h) != nil {
		return false
	}
	return true
}

func IsValidHexAddress(address string) bool {
	if len(address) != 2*factoid.ADDRESS_LENGTH {
		return false
	}
	if badHexChar.FindStringIndex(address) != nil {
		return false
	}
	return true
}

func IsValidNickname(nick string) bool {
	if len(nick) > NICKNAME_LENGTH {
		return false
	}
	if len(nick) == 0 {
		return false
	}
	if badChar.FindStringIndex(nick) != nil {
		return false
	}
	return true
}

func IsValidKey(key string) bool {
	return IsValidAddress(key) || IsValidHexAddress(key) || IsValidNickname(key)
}
