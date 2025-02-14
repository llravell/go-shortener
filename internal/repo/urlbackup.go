package repo

import (
	"bufio"
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/afero"

	"github.com/llravell/go-shortener/internal/entity"
)

const backupFilePermissions = 0o666

// URLBackup предоставляет интерфейс для сохранения таблицы урлов на диск.
type URLBackup struct {
	file afero.File
}

// NewURLBackup конфигурирует бэкап, открывает файл для записи.
func NewURLBackup(filename string) (*URLBackup, error) {
	fs := afero.NewOsFs()

	file, err := fs.OpenFile(filename, os.O_RDWR|os.O_CREATE, backupFilePermissions)
	if err != nil {
		return nil, err
	}

	return &URLBackup{
		file: file,
	}, nil
}

// Restore восстанавливает таблицу с диска, возвращает список сохраненных урлов.
func (u *URLBackup) Restore() ([]*entity.URL, error) {
	urls := make([]*entity.URL, 0)
	scanner := bufio.NewScanner(u.file)

	for {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return urls, err
			}

			break
		}

		var url entity.URL

		err := json.Unmarshal(scanner.Bytes(), &url)
		if err != nil {
			return urls, err
		}

		urls = append(urls, &url)
	}

	return urls, nil
}

// Store сохраняет переданные урлы на диск.
func (u *URLBackup) Store(urls []*entity.URL) error {
	_, err := u.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = u.file.Truncate(0)
	if err != nil {
		return err
	}

	wr := bufio.NewWriter(u.file)

	for _, url := range urls {
		var data []byte

		data, err = json.Marshal(url)
		if err != nil {
			return err
		}

		data = append(data, '\n')

		_, err = wr.Write(data)
		if err != nil {
			return err
		}
	}

	err = wr.Flush()
	if err != nil {
		return err
	}

	return nil
}

// Close закрывает файл бэкапа.
func (u *URLBackup) Close() error {
	return u.file.Close()
}
