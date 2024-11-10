package mediator

import (
	"context"
	"fmt"
	"strconv"
	"vieo/auth/internal/cache"
	"vieo/auth/internal/domain/models"
	"vieo/auth/internal/storage/psql"
)

type MovieMediator interface {
	NotifyMovieStart(ctx context.Context, movie *models.Movie)
	NotifyMovieStop(ctx context.Context, movie *models.Movie)
	RegisterUser(user *models.User)
	NotifyUserJoined(ctx context.Context, user *models.User, movie *models.Movie)
	NotifyUserLeft(ctx context.Context, user *models.User, movie *models.Movie)
	GetMovie(ctx context.Context, id string) (*models.Movie, error)
}

// ConcreteMovieMediator реализация медиатора
type ConcreteMovieMediator struct {
	users   []*models.User
	storage psql.Storage
	cache   cache.Cache
}

func NewMovieMediator(storage psql.Storage, cache cache.Cache) *ConcreteMovieMediator {
	return &ConcreteMovieMediator{
		users:   make([]*models.User, 0),
		storage: storage,
		cache:   cache,
	}
}

func (m *ConcreteMovieMediator) GetMovie(ctx context.Context, id string) (*models.Movie, error) {
	// Сначала пробуем получить из кэша
	movie, err := m.cache.GetMovie(ctx, id)
	if err == nil {
		return movie, nil
	}

	// Если в кэше нет, берем из БД
	movie, err = m.storage.GetMovie(ctx, id)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	err = m.cache.SetMovie(ctx, movie)
	if err != nil {
		// Логируем ошибку, но не прерываем работу
		fmt.Printf("Failed to cache movie: %v\n", err)
	}

	return movie, nil
}

func (m *ConcreteMovieMediator) NotifyMovieStart(ctx context.Context, movie *models.Movie) {
	fmt.Printf("Медиатор: Начало показа фильма '%s'\n", movie.Title)

	// Сохраняем информацию в БД и кэш
	err := m.storage.SaveMovie(ctx, movie)
	if err != nil {
		fmt.Printf("Failed to save movie to storage: %v\n", err)
	}

	err = m.cache.SetMovie(ctx, movie)
	if err != nil {
		fmt.Printf("Failed to cache movie: %v\n", err)
	}

	for _, user := range m.users {
		user.ReceiveNotification(fmt.Sprintf("Начался фильм: %s", movie.Title))
	}
}

func (m *ConcreteMovieMediator) NotifyMovieStop(ctx context.Context, movie *models.Movie) {
	fmt.Printf("Медиатор: Окончание показа фильма '%s'\n", movie.Title)

	// Обновляем информацию в БД и инвалидируем кэш
	movie.CurrentViewers = 0
	err := m.storage.SaveMovie(ctx, movie)
	if err != nil {
		fmt.Printf("Failed to update movie in storage: %v\n", err)
	}

	err = m.cache.InvalidateMovie(ctx, strconv.Itoa(movie.ID))
	if err != nil {
		fmt.Printf("Failed to invalidate movie cache: %v\n", err)
	}

	for _, user := range m.users {
		user.ReceiveNotification(fmt.Sprintf("Закончился фильм: %s", movie.Title))
	}
}

func (m *ConcreteMovieMediator) NotifyUserJoined(ctx context.Context, user *models.User, movie *models.Movie) {
	fmt.Printf("Медиатор: Пользователь %s присоединился к просмотру '%s'\n", user.Email, movie.Title)

	movie.CurrentViewers++

	// Обновляем количество зрителей в БД и кэше
	err := m.storage.UpdateViewers(ctx, strconv.Itoa(movie.ID), movie.CurrentViewers)
	if err != nil {
		fmt.Printf("Failed to update viewers in storage: %v\n", err)
	}

	err = m.cache.SetMovie(ctx, movie)
	if err != nil {
		fmt.Printf("Failed to update movie in cache: %v\n", err)
	}
}

func (m *ConcreteMovieMediator) NotifyUserLeft(ctx context.Context, user *models.User, movie *models.Movie) {
	fmt.Printf("Медиатор: Пользователь %s покинул просмотр '%s'\n", user.Email, movie.Title)

	movie.CurrentViewers--

	// Обновляем количество зрителей в БД и кэше
	err := m.storage.UpdateViewers(ctx, strconv.Itoa(movie.ID), movie.CurrentViewers)
	if err != nil {
		fmt.Printf("Failed to update viewers in storage: %v\n", err)
	}

	err = m.cache.SetMovie(ctx, movie)
	if err != nil {
		fmt.Printf("Failed to update movie in cache: %v\n", err)
	}
}
