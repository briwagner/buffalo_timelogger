package actions

import (
	"buftester/models"
	"net/http"
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
)

// BossesIndex shows all bosses with a param pager.
func BossesIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	bosses := []models.Boss{}

	// Use paginator.

	// Simple way.
	// q := tx.PaginateFromParams(c.Params())

	// Create our own so we can override the per-page.
	// TODO: is there an easier way to override this?
	perPage := 5
	page := 0
	if c.Params().Get("page") != "" {
		p, err := strconv.Atoi(c.Params().Get("page"))
		if err == nil {
			page = p
		}
	}

	q := tx.Paginate(page, perPage)
	q.Paginator.PerPage = 5
	err := q.All(&bosses)
	if err != nil {
		c.Flash().Add("warning", "No bosses found.")
		return c.Redirect(307, "/")
	}

	c.Set("bosses", bosses)
	c.Set("paginator", q.Paginator)
	return c.Render(http.StatusOK, r.HTML("bosses/index.html"))
}

// BossesNew shows the form to create a new Boss.
func BossesNew(c buffalo.Context) error {
	c.Set("boss", &models.Boss{})
	return c.Render(http.StatusOK, r.HTML("bosses/new.html"))
}

// BossesCreate responds to POST to create a new Boss.
func BossesCreate(c buffalo.Context) error {
	boss := &models.Boss{}
	if err := c.Bind(boss); err != nil {
		return err
	}

	newContract := boss.CreateContract

	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form.
	verrs, err := tx.ValidateAndCreate(boss)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("boss", boss)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("bosses/new.html"))
	}
	c.Flash().Add("success", "Boss was created successfully")

	if !newContract {
		return c.Redirect(303, "/bosses/%d", boss.ID)
	}

	uid := c.Session().Get("current_user_id")
	if uid == nil {
		return errors.WithStack(err)
	}

	u := &models.User{}
	err = tx.Find(u, uid)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("warning", "Set a rate for this contract")
	return c.Redirect(303, "/users/%s/contracts/new?bid=%d", u.ID, boss.ID)
}

// BossesShow returns detail for a single Boss.
func BossesShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	boss := models.Boss{}
	err := tx.Find(&boss, c.Param("boss_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that boss.")
		return c.Redirect(307, "/")
	}

	user := c.Value("current_user").(*models.User)

	// Get contracts for this user only.
	cs := models.Contracts{}
	q := tx.Where("user_id = ?", user.ID).Where("boss_id = ?", boss.ID)
	err = q.Eager("User").All(&cs)
	if err != nil {
		c.Flash().Add("warning", "Cannot find contracts.")
		return c.Redirect(307, "/bosses/index")
	}

	boss.Contracts = cs

	c.Set("user", user)
	c.Set("boss", boss)
	return c.Render(http.StatusOK, r.HTML("bosses/show.html"))
}
