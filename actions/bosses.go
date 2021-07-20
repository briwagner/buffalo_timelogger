package actions

import (
	"buftester/models"
	"log"
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

	page := 1
	if c.Param("page") != "" {
		p, err := strconv.Atoi(c.Param("page"))
		if err != nil {
			log.Printf("Cannot parse pager param")
		} else {
			page = p
		}
	}

	q := tx.Paginate(page, 12)
	err := q.All(&bosses)
	if err != nil {
		c.Flash().Add("warning", "No bosses found.")
	}

	// Handle possible empty result from pager.
	if page > 1 && len(bosses) == 0 {
		c.Flash().Add("warning", "No bosses found. Try a lower page number.")
	}

	c.Set("bosses", bosses)
	return c.Render(http.StatusOK, r.HTML("bosses/index.html"))
}

// BossesNew default implementation.
func BossesNew(c buffalo.Context) error {
	c.Set("boss", &models.Boss{})
	return c.Render(http.StatusOK, r.HTML("bosses/new.html"))
}

// BossesCreate responds to POST.
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

	if newContract == false {
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

// BossesShow default implementation.
func BossesShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	boss := &models.Boss{}
	err := tx.Eager("Contracts.User").Find(boss, c.Param("boss_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that boss.")
		return c.Redirect(307, "/")
	}
	c.Set("boss", boss)
	return c.Render(http.StatusOK, r.HTML("bosses/show.html"))
}
