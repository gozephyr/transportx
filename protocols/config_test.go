package protocols

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type dummyConfig struct{}

func (d *dummyConfig) Validate() error                    { return nil }
func (d *dummyConfig) Merge(other Config) (Config, error) { return d, nil }
func (d *dummyConfig) Clone() Config                      { return &dummyConfig{} }

func TestNewBaseConfigDefaults(t *testing.T) {
	cfg := NewBaseConfig()
	require.Equal(t, 30*time.Second, cfg.Timeout)
	require.Equal(t, 3, cfg.MaxRetries)
	require.NotNil(t, cfg.Metadata)
}

func TestBaseConfigValidate(t *testing.T) {
	cfg := NewBaseConfig()
	require.NoError(t, cfg.Validate())

	cfg.Timeout = 0
	require.Error(t, cfg.Validate())
	cfg.Timeout = 1 * time.Second
	cfg.MaxRetries = -1
	require.Error(t, cfg.Validate())
}

func TestBaseConfigMerge(t *testing.T) {
	cfg1 := NewBaseConfig()
	cfg1.Timeout = 10 * time.Second
	cfg1.MaxRetries = 2
	cfg1.Metadata["foo"] = "bar"

	cfg2 := NewBaseConfig()
	cfg2.Timeout = 20 * time.Second
	cfg2.MaxRetries = 5
	cfg2.Metadata["baz"] = "qux"

	merged, err := cfg1.Merge(cfg2)
	require.NoError(t, err)
	m := merged.(*BaseConfig)
	require.Equal(t, 20*time.Second, m.Timeout)
	require.Equal(t, 5, m.MaxRetries)
	require.Equal(t, "bar", m.Metadata["foo"])
	require.Equal(t, "qux", m.Metadata["baz"])

	_, err = cfg1.Merge(&dummyConfig{})
	require.Error(t, err)
}

func TestBaseConfigClone(t *testing.T) {
	cfg := NewBaseConfig()
	cfg.Timeout = 15 * time.Second
	cfg.MaxRetries = 7
	cfg.Metadata["x"] = "y"
	clone := cfg.Clone().(*BaseConfig)
	require.Equal(t, cfg.Timeout, clone.Timeout)
	require.Equal(t, cfg.MaxRetries, clone.MaxRetries)
	require.Equal(t, cfg.Metadata["x"], clone.Metadata["x"])
	// Ensure deep copy
	clone.Metadata["x"] = "z"
	require.NotEqual(t, cfg.Metadata["x"], clone.Metadata["x"])
}

func TestBaseConfigSettersAndGetters(t *testing.T) {
	cfg := NewBaseConfig()
	cfg.SetTimeout(42 * time.Second)
	require.Equal(t, 42*time.Second, cfg.GetTimeout())
	cfg.SetMaxRetries(9)
	require.Equal(t, 9, cfg.GetMaxRetries())
	cfg.SetMetadata("foo", "bar")
	v, ok := cfg.GetMetadata("foo")
	require.True(t, ok)
	require.Equal(t, "bar", v)
	_, ok = cfg.GetMetadata("notfound")
	require.False(t, ok)
}
