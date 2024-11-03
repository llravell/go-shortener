package repo

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/llravell/go-shortener/internal/entity"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeBackup(t *testing.T, data []byte) *URLBackup {
	t.Helper()

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "test.backup", data, backupFilePermissions)
	require.NoError(t, err)

	file, err := fs.OpenFile("test.backup", os.O_RDWR, backupFilePermissions)
	require.NoError(t, err)

	return &URLBackup{file}
}

func readNextURLFromFile(t *testing.T, file afero.File) *entity.URL {
	t.Helper()

	data, err := bufio.NewReader(file).ReadBytes('\n')
	require.NoError(t, err)

	var url entity.URL

	err = json.Unmarshal(data, &url)
	require.NoError(t, err)

	return &url
}

func TestURLBackup(t *testing.T) {
	url := entity.NewURL("https://foo.ru", "foo")
	urlJSON, err := json.Marshal(url)
	require.NoError(t, err)

	t.Run("Restore from file", func(t *testing.T) {
		backup := makeBackup(t, urlJSON)
		defer backup.Close()

		urls, err := backup.Restore()
		require.NoError(t, err)

		assert.Len(t, urls, 1)
		assert.ObjectsAreEqual(url, *urls[0])
	})

	t.Run("Store to file", func(t *testing.T) {
		backup := makeBackup(t, []byte{})
		defer backup.Close()

		err := backup.Store([]*entity.URL{url})
		require.NoError(t, err)

		_, err = backup.file.Seek(0, io.SeekStart)
		require.NoError(t, err)

		storedURL := readNextURLFromFile(t, backup.file)

		assert.ObjectsAreEqual(url, storedURL)
	})
}
