// Filename: internal/data/forum.go

package data

import (
	"time"

	"AWD_FinalProject.ryanarmstrong.net/internal/validator"
)

type Forum struct {
	ID        int64     `json:"id"` // Struct tags
	CreatedAt time.Time `json:"-"`  // doesn't display to client
	Name      string    `json:"name"`
	Level     string    `json:"level"`
	Contact   string    `json:"contat"`
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
