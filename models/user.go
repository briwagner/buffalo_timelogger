package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User is used by pop to map your users database table to your go code.
type User struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	Email                string     `json:"email" db:"email"`
	FirstName            string     `json:"first_name" db:"first_name" form:"firstname"`
	LastName             string     `json:"last_name" db:"last_name" form:"lastname"`
	Contracts            []Contract `json:"contracts,omitempty" has_many:"contracts"`
	PasswordHash         string     `json:"-" db:"password_hash"`
	Password             string     `json:"-" db:"-"`
	PasswordConfirmation string     `json:"-" db:"-"`
	Roles                string     `json:"roles" db:"roles"`
}

// String is not required by pop and may be deleted
func (u *User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Create wraps up the pattern of encrypting the password and
// running validations. Useful when writing tests.
func (u *User) Create(tx *pop.Connection) (*validate.Errors, error) {
	u.Email = strings.ToLower(u.Email)
	ph, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}
	u.PasswordHash = string(ph)
	return tx.ValidateAndCreate(u)
}

// FullName prints the first and last names.
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return "User has no name"
	}
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// IsAdmin checks the user role.
func (u *User) IsAdmin() bool {
	if u.Roles == "admin" {
		return true
	}
	return false
}

// MakeAdmin changes user role.
func (u *User) SetRole(r string) {
	u.Roles = r
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
	var err error
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.PasswordHash, Name: "PasswordHash"},
		// Check if email is already taken.
		&validators.FuncValidator{
			Field:   u.Email,
			Name:    "Email",
			Message: "%s is already taken",
			Fn: func() bool {
				var b bool
				q := tx.Where("email = ?", u.Email)
				if u.ID != uuid.Nil {
					q = q.Where("id != ?", u.ID)
				}
				b, err = q.Exists(u)
				if err != nil {
					log.Println("user not found")
					return false
				}
				return !b
			},
		},
	), err
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	var err error
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Password, Name: "Password"},
		&validators.StringsMatch{Name: "Password", Field: u.Password, Field2: u.PasswordConfirmation, Message: "Passwords do not match"},
	), err
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// GetContracts
func (u *User) GetContracts(tx *pop.Connection) error {
	contracts := []Contract{}
	err := tx.Where("user_id = ?", u.ID).Eager("Boss").All(&contracts)
	if err != nil {
		return err
	}

	u.Contracts = contracts
	return nil
}
