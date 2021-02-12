package cmd

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost-omnibus/mmomni/model"

	"github.com/stretchr/testify/require"
)

func TestGetAutoBackupPath(t *testing.T) {
	t.Run("Should correctly resolve the backup path", func(t *testing.T) {
		tm := time.Date(2009, time.November, 10, 23, 35, 44, 0, time.UTC)
		expected := model.AUTO_BACKUP_DIR + "/mmobackup_20091110_233544.tgz"

		res := getAutoBackupPath(model.AUTO_BACKUP_DIR, tm)
		require.Equal(t, expected, res)
	})
}
