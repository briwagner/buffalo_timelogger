package actions

import (
	"database/sql"
	"strings"

	"buftester/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/pkg/errors"
)

// AnonOnly rejects logged-in users.
func AnonOnly(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		uid := c.Session().Get("current_user_id")
		if uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err != nil {
				return errors.WithStack(err)
			}
			c.Set("current_user", u)
			c.Flash().Add("success", "You are logged in")
			return c.Redirect(302, "/users/%s", u.ID)
		}
		return next(c)
	}
}

// AuthNew loads the signin page
func AuthNew(c buffalo.Context) error {
	c.Set("user", models.User{})
	return c.Render(200, r.HTML("auth/new.plush.html"))
}

// AuthCreate attempts to log the user in with an existing account.
func AuthCreate(c buffalo.Context) error {
	u := &models.User{}
	if err := c.Bind(u); err != nil {
		return errors.WithStack(err)
	}

	tx := c.Value("tx").(*pop.Connection)

	// find a user with the email
	err := tx.Where("email = ?", strings.ToLower(u.Email)).First(u)

	// helper function to handle bad attempts
	bad := func() error {
		c.Set("user", u)
		verrs := validate.NewErrors()
		verrs.Add("email", "invalid email/password")
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("auth/new.plush.html"))
	}

	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			// couldn't find an user with the supplied email address.
			return bad()
		}
		return errors.WithStack(err)
	}

	// confirm that the given password matches the hashed password from the db
	if u.Authenticate() != true {
		return bad()
	}
	c.Session().Set("current_user_id", u.ID)
	c.Flash().Add("success", "You are logged in.")

	return c.Redirect(302, "/users/%s", u.ID)
}

// AuthDestroy clears the user session
func AuthDestroy(c buffalo.Context) error {
	c.Session().Clear()
	c.Flash().Add("success", "You have been logged out!")
	return c.Redirect(302, "/")
}
