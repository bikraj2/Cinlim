package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cinlim.bikraj.net/internal/validator"
	"github.com/lib/pq"
)

type Movie struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Year      int32     `json:"year"`
	Runtime   Runtime   `json:"runtime"`
	Genres    []string  `json:"genres"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, input *Movie) {

	v.Check(input.Title != "", "title", "The title cannot be empty")
	v.Check(len(input.Title) <= 500, "title", "The title cannot be longer than 500 characters.")
	//Check for Year
	v.Check(input.Year != 0, "year", "The must be provided")
	v.Check(input.Year >= 1888, "year", "Must be greater than 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "The movie cannot be infuture")
	// Check for Runtime
	v.Check(input.Runtime != 0, "runtime", "Must be provided")
	v.Check(input.Runtime > 0, "runtime", "Must be positive")
	//Check for genres
	v.Check(len(input.Genres) >= 1, "genre", "must contain at least one genre")
	v.Check(len(input.Genres) <= 5, "genre", "must cannot contain more than 5 genres")
	v.Check(input.Genres != nil, "genr ", "Must be provided")
	//Check for uniqueness

	v.Check(validator.Unique(input.Genres), "genres", "Must be Unique")
	// Return a Invalidated response if any of the check failed

}

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
  INSERT INTO movies(title,year,runtime,genres)
  VALUES($1,$2,$3,$4)
  RETURNING id, created_at,version
  `
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	// Using queryRow() as we need to execute the row
	return m.DB.QueryRow(query, args...).Scan(&movie.Id, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {

	if id < 1 {
		return nil, ErrNoRecordFound
	}

	query := `
  SELECT id,created_at,title,year,runtime,genres, version
  FROM movies
  Where id = $1 
  `
	var movie Movie
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(&movie.Id, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecordFound
		default:
			return nil, err
		}

	}
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query :=
		`
  UPDATE movies 
  SET title = $1,year=$2,runtime=$3,genres=$4,version=version+1
  WHERE id = $5  AND version  = $6
  RETURNING version
  `
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.Id,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {

		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrNoRecordFound
	}
	query :=
		`
DELETE FROM movies WHERE id = $1`
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoRecordFound
	}
	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filter) ([]*Movie, PageMetaData, error) {
	query := fmt.Sprintf(`   
  SELECT count(*) OVER(),id,created_at,title,year,runtime,genres,version
  FROM movies
  WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
  AND (genres @> $2 OR $2 ='{}')
  ORDER BY  %s %s ,id ASC
  LIMIT $3 OFFSET $4
  `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	fmt.Println(title)
	rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres), filters.PageSize, filters.Page)
	if err != nil {

		return nil, PageMetaData{}, err
	}
	defer rows.Close()
	movies := []*Movie{}
	var totalRecords int

	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&totalRecords,
			&movie.Id,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, PageMetaData{}, err
		}
		movies = append(movies, &movie)
	}
	if err := rows.Err(); err != nil {
		return nil, PageMetaData{}, err
	}

	return movies, calculateMetadata(totalRecords, filters.Page, filters.PageSize), nil
}
