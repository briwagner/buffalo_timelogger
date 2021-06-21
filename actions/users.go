package actions

import (
	"buftester/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
)

// SetCurrentUser attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err != nil {
				return errors.WithStack(err)
			}
			c.Set("current_user", u)
		}
		return next(c)
	}
}

// Authorize requires a user be logged in before accessing a route
func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid == nil {
			c.Flash().Add("danger", "You must be authorized to see that page")
			return c.Redirect(302, "/")
		}
		return next(c)
	}
}

// IsOwner checks if the current user has access to route.
func IsOwner(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		// Session().Get() returns interface, so we cast to string.
		if uid := c.Session().Get("current_user_id"); uid != nil {
			pathUserID := c.Param("user_id")
			if pathUserID != fmt.Sprintf("%s", uid) {
				c.Flash().Add("success", "You do not have access to that user.")
				return c.Redirect(302, "/")
			}
		}
		return next(c)
	}
}

// UsersNew returns a form to create new user.
func UsersNew(c buffalo.Context) error {
	c.Set("user", &models.User{})
	return c.Render(http.StatusOK, r.HTML("users/new.html"))
}

// UsersCreate responds to POST.
func UsersCreate(c buffalo.Context) error {
	u := &models.User{}
	if err := c.Bind(u); err != nil {
		return errors.WithStack(err)
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := u.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", u)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		return c.Render(200, r.HTML("users/new.html"))
	}

	// Fire event for new user
	e := events.Event{
		Kind:    "buftester:user:create",
		Message: fmt.Sprintf("New user created: %s", u.Email),
	}
	if err := events.Emit(e); err != nil {
		log.Printf("Failed to emit %v", err)
	}

	c.Flash().Add("success", "New account created. Please log in.")
	// User not logged in yet.
	return c.Redirect(303, "/signin")
}

// UsersUpdate changes stored values for user
func UsersUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	user := &models.User{}
	err := tx.Eager("Contracts.Boss").Find(user, c.Param("user_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that user.")
		return c.Redirect(307, "/")
	}

	c.Request().ParseForm()
	curr := c.Request().Form["CurrentPassword"]
	new := c.Request().Form["NewPassword"]
	if len(curr) == 0 || len(new) == 0 {
		c.Flash().Add("warning", "Form is incomplete.")
		return c.Redirect(302, "/users/%s", user.ID)
	}

	user.Password = curr[0]
	if user.Authenticate() != true {
		c.Flash().Add("warning", "Password does not match the one on record.")
		return c.Redirect(303, "/users/%s", user.ID)
	}

	// Update password on User and generate hash.
	user.Password = new[0]
	_, err = user.UpdatePassword(tx)
	if err != nil {
		c.Flash().Add("warning", "Error saving new password.")
		return c.Redirect(303, "/users/%s", user.ID)
	}

	c.Flash().Add("success", "Password changed.")
	return c.Redirect(303, "/users/%s", user.ID)
}

// UsersShow renders one user.
func UsersShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	err := tx.Eager("Contracts.Boss").Find(user, c.Param("user_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that user.")
		return c.Redirect(307, "/")
	}

	c.Set("user", user)
	return c.Render(http.StatusOK, r.HTML("users/show.html"))
}
