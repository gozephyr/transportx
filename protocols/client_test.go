package protocols

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientInterfaceDummy(t *testing.T) {
	var c Client = nil
	require.Nil(t, c)
}
