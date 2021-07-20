package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
)

// Boss is used by pop to map your bosses database table to your go code.
type Boss struct {
	ID             int        `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Contracts      []Contract `json:"contracts,omitempty" has_many:"contracts"`
	CreateContract bool       `json:"-" db:"-"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (b Boss) String() string {
	jb, _ := json.Marshal(b)
	return string(jb)
}

// https://medium.com/@KagundaJM/buffalo-select-tags-example-9d94214c5248
// SelectLabel provides label for select-list.
func (b Boss) SelectLabel() string {
	return b.Name
}

// SelectValue provides value for select-list.
func (b Boss) SelectValue() interface{} {
	return b.ID
}

type Bosses []Boss

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (b *Boss) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (b *Boss) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = time.Now()
	}
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (b *Boss) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
