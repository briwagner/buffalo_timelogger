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
		tx := c.Value("tx").(*pop.Connection)

		user := c.Value("current_user").(*models.User)
		err := user.GetContracts(tx)
		if err != nil {
			return errors.WithStack(err)
		}
		c.Set("current_user", user)
	}
	return c.Render(http.StatusOK, r.HTML("home/index.html"))
}
