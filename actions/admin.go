package actions

import (
	"buftester/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
)

// IsAdmin is middleware to enforce user role
func IsAdmin(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		uid := c.Session().Get("current_user_id")
		if uid == nil {
			c.Flash().Add("danger", "You must be logged in to see that page")
			return c.Redirect(302, "/")
		}

		if uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err != nil {
				return errors.WithStack(err)
			}
			c.Set("current_user", u)
			if u.Roles != "admin" {
				c.Flash().Add("danger", "You are not authorized to view that page")
				return c.Redirect(302, "/")
			}
		}
		return next(c)
	}
}

// AdminUsersIndex shows all users.
func AdminUsersIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	users := []models.User{}

	err := tx.All(&users)
	if err != nil {
		c.Flash().Add("warning", "No users found.")
	}

	c.Set("users", users)
	return c.Render(http.StatusOK, r.HTML("users/index.html"))
}

// AdminUserShow shows details for one user.
func AdminUserShow(c buffalo.Context) error {
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

// AdminUserUpdate handles post to modify user account.
func AdminUserUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	err := tx.Eager("Contracts.Boss").Find(user, c.Param("user_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that user.")
		return c.Redirect(307, "/")
	}

	c.Request().ParseForm()
	v := c.Request().Form["AdminRole"]

	// Checkbox field is totally empty if trying to unset the value.
	if len(v) == 0 {
		user.SetRole("")
	} else if v[0] == "set_admin" {
		user.SetRole("admin")
	}
	tx.Update(user)

	return c.Redirect(303, "/admin/users/%s", user.ID)
}
