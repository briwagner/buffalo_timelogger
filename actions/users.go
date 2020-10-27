package actions

import (
	"buftester/models"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo"
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

// UsersIndex default implementation.
func UsersIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	users := []models.User{}

	err := tx.All(&users)
	if err != nil {
		c.Flash().Add("warning", "No users found.")
	}

	c.Set("users", users)
	return c.Render(http.StatusOK, r.HTML("users/index.html"))
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
	c.Flash().Add("success", "User was created successfully")
	return c.Redirect(303, "/users/%s", u.ID)
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

	c.Set("current_user", user)
	return c.Render(http.StatusOK, r.HTML("users/show.html"))
}

// UsersContractsIndex gets all contracts for the User.
func UsersContractsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	contracts := []models.Contract{}

	// Try to load user first.
	user := &models.User{}
	err := tx.Find(user, c.Param("user_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that user.")
		return c.Redirect(307, "/")
	}

	// Load contracts.
	err = tx.Where("user_id = ?", user.ID).Eager("Boss").All(&contracts)
	if err != nil {
		c.Flash().Add("warning", "No contracts found.")
	}

	c.Set("current_user", user)
	c.Set("contracts", contracts)
	return c.Render(http.StatusOK, r.HTML("users/contracts_index.html"))
}

// UsersContractsNew returns the form for a new contract.
func UsersContractsNew(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	// Try to load user first.
	user := &models.User{}
	err := tx.Find(user, c.Param("user_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that user.")
		return c.Redirect(307, "/")
	}

	// Load bosses for select.
	bosses := []models.Boss{}
	err = tx.All(&bosses)
	if err != nil {
		c.Flash().Add("warning", "No bosses found.")
	}

	// Pass empty struct to form; preset value below.
	contract := &models.Contract{}

	// Check preset boss if passed, i.e. from create-boss route.
	bossID := c.Param("bid")
	if bossID != "" {
		boss := &models.Boss{}
		err = tx.Find(boss, bossID)
		if err != nil {
			fmt.Printf("Cannot find boss %v", err)
		}
		contract.Boss = boss
		contract.BossID = boss.ID
	}

	c.Set("user", user)
	c.Set("contract", contract)
	c.Set("bosses", bosses)
	return c.Render(http.StatusOK, r.HTML("users/contracts_new.html"))
}

// UsersContractCreate responds to POST for form.
func UsersContractCreate(c buffalo.Context) error {
	contract := &models.Contract{}
	if err := c.Bind(contract); err != nil {
		return err
	}

	tx := c.Value("tx").(*pop.Connection)

	// Try to load user first.
	user := &models.User{}
	err := tx.Find(user, c.Param("user_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that user.")
		return UsersContractsNew(c)
	}
	c.Set("user", user)
	contract.UserID = user.ID

	// Try to load boss.
	boss := &models.Boss{}
	err = tx.Find(boss, contract.BossID)
	if err != nil {
		c.Flash().Add("warning", "Cannot find that Employer.")
		return UsersContractsNew(c)
	}

	// Guard against duplicate combo user-boss to prevent duplicates.
	previous := []models.Contract{}
	err = tx.Where("user_id = ? AND boss_id = ?", user.ID, boss.ID).All(&previous)
	if err != nil {
		c.Flash().Add("warning", "Error finding user jobs.")
	}
	if len(previous) > 0 {
		c.Flash().Add("warning", "Contract already exists.")
		return UsersContractsNew(c)
	}

	// Validate the data from the html form.
	verrs, err := tx.ValidateAndCreate(contract)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("contract", contract)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("users/new.html"))
	}

	c.Flash().Add("success", "Contract created.")
	return c.Redirect(303, "/users/%s", user.ID)
}

// UsersContractShow gets a single contract for the User ID.
func UsersContractShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	// Try to load user first.
	user := &models.User{}
	err := tx.Find(user, c.Param("user_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that user.")
		return c.Redirect(307, "/")
	}
	c.Set("current_user", user)

	// Load contract
	contract := &models.Contract{}
	err = tx.Eager().Find(contract, c.Param("contract_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that contract.")
		return c.Redirect(307, "/users/%s", user.ID)
	}
	c.Set("contract", contract)

	sort.SliceStable(contract.Tasks, func(i, j int) bool {
		return contract.Tasks[i].StartTime.Before(contract.Tasks[j].StartTime)
	})

	// Create empty task; set visible dates to now.
	task := &models.Task{}
	task.StartTime = time.Now()
	task.EndTime = time.Now()
	c.Set("task", task)

	return c.Render(http.StatusOK, r.HTML("users/contract_show.html"))
}

// UserTaskCreate responds to POST for new Task.
func UserTaskCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	// Try to load user first.
	user := &models.User{}
	err := tx.Find(user, c.Param("user_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that user.")
		return c.Redirect(307, "/")
	}

	// Load contract
	contract := &models.Contract{}
	err = tx.Eager().Find(contract, c.Param("contract_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that contract.")
		return c.Redirect(307, "/users/%s", user.ID)
	}

	task := &models.Task{}
	if err := c.Bind(task); err != nil {
		return err
	}

	task.ContractID = contract.ID
	// Validate the data from the html form.
	verrs, err := tx.ValidateAndCreate(task)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("task", task)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("users/contract_show.html"))
	}
	c.Flash().Add("success", "Task created successfully")
	return c.Redirect(303, "/users/%s/contracts/%s", user.ID, strconv.Itoa(contract.ID))
}
