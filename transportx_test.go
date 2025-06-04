package transportx

import (
	"testing"

	"github.com/gozephyr/transportx/errors"
	"github.com/gozephyr/transportx/protocols/grpc"
	"github.com/gozephyr/transportx/protocols/http"
	"github.com/stretchr/testify/require"
)

func TestNewTransportxAndWithMethods(t *testing.T) {
	cfg := http.NewConfig()
	tx := NewTransportx(TypeHTTP, cfg)
	require.NotNil(t, tx)
	require.Equal(t, TypeHTTP, tx.transportType)
	require.Equal(t, cfg, tx.config)

	// Test WithType
	tx2 := tx.WithType(TypeGRPC)
	require.Equal(t, TypeGRPC, tx2.transportType)

	// Test WithConfig (using http.Config, which implements protocols.Config)
	cfg2 := http.NewConfig()
	tx3 := tx2.WithConfig(cfg2)
	require.Equal(t, cfg2, tx3.config)
}

func TestCreateHTTP(t *testing.T) {
	tx := NewTransportx(TypeHTTP, nil)
	tr, err := tx.CreateHTTP(nil)
	require.NoError(t, err)
	require.NotNil(t, tr)

	cfg := http.NewConfig()
	tr2, err := tx.CreateHTTP(cfg)
	require.NoError(t, err)
	require.NotNil(t, tr2)
}

func TestCreateGRPC(t *testing.T) {
	tx := NewTransportx(TypeGRPC, nil)
	tr, err := tx.CreateGRPC(nil)
	require.NoError(t, err)
	require.NotNil(t, tr)

	cfg := grpc.NewConfig()
	tr2, err := tx.CreateGRPC(cfg)
	require.NoError(t, err)
	require.NotNil(t, tr2)
}

func TestCreateUnsupportedTransports(t *testing.T) {
	tx := NewTransportx(TypeHTTP, nil)

	_, err := tx.CreateTCP()
	require.ErrorIs(t, err, errors.ErrUnsupportedTransport)

	_, err = tx.CreateUDP()
	require.ErrorIs(t, err, errors.ErrUnsupportedTransport)

	_, err = tx.CreateWebSocket()
	require.ErrorIs(t, err, errors.ErrUnsupportedTransport)

	_, err = tx.CreateQUIC()
	require.ErrorIs(t, err, errors.ErrUnsupportedTransport)

	_, err = tx.CreateMQTT()
	require.ErrorIs(t, err, errors.ErrUnsupportedTransport)

	_, err = tx.CreateAMQP()
	require.ErrorIs(t, err, errors.ErrUnsupportedTransport)
}
