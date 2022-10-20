// Filename: internal/data/todo.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"AWD_Quiz3.ryanarmstrong.net/internal/validator"
)

type Todo struct {
	ID        int64     `json:"id"` // Struct tags
	CreatedAt time.Time `json:"-"`  // doesn't display to client
	Task      string    `json:"task"`
	Complete  string    `json:"complete"`
	Version   int32     `json:"version"`
}

func ValidateTodo(v *validator.Validator, todo *Todo) {
	// Use the Check() method to execute our validation checks
	v.Check(todo.Task != "", "task", "must be provided")
	v.Check(len(todo.Task) <= 200, "task", "must not be more than 200 bytes long")

	//v.Check(todo.Complete != "", "complete", "must be provided")
	//v.Check(len(todo.Complete) <= 200, "complete", "must not be more than 200 bytes long")

}

// Define a TodoModel which wraps a sql.DB connection pool
type TodoModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new Task
func (m TodoModel) Insert(todo *Todo) error {
	query := `
		INSERT INTO todos (task)
		VALUES ($1)
		RETURNING id, created_at, version, complete
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Collect the data fields into a slice
	args := []interface{}{
		todo.Task,
	}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&todo.ID, &todo.CreatedAt, &todo.Version, &todo.Complete)
}

// Get() allows us to recieve a specific Task
func (m TodoModel) Get(id int64) (*Todo, error) {
	// Ensure that there is a valid id
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create the query
	query := `
		SELECT id, created_at, task, complete, version
		FROM todos
		WHERE id = $1
	`
	// Declare a Todo variable to hold the returned data
	var todo Todo
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query using QueryRow()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.CreatedAt,
		&todo.Task,
		&todo.Complete,
		&todo.Version,
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &todo, nil
}

// Update() allows us to edit/alter a specific Task
// Optimistic locking (version number)
func (m TodoModel) Update(todo *Todo) error {
	// Create a query
	query := `
		UPDATE todos
		SET task = $1, complete = $2, version = version + 1
		WHERE id = $3
		AND version = $4
		RETURNING version
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	args := []interface{}{
		todo.Task,
		todo.Complete,
		todo.ID,
		todo.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&todo.Version)
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

// Delete() removes a specific Task
func (m TodoModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM todos
		WHERE id = $1
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operation. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// the GetAll() method returns a list of all the tasks sorted by id
func (m TodoModel) GetAll(task string, complete string, filters Filters) ([]*Todo, Metadata, error) {
	// Construct the query
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, created_at, task, complete, version
		FROM todos
		WHERE (to_tsvector('simple', task) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple', complete) @@ plainto_tsquery('simple', $2) OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortOrder())

	// Create a 3-seconds-timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Execute the query
	args := []interface{}{task, complete, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	// Close the resultset
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice to hold the Todo data
	todos := []*Todo{}
	// Iterate over the rows in the resultset
	for rows.Next() {
		var todo Todo
		// Scan the values from the row into the task
		err := rows.Scan(
			&totalRecords,
			&todo.ID,
			&todo.CreatedAt,
			&todo.Task,
			&todo.Complete,
			&todo.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// Add the Todo to our slice
		todos = append(todos, &todo)
	}
	// Check for errors after looping through the resultset
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Return the slice of Forums
	return todos, metadata, nil
}
