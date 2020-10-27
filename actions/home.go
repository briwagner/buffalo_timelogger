package actions

import (
	"buftester/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// HomeHandler serves up home page.
func HomeHandler(c buffalo.Context) error {
	u := c.Value("current_user")
	if u == nil {
		c.Set("new_user", models.User{})
	}
	return c.Render(http.StatusOK, r.HTML("home/index.html"))
}
