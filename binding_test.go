package platinum

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// TestLocalCopyBindings binds every control that edits a value to a copy
// the view makes anew each build, from the app's state it writes back
// after the control, and checks that the user's input reaches the app.
func TestLocalCopyBindings(t *testing.T) {
	freeze(t)
	items := []string{"Alpha", "Beta", "Gamma"}
	day := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	var app any
	keys := func(ks ...ui.Key) func(*testing.T, *ui.Tester) {
		return func(t *testing.T, tt *ui.Tester) {
			tt.Key(0, ui.KeyTab)
			tt.Frame()
			for _, k := range ks {
				tt.Key(0, k)
				tt.Frame()
			}
		}
	}
	typing := func(t *testing.T, tt *ui.Tester) {
		keys()(t, tt)
		tt.Type("hi")
		tt.Frame()
	}
	click := func(labels ...string) func(*testing.T, *ui.Tester) {
		return func(t *testing.T, tt *ui.Tester) {
			for _, l := range labels {
				if err := tt.Click(l); err != nil {
					t.Fatal(err)
				}
				tt.Frame()
			}
		}
	}
	for _, tc := range []struct {
		name  string
		build func(c *ui.Context)
		act   func(*testing.T, *ui.Tester)
		zero  any
	}{
		{"Checkbox", func(c *ui.Context) { v := app.(bool); Checkbox(c, &v, "Sound"); app = v }, click("Sound"), false},
		{"Radio", func(c *ui.Context) {
			v := app.(int)
			RadioGroup(c, func() {
				Radio(c, &v, 0, "Alpha")
				Radio(c, &v, 1, "Beta")
			})
			app = v
		}, click("Beta"), 0},
		{"Disclosure", func(c *ui.Context) { v := app.(bool); Disclosure(c, &v, "More"); app = v }, click("More"), false},
		{"Segmented", func(c *ui.Context) { v := app.(int); Segmented(c, &v, items...); app = v }, click("Beta"), 0},
		{"TabControl", func(c *ui.Context) { v := app.(int); TabControl(c, &v, items, func(int) {}); app = v }, click("Beta"), 0},
		{"DocumentTabs", func(c *ui.Context) { v := app.(int); DocumentTabs(c, &v, items, func(int) {}); app = v }, click("Beta"), 0},
		{"PopupMenu", func(c *ui.Context) { v := app.(int); PopupMenu(c, &v, "Sort:", items); app = v }, func(t *testing.T, tt *ui.Tester) {
			r, _ := tt.Find("Sort:")
			tt.Press(r.X+r.W-10, r.Y+10)
			tt.Frame()
			g, ok := tt.Find("Gamma")
			if !ok {
				t.Fatalf("menu not open: %q", tt.Texts())
			}
			tt.Move(g.X+10, g.Y+5)
			tt.Release(g.X+10, g.Y+5)
			tt.Frame()
		}, 0},
		{"Slider", func(c *ui.Context) { v := app.(float64); Slider(c, &v, 0, 10); app = v }, keys(ui.KeyRight), 0.0},
		{"Stepper", func(c *ui.Context) { v := app.(float64); Stepper(c, &v, 0, 10, 1); app = v }, keys(ui.KeyUp), 0.0},
		{"NumberField", func(c *ui.Context) { v := app.(float64); NumberField(c, &v, 0, 10, 1, 0); app = v }, keys(ui.KeyUp), 0.0},
		{"TextField", func(c *ui.Context) { v := app.(string); TextField(c, &v); app = v }, typing, ""},
		{"TextArea", func(c *ui.Context) { v := app.(string); TextArea(c, &v); app = v }, typing, ""},
		{"SearchField", func(c *ui.Context) { v := app.(string); SearchField(c, &v); app = v }, typing, ""},
		{"Combobox", func(c *ui.Context) { v := app.(string); Combobox(c, &v, items); app = v }, typing, ""},
		{"ListBox", func(c *ui.Context) { v := app.(int); ListBox(c, &v, items); app = v }, click("Beta"), 0},
		{"IconView", func(c *ui.Context) {
			v := app.(int)
			IconView(c, []IconItem{{Name: "Alpha"}, {Name: "Beta"}}, &v)
			app = v
		}, click("Beta"), -1},
		{"Calendar", func(c *ui.Context) { v := app.(time.Time); Calendar(c, &v); app = v }, keys(ui.KeyRight), day},
		{"ColorWell", func(c *ui.Context) { v := app.(ui.Color); ColorWell(c, &v).Label("Tint"); app = v }, click("Tint", "OK"), ui.Color{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app = tc.zero
			view := func(c *ui.Context) {
				Use(c, nil)
				ui.Column(c).Padding(10).Children(func() { tc.build(c) })
			}
			tt := ui.NewTester(view, 500, 360)
			tc.act(t, tt)
			tt.Frame()
			if fmt.Sprint(app) == fmt.Sprint(tc.zero) {
				t.Errorf("the input did not reach the app: still %v", app)
			}
		})
	}
}
