package actions

import (
	"buftester/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
)

// HomeHandler serves up home page.
func HomeHandler(c buffalo.Context) error {
	u := c.Value("current_user")
	if u == nil {
		c.Set("new_user", models.User{})
	} else {
		// Reload user similar to SetCurrentUser middleware.
		// Todo: can we avoid this replay?
		tx := c.Value("tx").(*pop.Connection)
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			err := tx.Eager("Contracts.Boss").Find(u, uid)
			if err != nil {
				return errors.WithStack(err)
			}
			c.Set("current_user", u)
		}
	}
	return c.Render(http.StatusOK, r.HTML("home/index.html"))
}
