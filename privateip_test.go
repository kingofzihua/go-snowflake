package snowflake_test

import (
	"testing"

	"github.com/kingofzihua/go-snowflake"
)

func TestPrivateIPToMachineID(t *testing.T) {
	mid := snowflake.PrivateIPToMachineID()
	if mid <= 0 {
		t.Error("MachineID should be > 0")
	}

	t.Log(mid)
}
