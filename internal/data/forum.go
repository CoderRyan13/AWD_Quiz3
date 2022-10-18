// Filename: internal/data/forum.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"AWD_FinalProject.ryanarmstrong.net/internal/validator"
	"github.com/lib/pq"
)

type Forum struct {
	ID        int64     `json:"id"` // Struct tags
	CreatedAt time.Time `json:"-"`  // doesn't display to client
	Name      string    `json:"name"`
	Level     string    `json:"level"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email,omitempty"`
	Website   string    `json:"website,omitempty"`
	Address   string    `json:"address"`
	Mode      []string  `json:"mode"`
	Version   int32     `json:"version"`
}

func ValidateForum(v *validator.Validator, forum *Forum) {
	// Use the Check() method to execute our validation checks
	v.Check(forum.Name != "", "name", "must be provided")
	v.Check(len(forum.Name) <= 200, "name", "must not be more than 200 bytes long")

	v.Check(forum.Level != "", "level", "must be provided")
	v.Check(len(forum.Level) <= 200, "level", "must not be more than 200 bytes long")

	v.Check(forum.Contact != "", "contact", "must be provided")
	v.Check(len(forum.Contact) <= 200, "contact", "must not be more than 200 bytes long")

	v.Check(forum.Phone != "", "phone", "must be provided")
	v.Check(validator.Matches(forum.Phone, validator.PhoneRX), "phone", "must be a valid phone number")

	v.Check(forum.Email != "", "email", "must be provided")
	v.Check(validator.Matches(forum.Email, validator.EmailRX), "email", "must be a valid email address")

	v.Check(forum.Website != "", "website", "must be provided")
	v.Check(validator.ValidWebsite(forum.Website), "website", "must be a valid URL")

	v.Check(forum.Address != "", "address", "must be provided")
	v.Check(len(forum.Address) <= 500, "address", "must not be more than 500 bytes long")

	v.Check(forum.Mode != nil, "mode", "must be provided")
	v.Check(len(forum.Mode) >= 1, "mode", "must contain at least 1 entry")
	v.Check(len(forum.Mode) <= 5, "mode", "must contain at most 5 entries")
	v.Check(validator.Unique(forum.Mode), "mode", "must not contain duplicate entries")
}

// Define a ForumModel which wraps a sql.DB connection pool
type ForumModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new Forum
func (m ForumModel) Insert(forum *Forum) error {
	query := `
		INSERT INTO forums (name, level, contact, phone, email, website, address, mode)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, version
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Collect the data fields into a slice
	args := []interface{}{
		forum.Name, forum.Level,
		forum.Contact, forum.Phone,
		forum.Email, forum.Website,
		forum.Address, pq.Array(forum.Mode),
	}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&forum.ID, &forum.CreatedAt, &forum.Version)
}

// Get() allows us to recieve a specific Forum
func (m ForumModel) Get(id int64) (*Forum, error) {
	// Ensure that there is a valid id
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create the query
	query := `
		SELECT id, created_at, name, level, contact, phone, email, website, address, mode, version
		FROM forums
		WHERE id = $1
	`
	// Declare a Forum variable to hold the returned data
	var forum Forum
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query using QueryRow()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&forum.ID,
		&forum.CreatedAt,
		&forum.Name,
		&forum.Level,
		&forum.Contact,
		&forum.Phone,
		&forum.Email,
		&forum.Website,
		&forum.Address,
		pq.Array(&forum.Mode),
		&forum.Version,
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
	return &forum, nil
}

// Update() allows us to edit/alter a specific Forum
// Optimistic locking (version number)
func (m ForumModel) Update(forum *Forum) error {
	// Create a query
	query := `
		UPDATE forums
		SET name = $1, level = $2, contact = $3, 
			phone = $4, email = $5, website = $6,
			address = $7, mode = $8, version = version + 1
		WHERE id = $9
		AND version = $10
		RETURNING version
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	args := []interface{}{
		forum.Name,
		forum.Level,
		forum.Contact,
		forum.Phone,
		forum.Email,
		forum.Website,
		forum.Address,
		pq.Array(forum.Mode),
		forum.ID,
		forum.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&forum.Version)
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

// Delete() removes a specific Forum
func (m ForumModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM forums
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

// the GetAll() method returns a list of all the schools sorted by id
func (m ForumModel) GetAll(name string, level string, mode []string, filters Filters) ([]*Forum, Metadata, error) {
	// Construct the query
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, created_at, name, level, 
			   contact, phone, email, website, 
			   address, mode, version
		FROM forums
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple', level) @@ plainto_tsquery('simple', $2) OR $2 = '')
		AND (mode @> $3 OR $3 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortOrder())

	// Create a 3-seconds-timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Execute the query
	args := []interface{}{name, level, pq.Array(mode), filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	// Close the resultset
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice to hold the Forum data
	forums := []*Forum{}
	// Iterate over the rows in the resultset
	for rows.Next() {
		var forum Forum
		// Scan the values from the row into the forum
		err := rows.Scan(
			&totalRecords,
			&forum.ID,
			&forum.CreatedAt,
			&forum.Name,
			&forum.Level,
			&forum.Contact,
			&forum.Phone,
			&forum.Email,
			&forum.Website,
			&forum.Address,
			pq.Array(&forum.Mode),
			&forum.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// Add the Forum to our slice
		forums = append(forums, &forum)
	}
	// Check for errors after looping through the resultset
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Return the slice of Forums
	return forums, metadata, nil
}
