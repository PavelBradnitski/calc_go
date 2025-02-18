package repositories

import (
	"context"
	"database/sql"

	"github.com/PavelBradnitski/calc_go/internal/models"
)

type RateRepository struct {
	db *sql.DB
}
type RateRepositoryInterface interface {
	Add(ctx context.Context, result float64) (int64, error)
	Get(ctx context.Context) ([]models.Expression, error)
	GetById(ctx context.Context, id int) (*models.Expression, error)
}

func NewRateRepository(db *sql.DB) RateRepositoryInterface {
	return &RateRepository{db: db}
}

func (r *RateRepository) Add(ctx context.Context, result float64) (int64, error) {
	query := `INSERT INTO expressions (status,result) VALUES (?, ?)`
	res, err := r.db.ExecContext(ctx, query, "DONE", result)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RateRepository) Get(ctx context.Context) ([]models.Expression, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT ID,Status,Result FROM expressions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expressions []models.Expression
	for rows.Next() {
		var expression models.Expression
		err := rows.Scan(&expression.ID, &expression.Status, &expression.Result)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expression)
	}

	return expressions, nil
}

func (r *RateRepository) GetById(ctx context.Context, id int) (*models.Expression, error) {
	var expression models.Expression
	query := `
		SELECT ID,Status,Result FROM expressions
		WHERE ID = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&expression.ID, &expression.Status, &expression.Result)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	} else if err != nil {
		return nil, err
	}
	return &expression, nil
}
