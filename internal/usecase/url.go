package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog"

	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/repo"
)

// ErrURLDuplicate ошибка создания дубля.
var ErrURLDuplicate = errors.New("duplicate url")

// ErrURLInvalid ошибка получения некорректного урла.
var ErrURLInvalid = errors.New("invalid url")

// URLDeleteWorkerPool пул, обрабатывающий удаление урлов.
type URLDeleteWorkerPool interface {
	QueueWork(w *URLDeleteWork) error
}

// URLDeleteWork задача удаления урлов.
type URLDeleteWork struct {
	repo     URLRepo
	log      *zerolog.Logger
	UserUUID string
	Hashes   []string
}

// Do удаляет урлы пользователя.
func (w *URLDeleteWork) Do(ctx context.Context) {
	err := w.repo.DeleteMultipleURLs(ctx, w.UserUUID, w.Hashes)
	if err != nil {
		w.log.Error().
			Err(err).
			Str("userUUID", w.UserUUID).
			Msg("delete urls failed")

		return
	}

	w.log.Info().
		Str("userUUID", w.UserUUID).
		Msg("delete urls successeded")
}

// URLUseCase юзкейс базовых операций с урлами.
type URLUseCase struct {
	repo            URLRepo
	wp              URLDeleteWorkerPool
	gen             HashGenerator
	log             zerolog.Logger
	baseRedirectURL string
}

// NewURLUseCase создает юзкейс.
func NewURLUseCase(
	repo URLRepo,
	wp URLDeleteWorkerPool,
	gen HashGenerator,
	baseRedirectURL string,
	log zerolog.Logger,
) *URLUseCase {
	return &URLUseCase{
		repo:            repo,
		wp:              wp,
		gen:             gen,
		log:             log,
		baseRedirectURL: baseRedirectURL,
	}
}

// SaveURL сохраняет урл.
func (uc *URLUseCase) SaveURL(ctx context.Context, url string, userUUID string) (*entity.URL, error) {
	hash, err := uc.gen.Generate()
	if err != nil {
		return nil, err
	}

	urlObj := &entity.URL{Original: url, Short: hash, UserUUID: userUUID}

	storedURL, err := uc.repo.Store(ctx, urlObj)
	if errors.Is(err, repo.ErrOriginalURLConflict) {
		return storedURL, ErrURLDuplicate
	}

	return storedURL, err
}

// SaveURLMultiple сохраняет несколько урлов.
func (uc *URLUseCase) SaveURLMultiple(ctx context.Context, urls []string, userUUID string) ([]*entity.URL, error) {
	urlObjs := make([]*entity.URL, 0, len(urls))

	if len(urls) == 0 {
		return urlObjs, nil
	}

	for _, url := range urls {
		hash, err := uc.gen.Generate()
		if err != nil {
			return urlObjs, err
		}

		urlObjs = append(urlObjs, &entity.URL{Original: url, Short: hash, UserUUID: userUUID})
	}

	return urlObjs, uc.repo.StoreMultipleURLs(ctx, urlObjs)
}

// ResolveURL определяет полный урл по сокращенному.
func (uc *URLUseCase) ResolveURL(ctx context.Context, shortenURL string) (*entity.URL, error) {
	baseURLObj, err := url.Parse(uc.baseRedirectURL)
	if err != nil {
		return nil, err
	}

	shortenURLObj, err := url.Parse(shortenURL)
	if err != nil {
		return nil, err
	}

	if baseURLObj.Scheme != shortenURLObj.Scheme ||
		baseURLObj.Host != shortenURLObj.Host {
		return nil, ErrURLInvalid
	}

	return uc.ResolveURLByHash(ctx, strings.TrimPrefix(shortenURLObj.Path, "/"))
}

// ResolveURLByHash определяет полный урл по хэшу.
func (uc *URLUseCase) ResolveURLByHash(ctx context.Context, hash string) (*entity.URL, error) {
	return uc.repo.GetURL(ctx, hash)
}

// GetUserURLS находит все урлы пользователя.
func (uc *URLUseCase) GetUserURLS(ctx context.Context, userUUID string) ([]*entity.URL, error) {
	return uc.repo.GetUserURLS(ctx, userUUID)
}

// BuildRedirectURL формирует урл для редиректа.
func (uc *URLUseCase) BuildRedirectURL(url *entity.URL) string {
	return fmt.Sprintf("%s/%s", uc.baseRedirectURL, url.Short)
}

// QueueDelete отправляет задачу на удаление урлов в пул воркеров.
func (uc *URLUseCase) QueueDelete(deleteItem *entity.URLDeleteItem) error {
	deleteWork := &URLDeleteWork{
		repo:     uc.repo,
		log:      &uc.log,
		UserUUID: deleteItem.UserUUID,
		Hashes:   deleteItem.Hashes,
	}

	return uc.wp.QueueWork(deleteWork)
}
