package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Storage struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewPostgresStorage(db *pgxpool.Pool, logger *zap.Logger) *Storage {
	return &Storage{
		db:  db,
		log: logger,
	}
}

func (s *Storage) IsUserExists(userId int64) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE id = $1
		)
	`
	err := s.db.QueryRow(context.Background(), query, userId).Scan(&exists)
	if err != nil {
		s.log.Error("Ошибка проверки существования пользователя", zap.Int64("userID", userId), zap.Error(err))
		return false, fmt.Errorf("ошибка проверки существования пользователя: %w", err)
	}

	s.log.Debug("Проверка существования пользователя", zap.Int64("userID", userId), zap.Bool("exists", exists))
	return exists, nil
}

func (s *Storage) AddUser(userId int64) error {
	query := `
		INSERT INTO users (id)
		VALUES ($1)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := s.db.Exec(context.Background(), query, userId)
	if err != nil {
		s.log.Error("Ошибка добавления пользователя", zap.Int64("userID", userId), zap.Error(err))
		return fmt.Errorf("ошибка добавления пользователя: %w", err)
	}

	s.log.Info("Пользователь успешно добавлен", zap.Int64("userID", userId))
	return nil
}

func (s *Storage) SaveModel(userID int64, modelName string) error {
	query := `
		INSERT INTO models (user_id, name)
		VALUES ($1, $2)
	`
	_, err := s.db.Exec(context.Background(), query, userID, modelName)
	if err != nil {
		s.log.Error("Ошибка сохранения модели", zap.Int64("userID", userID), zap.String("modelName", modelName), zap.Error(err))
		return fmt.Errorf("ошибка сохранения модели: %w", err)
	}

	s.log.Info("Модель успешно сохранена", zap.Int64("userID", userID), zap.String("modelName", modelName))
	return nil
}

func (s *Storage) GetUserModels(userID int64) ([]string, error) {
	query := `
		SELECT name
		FROM models
		WHERE user_id = $1
		ORDER BY created_at
	`
	rows, err := s.db.Query(context.Background(), query, userID)
	if err != nil {
		s.log.Error("Ошибка получения списка моделей", zap.Int64("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("ошибка получения моделей: %w", err)
	}
	defer rows.Close()

	var models []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			s.log.Error("Ошибка чтения модели из строки", zap.Int64("userID", userID), zap.Error(err))
			return nil, fmt.Errorf("ошибка чтения модели: %w", err)
		}
		models = append(models, name)
	}

	if err := rows.Err(); err != nil {
		s.log.Error("Ошибка после чтения всех строк моделей", zap.Int64("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("ошибка в строках результата: %w", err)
	}

	s.log.Info("Успешно получены модели пользователя", zap.Int64("userID", userID), zap.Int("modelsCount", len(models)))
	return models, nil
}

func (s *Storage) CountModels(userID int64) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM models WHERE user_id = $1
	`
	err := s.db.QueryRow(context.Background(), query, userID).Scan(&count)
	if err != nil {
		s.log.Error("Ошибка подсчёта моделей пользователя", zap.Int64("userID", userID), zap.Error(err))
		return 0, err
	}

	s.log.Info("Подсчитано количество моделей", zap.Int64("userID", userID), zap.Int("modelsCount", count))
	return count, nil
}

func (s *Storage) DeleteModel(userID int64, modelName string) error {
	query := `
		DELETE FROM models
		WHERE user_id = $1 AND name = $2
	`
	_, err := s.db.Exec(context.Background(), query, userID, modelName)
	if err != nil {
		return fmt.Errorf("ошибка удаления модели: %w", err)
	}
	return nil
}
