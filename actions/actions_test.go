package actions

import (
	"testing"

	_ "github.com/gobuffalo/helpers"
	"github.com/gobuffalo/helpers/paths"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite"
	_ "github.com/gobuffalo/tags"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	action, err := suite.NewActionWithFixtures(App(), packr.New("Test_ActionSuite", "../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}

func Test_Helpers(t *testing.T) {
	type Test struct {
		input string
		want  string
	}

	tests := []Test{
		{input: "bosses", want: "/bosses"},
		{input: "newBossesPath()", want: "/bosses/new"},
	}

	for _, tc := range tests {
		p, _ := paths.PathFor(tc.input)
		if p != tc.want {
			t.Errorf("Got %s, want %s", p, tc.want)
		}
	}
}
