package models

import (
	"encoding/json"
	"fmt"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
)

// User is used by pop to map your users database table to your go code.
type User struct {
	ID        int        `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	FirstName string     `json:"first_name" db:"first_name" form:"firstname"`
	LastName  string     `json:"last_name" db:"last_name" form:"lastname"`
	Contracts []Contract `json:"contracts,omitempty" has_many:"contracts"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// FullName prints the first and last names.
func (u User) FullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
