package models

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Contract is a User's record for a specific boss.
type Contract struct {
	ID        int       `json:"id" db:"id"`
	Rate      int       `json:"rate" db:"rate"`
	BossID    int       `json:"-" db:"boss_id"`
	Boss      *Boss     `json:"boss" belongs_to:"boss"`
	UserID    uuid.UUID `json:"-" db:"user_id"`
	User      *User     `json:"user" belongs_to:"user"`
	Tasks     []Task    `json:"tasks,omitempty" has_many:"tasks"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (c Contract) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

type Contracts []Contract

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (c *Contract) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (c *Contract) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now()
	}
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (c *Contract) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// LoadContract gets a contract with sorted tasks
func (c *Contract) LoadContract(tx *pop.Connection, cid string) error {
	err := tx.Eager().Find(c, cid)
	if err != nil {
		return err
	}

	sort.SliceStable(c.Tasks, func(i, j int) bool {
		return c.Tasks[i].StartTime.Before(c.Tasks[j].StartTime)
	})

	return nil
}
