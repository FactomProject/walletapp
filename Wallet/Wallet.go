package Wallet

import (
	"fmt"
	"regexp"

	fct "github.com/FactomProject/factoid"
)

/******************************************
 * Helper Functions
 ******************************************/

var badChar, _ = regexp.Compile("[^A-Za-z0-9_-]")
var badHexChar, _ = regexp.Compile("[^A-Fa-f0-9]")

type Response struct {
	Response string
	Success  bool
}

func ValidateKey(key string) error {
	if len(key) > fct.ADDRESS_LENGTH {
		return fmt.Errorf("Key is too long.")
	}
	if badChar.FindStringIndex(key) != nil {
		return fmt.Errorf("Key contains invalid characters.")
	}
	return nil
}

func GetTransaction(key string) (trans fct.ITransaction, err error) {
	err = ValidateKey(key)
	if err != nil {
		return nil, err
	}

	// Now get the transaction.  If we don't have a transaction by the given
	// keys there is nothing we can do.  Now we *could* create the transaaction
	// and tie it to the key.  Something to think about.
	ib := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))

	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return nil, fmt.Errorf("Unknown Transaction: %s", key)
	}
	return
}
