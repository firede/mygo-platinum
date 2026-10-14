// Package platinum is a set of controls for MyGo's native UI in the manner
// of the late-1990s gray "Platinum" desktop look: beveled gray buttons,
// striped title bars, pop-up menus and the rest, drawn pixel by pixel on
// the GPU with no web view.
//
// It is a tribute in style only. Its fonts, icons and drawings are its own,
// and it is not affiliated with or endorsed by Apple Inc.
//
// Every control is drawn on a grid of whole DIPs, so it is crisp at any
// whole scale of the display, and its text is drawn with two bitmap faces
// made for the library (see Font). Call Use at the top of a view to give
// the window its theme, as the package's example shows; every control has
// an example of its own.
//
// Controls are built on MyGo's widget bases, so they take the keyboard
// focus, answer the keys their kind does and describe themselves to
// assistive technology. docs/spec.md gives the measurements and colors they
// are drawn with.
//
// # Conventions
//
// Controls bind what they edit by pointer, as MyGo's widgets do:
// Checkbox(c, &app.on, …), ListBox(c, &app.chosen, …). Bind the app's state;
// a variable of the view works too, as the controls carry what the pointer
// does between frames.
//
// Controls that take input return a *…Widget. Its setters (Default,
// Disabled, Width, Label, …) return it, to chain, and come first; its
// queries (Clicked, Changed, Submitted, Activated, …) report what the user
// did since the last frame, and come after. Element returns the control's
// ui.Element, for layout. Controls that only show something (ProgressBar,
// GroupBox, Separator, Placard, Badge, …) return their ui.Element. Dialogs
// and alerts show while a bool they are given is true, and return what the
// user chose.
//
// Each control with a look of its own (not the dialogs made of others, as
// the file dialogs) has a Draw… function too, drawing it in a given state
// with a Pen, for controls of your own that look like it. A Pen draws on whole
// pixels of the grid; Font measures and draws the library's two faces.
package platinum
