package actions

import (
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/plush"
)

var r *render.Engine
var assetsBox = packr.New("app:assets", "../public")

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.plush.html",

		// Box containing all of the templates:
		TemplatesBox: packr.New("app:templates", "../templates"),
		AssetsBox:    assetsBox,

		// Add template helpers here:
		Helpers: render.Helpers{
			// for non-bootstrap form helpers uncomment the lines
			// below and import "github.com/gobuffalo/helpers/forms"
			// forms.FormKey:     forms.Form,
			// forms.FormForKey:  forms.FormFor,
			"isActiveNav": func(m string, cp string) string {
				if m == cp {
					return "nav-link active"
				}
				return "nav-link"
			},
			"isLoggedIn": func(help plush.HelperContext) bool {
				if _, ok := help.Value("current_user_id").(string); ok {
					return true
				}
				return false
			},
		},
	})
}
