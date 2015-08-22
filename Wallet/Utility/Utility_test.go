package Utility_test

import (
	//"fmt"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
	"testing"
)

func TestIsValidHexAddress(t *testing.T) {
	validHexes := []string{"dceb1ce5778444e7777172e1f586488d2382fb1037887cd79a70b0cba4fb3dce", "9881aeb264452a4f7fafa1cc7bc4b93a05c55537c0703453e585f6d83ce77dca"}
	invalidHexes := []string{"", "cat", "deadbeef", "FA3eNd17NgaXZA3rXQVvzSvWHrpXfHWPzLQjJy2PQVQSc4ZutjC1", "FA38F8fY6duMqDLyCNUYWemdFSWgXDSteeNvNCmJ1Eyb86Z3VNZo"}

	for _, v := range validHexes {
		valid := Utility.IsValidHexAddress(v)
		if valid == false {
			t.Errorf("IsValidHexAddress returned false for valid key `%v`\n", v)
		}
	}
	for _, v := range invalidHexes {
		valid := Utility.IsValidHexAddress(v)
		if valid {
			t.Errorf("IsValidHexAddress returned true for invalid key `%v`\n", v)
		}
	}
}
