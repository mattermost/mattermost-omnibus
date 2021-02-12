package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigPreSave(t *testing.T) {
	t.Run("nginx_template pointer should be nil after PreSave if its empty string", func(t *testing.T) {
		baseConfig := &Config{
			NginxTemplate: NewString("/some/path.conf"),
		}

		cfg, err := baseConfig.PreSave()
		require.NoError(t, err)
		require.NotNil(t, cfg.NginxTemplate)
		require.Equal(t, *baseConfig.NginxTemplate, *cfg.NginxTemplate)

		baseConfig.NginxTemplate = NewString("")
		cfg, err = baseConfig.PreSave()
		require.NoError(t, err)
		require.Nil(t, cfg.NginxTemplate)
	})
}
