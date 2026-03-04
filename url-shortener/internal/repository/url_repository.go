package repository

import (
	"database/sql"
	"url-shortener/internal/model"
)

type URLRepository struct {
	DB *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{DB: db}
}
func (r *URLRepository) CreateURL(originalURL string) (*model.URL, error) {

	query := `
	INSERT INTO urls (original_url)
	VALUES ($1)
	RETURNING id
	`

	var id int64

	err := r.DB.QueryRow(query, originalURL).Scan(&id)
	if err != nil {
		return nil, err
	}

	url := &model.URL{
		ID:          id,
		OriginalURL: originalURL,
	}

	return url, nil
}
func (r *URLRepository) UpdateShortCode(id int64, code string) error {

	query := `
	UPDATE urls
	SET short_code = $1
	WHERE id = $2
	`

	_, err := r.DB.Exec(query, code, id)

	return err
}
func (r *URLRepository) GetByShortCode(code string) (*model.URL, error) {

	query := `
	SELECT id, original_url, short_code, created_at
	FROM urls
	WHERE short_code = $1
	`

	row := r.DB.QueryRow(query, code)

	var url model.URL

	err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &url, nil
}
