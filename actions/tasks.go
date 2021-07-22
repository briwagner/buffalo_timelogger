package actions

import (
	"buftester/models"
	"log"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
)

// TasksShow default implementation.
func TasksShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	err := tx.Find(task, c.Param("task_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that task.")
		return c.Redirect(307, "/")
	}
	c.Set("task", task)
	return c.Render(http.StatusOK, r.HTML("tasks/show.html"))
}

// TasksEdit shows the edit form.
func TasksEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	err := tx.Eager().Find(task, c.Param("task_id"))
	if err != nil {
		c.Flash().Add("warning", "Cannot find that task.")
		return c.Redirect(307, "/")
	}
	c.Set("task", task)
	return c.Render(http.StatusOK, r.HTML("tasks/edit.html"))
}

// TasksUpdate responds to POST.
func TasksUpdate(c buffalo.Context) error {
	task := &models.Task{}

	tx := c.Value("tx").(*pop.Connection)
	err := tx.Eager().Find(task, c.Param("task_id"))
	if err != nil {
		log.Println(err)
		c.Flash().Add("warning", "Cannot find that task.")
		return c.Redirect(307, "/")
	}

	// Bind entity to the HTML form.
	if err := c.Bind(task); err != nil {
		return err
	}

	task.UpdatedAt = time.Now()

	// Validate the data from the html form.
	verrs, err := tx.ValidateAndUpdate(task)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("task", task)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("tasks/edit.html"))
	}

	c.Flash().Add("success", "Task updated.")
	return c.Redirect(303, "/users/%s/contracts/%d", task.Contract.UserID, task.Contract.ID)
}
