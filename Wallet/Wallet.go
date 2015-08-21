package Wallet

import (
	"fmt"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
)

/******************************************
 * Helper Functions
 ******************************************/
type Response struct {
	Response string
	Success  bool
}

func ValidateKey(key string) error {
	if Utility.IsValidKey(key) {
		return nil
	}
	return fmt.Errorf("Invalid key")
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
