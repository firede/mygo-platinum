# Platinum design specification

This is the design this library draws its controls to. It describes a gray,
beveled desktop look in the spirit of the late-1990s "Platinum" appearance,
written from scratch for this library: the numbers below are ours, chosen to
read well on a grid of whole pixels, and the tests in this repository check
the controls against them (`spec_test.go`).

Measurements are in **pixels of the Platinum grid**, which are DIPs: one
device pixel at a scale of 1, a block of 2×2 at a scale of 2. Every line is
one pixel wide and every edge falls on a whole pixel; nothing is
antialiased. Coordinates are from the top left corner of a control's box.

## 1. Color

### 1.1 The gray ramp

Everything is drawn from twelve grays, black and white, plus the accent
(1.2); only pictures — the alert icons, the color picker's palette — take
colors of their own. Each gray has a job; a control that needs a shade uses
the one named for it rather than a new one.

| Name   | Value     | Used for |
|--------|-----------|----------|
| White  | `#FFFFFF` | highlights of raised edges, document backgrounds, text on the accent |
| Gray1  | `#EEEEEE` | inner highlights; the face of disabled bevels |
| Gray2  | `#DDDDDD` | **the face**: windows, dialogs, buttons, menus, the menu bar |
| Gray3  | `#CCCCCC` | unfocused selected text |
| Gray4  | `#BBBBBB` | tracks of scroll bars and sliders, unselected tabs, inactive selection |
| Gray5  | `#AAAAAA` | inner shadows; pressed small controls; default button rings |
| Gray6  | `#999999` | frames of disabled controls; the pressed part of a scroll track |
| Gray7  | `#888888` | **the shadow** of raised edges; etched lines; title bar stripes; placeholder text |
| Gray8  | `#777777` | disabled text; outlines and labels of inactive controls |
| Gray9  | `#666666` | **the face of a pressed button**; menu shadows |
| Gray10 | `#555555` | inner shadow of a pressed button; outlines of bevel buttons; zoom rectangles |
| Gray11 | `#444444` | outer shadow of a pressed button; slider track outline |
| Gray12 | `#222222` | a gray of the color picker's palette |
| Black  | `#000000` | outlines, text |

### 1.2 Accents

An accent is three colors: **Dark** fills highlighted menu items and
selected rows under white text, and draws the focus ring; **Mid** fills
progress bars; **Light** highlights
selected text under black text. The faces of thumbs (scroll bars, sliders)
are halfway between Light and Mid.

| Accent   | Dark      | Mid       | Light     |
|----------|-----------|-----------|-----------|
| Lavender (default) | `#4C4CA8` | `#8888DD` | `#CCCCFF` |
| Graphite | `#4A4A4A` | `#8C8C8C` | `#CCCCCC` |
| Harbor   | `#1F5A8C` | `#5C99CC` | `#BBDDFF` |
| Moss     | `#3A6B2A` | `#77AA5E` | `#CCEEBB` |
| Ember    | `#993D1F` | `#DD8855` | `#FFDDBB` |
| Orchid   | `#7A2E7A` | `#BB77BB` | `#EECCEE` |

## 2. Light and bevels

Light comes from the **top left**. A raised surface is lighter along its top
and left edges and darker along its bottom and right; a sunken one is the
reverse. A bevel ring is one pixel wide and drawn just inside the outline:

- **Raised, outer ring:** top and left White, bottom and right Gray7.
- **Raised, inner ring** (buttons only): top and left Gray1, bottom and
  right Gray5.
- **Sunken ring:** top and left Gray7, bottom and right White.
- The two corners where the colors meet — top right and bottom left — are
  left in the face color, never given to either side.
- **Pressed** inverts a raised control: its face darkens (Gray9 for push
  buttons, Gray4–Gray5 for small controls) and its rings swap to a dark top
  left (Gray11 / Gray10) and a lighter bottom right (Gray7 / Gray8).
- **Etched** lines (group boxes, separators, menu separators) are a Gray7
  line with a White line one pixel below or to the right of it.

Rounded corners are cut in pixel steps. A corner of **radius r** insets its
rows from the side by:

| r | rows from the edge, in order |
|---|------------------------------|
| 1 | 1 |
| 2 | 2, 1 |
| 3 | 3, 1, 1 |
| 4 | 4, 2, 1, 1 |

Where a step jumps by more than one pixel, the outline fills the gap so it
stays connected.

## 3. Type

The two faces are those of the platinum-fonts module (section 10 for their
license), drawn for it; neither copies the shapes of any existing system
font.

| Face | Use | Cell height | Ascent | Cap height | x-height | Strokes |
|------|-----|-------------|--------|------------|----------|---------|
| **System** (Platinum Sans Bold, `SystemFont`) | menus, buttons, labels, titles, tabs | 12 | 9 | 9 | 6 | 2 px vertical stems |
| **App** (Platinum Sans Small Regular, `AppFont`) | documents, text fields, lists, icon names | 11 | 8 | 8 | 5 | 1 px |

One pixel of tracking follows every glyph. Text in a control is placed by
its **capitals**: a label is centered when its cap height is centered in the
control's height (`(h − cap) / 2` from the top, rounded down). Bold text
takes the faces' bold styles, a glyph drawn again a pixel to its right.
The faces cover ASCII, Latin-1, the letters of Latin Extended-A that
Polish, Czech, Slovak, Hungarian and Turkish write most, and typographic
punctuation; a mark over a capital rises at most **2** rows above the cell
into the leading.

East Asian characters draw with the 12-pixel faces of platinum-fonts'
`cjk12` that the app gives (`SetEastAsian`), in the forms of its language:
the bold face for the System face and bold text. An ideograph is **11×11**
of ink, from **1** under the baseline to **10** over it, advancing **12**.

Characters no face has — other scripts — draw in the system's font at 12
(bold) or 11 points on the face's baseline, taking their own width, so that
any text can be read and typed; only where the system has no glyph either
do they draw as a hollow box. They are the one thing not drawn on the pixel
grid. Text never wraps inside a control; labels that do not fit end in an
ellipsis (…).

## 4. States

Every control draws these states where they apply:

- **Normal.**
- **Pressed:** while the pointer is held down inside the control. Dragging
  out of the control shows it normal again; releasing outside does nothing.
- **Disabled:** outline Gray6, label Gray8, bevels flattened (Gray1 / Gray4,
  or none). The control takes no input and is skipped by Tab.
- **Focused:** the keyboard focus is shown by a **focus ring**, two pixels of
  the accent's Dark, one pixel outside the control's outline, its corners
  following the control's (one step rounder). Text fields and lists show it
  whenever they have the focus, as a sign of where typing goes; buttons,
  check boxes, radio buttons and other controls show it only when the focus
  came by the keyboard (as CSS's focus-visible), for full keyboard access and
  assistive technology. The original look had no ring on these controls.
- **Default** (push buttons): a ring around the button (5.1).
- **Selected / checked / on:** per control, below.
- **Inactive:** the controls of a window in the background: outlines and
  labels Gray8 and bevels flat (push buttons, check boxes, radio buttons,
  disclosure triangles, sliders, pop-up menus, bevel buttons, little
  arrows, tabs, text fields' outlines; a calendar, a time field, the icons
  of bevel buttons and a ruler drawn with Black and the darker grays as
  Gray8 and colors as light grays; group box titles, and the headers of
  tables and sheets, likewise, their cells as they are); scroll bars only a
  flat Gray2 track
  in their outline, with no thumb or arrows; selections Gray4. The window's
  own title bar loses its stripes and boxes (5.13). Its first click only
  brings the window to the front.

The look has no hover state: nothing changes as the pointer passes over.

## 5. Controls

### 5.1 Push button

- Height **20**; width the label plus **12** on each side, at least **58**.
- Outline Black with corners of radius **3**; face Gray2; raised outer and
  inner rings (section 2), following the corners.
- Label: System face, Black, centered.
- **Pressed:** face Gray9, rings Gray11/Gray7 and Gray10/Gray8, label White.
- **Disabled:** outline Gray6, rings Gray1/Gray4 (inner ring in the face),
  label Gray8.
- **Default:** a ring **4** pixels wide around the button: Black outline with
  corners of radius 4, a White/Gray7 bevel, a Gray4 face and a Gray5 edge
  against the button. Return and Enter press the default button; it shows
  pressed for **130 ms** when the keyboard presses it.
- **Cancel:** Escape and Command-period press the cancel button.

### 5.2 Check box

- Box **12×12**, label **5** pixels to its right, row height **16**.
- Box: Black square outline, face Gray2, raised ring White/Gray7.
- **Checked:** a cross of two diagonal strokes two pixels thick filling the
  8×8 inside the ring. **Mixed:** a dash 6×2 across the middle.
- **Pressed:** face Gray4, ring Gray7/Gray1.
- **Disabled:** outline Gray6, ring in the face, mark Gray6, label Gray8.
- Space toggles it; a click on the box or its label toggles it.

### 5.3 Radio button

- Circle **12×12** (the pixel circle in `radioMask`), label as a check box.
- Outline Black, face Gray2; the pixels along the outline take White toward
  the top left and Gray7 toward the bottom right (by `x + y` against 11).
- **On:** a black dot 4×4, its corners cut, centered.
- **Focused:** a round ring two pixels wide, a pixel outside the circle.
- Pressed and disabled as the check box. Radio buttons of one choice form a
  group: one stop of Tab, the arrows move the choice within it.

### 5.4 Slider

- Height **18** (23 with tick marks); track **6** tall, from y = 4.
- Track: sunken ring Gray7/White, a Gray11 outline, Gray4 inside, along its
  whole length: the track is never filled on one side of the thumb (a later
  convention); only the thumb carries the accent.
- Thumb **11×16**, pointing down: Black outline, face the thumb face (1.2),
  White top/left edge, the accent's Dark bottom/right edges, three ridges
  (Dark with White below). Pressed: face Mid, edges Dark/Light.
- Tick marks: Gray11 lines 4 tall, two pixels under the thumb's point.
- The arrows, Page Up/Down, Home and End move it.

### 5.5 Progress bar

- Height **14**. A Black outline, the channel Gray4 with Gray7 shadow along
  its top and left and Gray1 along its bottom and right.
- **Determinate:** the done part filled with Mid, its top row Light, its
  second row between Light and Mid, its second-to-last between Mid and Dark,
  its bottom row Dark, and a Dark edge where it ends.
- **Indeterminate:** stripes **8** pixels wide, Mid on Light, leaning right
  (each row one pixel to the left of the row above), moving right at **24**
  pixels a second, under a half-transparent White top row and a Dark bottom
  row.

### 5.6 Disclosure triangle

- Box **12×12**. Pointing right (closed) or down (open): a Black outline,
  Gray6 fill, Gray4 along its upper edge. Its label **5** to its right, row
  **16**.
- **Pressed:** filled Black. **Disabled:** outline Gray6, fill Gray3.
- Click, Space or Enter turns it; Right opens and Left closes.

### 5.7 Group box and separator

- An etched rectangle (Gray7 with White below and right); the title, in the
  System face, interrupts the top edge **8** pixels in, with **3** pixels of
  face on either side. Content inset **12**, below the title **20**.
- A separator is a horizontal etched line: Gray7 over White.

### 5.8 Tab control

- Tabs **22** tall, the selected one **24** and joined to the pane; each as
  wide as its label plus **12** on each side; the first **8** in from the
  pane's edge.
- Tab: Black outline, top corners radius 3, White left and top edge, Gray7
  right edge; face Gray4, the selected tab Gray2. Unselected tabs have a
  Black bottom edge; the selected tab has none.
- Pane: raised (White/Gray7 inside a Black outline), face Gray2, its top edge
  open under the selected tab. Content inset **12**.
- The keyboard focus shows as a bar 2 pixels tall of the accent's Dark under
  the focused tab's label. The arrows, Home and End choose tabs.

### 5.9 Pop-up menu

- Height **20**; as wide as the longest item plus **9** before it, **8** after
  it and the arrow box (**18**).
- Drawn as a push button with a label (5.1); the arrow box is set off by an
  etched line and holds a Black arrow 7×4 pointing down.
- The menu opens over the button with the chosen item on it, at least as wide
  as the button, and checks the chosen item.
- A press, drag and release chooses; a quick click (moving under **3**
  pixels, released within **400 ms**) leaves the menu open for a second
  click. Space, Return, Up and Down open it from the keyboard. A label may
  stand **6** before the button.

### 5.10 Text field and text area

- Text field height **22**. A sunken ring Gray7/White, a Black outline, a
  White face; text in the App face, **5** in from the field's edge, its
  capitals centered.
- Text area: the same frame, lines **14** tall from **4** below the top,
  wrapping at its width, a scroll bar at its right. Each line centers the
  ink of its glyphs — the face's ascent and descent, or as far as East
  Asian glyphs reach (3), where further.
- Caret: a Black line one pixel wide, a pixel short of the line at either
  end, or as far as its glyphs reach where further, blinking every
  **530 ms**. Selection: Light while focused, Gray3 otherwise.
- An input method's composition shows in place of the selection,
  underlined, the caret within it. A password field shows • for each
  character and neither copies nor cuts.
- Command-A, -C, -X, -V, -Z and Shift-Command-Z edit; `EditCommand` gives
  the same commands to an app's Edit menu.
- Placeholder text Gray7. **Disabled:** outline Gray6, face Gray2, text Gray8.
- **Document content** is the exception to the focus ring of text: the text
  area that is a document window's content fills the window to its edges
  in White, with no ring, outline or focus ring of its own (the window in
  front shows where typing goes), its scroll bar against the window's right
  side down to the row of the resize corner, its text starting at the
  ruler's left margin and wrapping at its right one.

### 5.11 List box

- The frame of every content box — list box, tree view, table, icon view,
  scroll view and text area — is the text field's: a sunken ring
  (Gray7/White), a Black outline, a White face. The ring goes round the
  scroll bar too, which shares the content's outline.
- Rows **16** tall in the App face, text **4** in.
- The selected row: the accent's Dark with White text while the list has the
  focus, Gray4 with Black text otherwise.
- Up, Down, Page Up/Down, Home and End move the choice; typing the start of
  an item chooses it; Return and a double click activate.
- A list may give its rows small icons (at most 16×16) **3** in, the text
  **23** in; dim rows in Gray8 that cannot be chosen (the keys and
  clicks pass over them); and keep its scroll bar when the rows fit, as a
  flat track.

### 5.12 Scroll bar

- **16** wide. A Black outline; arrow boxes 16×16 at both ends, raised
  (White/Gray7) on Gray2, with Black arrows 7×4; Black lines between the
  arrows and the track.
- Track: Gray4 with a Gray7 shadow along its top and left and a Gray3 edge on
  its right. The pressed part of the track darkens to Gray6.
- Thumb: as long as the part of the content in view (at least **16**): a
  Black outline, the thumb face, White top/left, the accent's Dark
  bottom/right, four grip ridges across its middle. Pressed: the accent's
  Mid.
- Nothing to scroll: the track is Gray2, the thumb hides and the arrows
  turn Gray6.
- A pressed arrow: face Gray5, ring Gray7/Gray3.
- Arrows and the track repeat after **300 ms**, every **60 ms**, while the
  pointer stays over the part pressed.

### 5.13 Window

- Title bar **19** tall, the frame **6** at the sides and bottom (**8** for
  a movable modal dialog), both including the outline.
- Black outline; a raised ring inside it (White/Gray7); face Gray2. The
  content sits in a sunken ring (Gray7/White) and a Black line.
- Title: System face, centered. **Stripes:** six pairs of rows from y = 4,
  each a Gray7 row with a White row below it offset one pixel right, from the
  close box to the zoom box, broken **6** pixels either side of the title.
- Boxes **11×11** at y = 4: the close box **8** from the left; the collapse
  box **8** from the right and the zoom box **4** left of it. Each is a
  raised button (Black outline, White/Gray7) in a sunken well. The zoom box
  shows a 5×5 square in its top left, the collapse box two horizontal lines.
  Pressed: face Gray6, ring Gray9/Gray4.
- Resize corner **15×15** in the bottom right of the content: Black edges
  on its top and left, three ridges (Gray7 with White) rising to the right.
- **Inactive:** no stripes, no boxes, no ridges, title Gray8, rings softened
  to Gray1/Gray5.
- Dragging the title bar moves the window; double-clicking it, where it has
  a collapse box, rolls it up to its title bar. The zoom box toggles the
  window between the user's size and the app's; the corner resizes it,
  down to **120×80** unless set. Command-W closes it. A click in the
  content of a window in the background only brings it to the front, and
  shows so in the frame of the click; the controls in it take the next
  one, though a drag under way drops in it.
- **Path menu:** a Command-press on the title of a folder's window drops a
  menu from under the title, **16** left of it: the window's folder,
  checked, then each folder it is in, out to the disk; releasing on one, or
  choosing it after a quick click, goes to it.

### 5.14 Menu bar and menus

- Menu bar **20** tall: face Gray2, a White top row and a single Black row
  at its bottom. Menus marked Right sit at its right end, **10** from the
  edge: a clock (a title without a menu, which does not highlight; a click
  on it runs its action, showing the date in place of the time and back)
  and the application menu, its small picture and name side by side **4**
  apart, the picture centered by the rows it draws on. Titles in the System
  face with **9** on each side, the first **10** from the left. An open
  menu's title is highlighted in the accent's Dark with White text.
- A menu: a Black outline, a raised ring (White/Gray7), face Gray2, a shadow
  one pixel deep along its right and bottom. Items **18** tall, titles **18**
  in from the outline, keyboard equivalents right-aligned **12** from the
  edge, separators **8** tall (an etched line at their middle). A check mark
  ✓ sits **6** in from the outline. On a Mac, and in a browser on one,
  equivalents are marked by symbols, as ↑⌘S (Control ^, Option ~, Return
  ↵, Delete ⌫); elsewhere by name, as Ctrl+Shift+S, Command counting as
  Ctrl (`MacKeys`).
- Highlighted item: the accent's Dark, text White. Disabled items Gray8.
- A chosen item blinks for **150 ms** before its action runs. Submenus open
  to the right, overlapping by 3 pixels, marked by a right-pointing triangle.
- Menus open on press and choose on release; a click on a title leaves its
  menu open until the next click. Ctrl-F2 opens the first menu; arrows move
  through items and menus; Return or Space chooses; Escape or Command-period
  closes. A keyboard equivalent blinks its menu's title.

### 5.15 Alert box and dialog

- Alert: **340** wide; a Black outline, a raised ring, a sunken ring and a
  second raised ring (the thick border), padding **15**. A **32×32** icon at
  the top left (note, caution, stop: original drawings), the message in the
  System face, detail in the App face, buttons at the bottom right **10**
  apart, the last the default. It shows centered in the window, with a
  half-transparent Gray9 shadow a pixel deep.
- Movable modal dialog: a window frame (5.13) without boxes or resize
  corner, a thicker side frame, content inset **12**. It first shows
  centered across the window, a fifth of the way down; Escape closes it.

### 5.16 Bevel button, icon button, segmented control

- Bevel button: height **22** (more for a tall icon), its content plus **8**
  on each side; an icon and a label **4** apart, the icon centered by the
  rows it draws on. **Square corners** and a
  **Gray10** (dark gray) outline, two raised rings (White/Gray7,
  Gray1/Gray5), face Gray2. Bevel buttons of the period were square-cornered
  with a dark gray rather than black outline, which sets them apart from
  rounded, black-outlined push buttons; this specification follows that.
- **Pressed:** face Gray6, rings Gray9/Gray4 and Gray8/Gray5; its content
  sinks one pixel right and down. **On** (a toggle): face Gray4, rings
  Gray7/Gray1 and Gray6/Gray3, content sunk. **Disabled:** outline Gray6.
- A bevel button with a menu shows a Black arrow 5×3 at its right and opens
  the menu under itself.
- Icon buttons are square bevel buttons holding an icon alone, three pixels
  round it: **22×22** for the **16×16** icons of toolbars.
- A segmented control joins bevel buttons edge to edge, sharing outlines;
  the chosen segment is on. The focus shows as a 2-pixel accent bar under
  the focused segment's label.

### 5.17 Toolbar

- Height **32**: face Gray2, a White top row, a Gray7 and a White row at
  its bottom (an etched line). Items **4** apart from **6** in.
- Separators are etched vertical lines (Gray7 and White) 22 tall in a
  6-wide slot.
- Items that do not fit move into a menu opened by a **22**-wide » bevel
  button at the toolbar's right end.
- The toolbar is one stop of Tab; the arrows move between its buttons.

### 5.18 Little arrows, number field, search field

- Little arrows: **13×23**, a Black outline with corners of radius 2, split
  at its middle by a Black line into two raised halves holding Black arrows
  5×3. The half pressed: face Gray6, ring Gray9/Gray4. They repeat as the
  scroll bar's arrows do. Up and Down step them from the keyboard.
- Number field: a text field **56** wide and little arrows **3** to its
  right. Up and Down step it from the text too; Return, or leaving the
  field, writes the number tidily.
- Search field: a text field with a Gray9 magnifier 6×7 at its left (text
  moves **14** right) and, once there is text, a Gray7 clear button 7×7 at
  its right. Escape clears it.

### 5.19 Status controls

- Chasing arrows: **16×16**, two Black arcs (radius 4.3–6.7) with
  arrowheads, chasing each other round in **8** steps, **10** a second.
- Placard: height **16**, a raised band — face Gray2, White along its top
  and left, Gray7 along its bottom and right — with **no** outline; text in
  the App face **6** in.
- Badge: height **14**, a square box, inverted: Black with White text in the
  App face, at least as wide as tall, **4** of room either side, as the
  desktop marks what is selected. Tag: a square box of Gray1 with a Black
  outline and Black text. (Rounded "pills" belong to later desktops.)
- Image well: a sunken ring (Gray7/White), a Black outline and a White face,
  48×48 unless sized.

### 5.20 Help balloon and context menu

- Help balloon: a White balloon with corners of radius 4, a Black outline
  and a half-transparent Gray9 shadow one pixel down and right; a tail **8**
  tall at its top left, its left edge upright, rising to the anchor. Text in
  the App face, **10** in, wrapping at **200**. It shows while the pointer
  rests on its anchor, or the keyboard focus is on it.
- Context menu: a menu (5.14) opening at the pointer on a secondary click or
  a Control-press, and from the keyboard with Shift-F10.

### 5.21 Tree view, table, icon view

- Tree view: as a list box (5.11), each level indented **16**; a disclosure
  triangle (5.6) **3** in from the level's start, an optional icon, the
  title **5** after the triangle. The selected row's highlight starts after
  its triangle. Right opens a row, or goes into it; Left closes it, or
  goes to its parent.
- Table: column headers **18** tall, raised (White/Gray7) on Gray2 with a
  Black line under them, titles in the App face; the sorted column's header
  Gray4 with a Black sort arrow 5×3 at its right (up ascending, down
  descending), and the column shaded Gray1 all the way down. Rows **16**,
  cells **5** in, a column's small icon (at most 16×16) **4** in from the
  left of its cells with the text **4** after a 16-wide icon, right-aligned columns flush **5**
  from their right edge, Gray3 lines between columns. Rows are not striped
  (a later convention). A table may keep its scroll bar when its rows fit,
  as a flat track, and may draw no focus ring, as a list view filling a
  window. Columns resize from **3** pixels either side of their
  right edge, down to **30** wide. Cells sort as numbers where both
  are numbers and as text otherwise, unless the column orders its rows
  itself. A click on a header sorts by its column, a second the other way.
  Shift-click and Shift with the arrows extend the selection, Command-click
  toggles a row, Command-A selects all.
- Icon view: cells **84×64**; the icon centered **4** down, the name under it
  **38** down on a White plate. A selected icon darkens by half and its name
  inverts (White on the accent's Dark while focused, on Black otherwise).
  Rows start **4** down and leave **4** under the last. When they do not
  all fit, a scroll bar (5.12) stands at the right and the columns fill
  what is left. The arrows, Page Up/Down, Home and End move the choice and
  scroll it into view.

### 5.22 Split, scroll view, utility window

- Split divider: **7** wide (or tall), face Gray2 between Black edges, a
  White line inside the first edge and a Gray7 one inside the second; a grip
  of five dots (Gray7 with White) 4 apart at its middle. Each pane keeps at
  least **40**; the arrows move the divider by **8**.
- Scroll view: the content frame (5.11), the content **6** in from the
  sides and **4** from the top, a scroll bar (5.12) at its right.
- Utility window: a title bar **13** tall with four pairs of stripes from
  y = 2, its title in the App face, a close box (and collapse box) **9×9**
  at y = 2, **4** from the sides; a frame **3** wide and a Black line round
  the content; no zoom box or resize corner.

### 5.23 Calendar, date picker, time field, combo box, color well

- Calendar: days **22×16** in seven columns, inside the content frame (a
  sunken ring, a Black outline, White). A raised header **20** tall holds the
  month and year in the System face between Black arrows 4×7 that turn the
  month; the weekdays' initials in the App face, Gray8, over a Gray4 line.
  Days of other months are Gray6; the chosen day takes the selection's
  colors (the accent's Dark while focused, Gray4 otherwise); today has a
  Black outline. Arrows move by a day or a week, Page Up/Down by a month.
  Weeks start on the user's first weekday; the month, the year and the
  weekdays are written in the window's language (9).
- Date picker: a pop-up button (5.9) showing the date as the window's
  language writes it short, as wide as the widest date. Space, Return or
  Down opens a calendar two pixels under it; choosing a day, Return,
  Escape or a click elsewhere closes it.
- Time field: one text field (5.10), as wide as the language's widest time
  plus **18**, with one pair of little arrows **3** to its right. It counts
  the hours as the user's regional formats do, as "9:41 AM" or "21:41",
  its parts in the order and with the separators of the window's language
  (the half of the day first where the language puts it first). Its parts
  are chosen one at a time, by a click or Left and Right, the part chosen
  highlighted in the accent's Light while the field has the focus; Up,
  Down and the arrows step that part, digits type it, A and P set the half
  of the day.
- Combo box: a text field with a **20×20** bevel button holding the pop-up
  arrow joined to its right, within the field's ring like its outline; the
  menu under it lists the items that begin as the text does.
- Color well: a bevel button **44×22** holding a swatch in a Black outline,
  **5** in. It opens a color picker: a movable modal dialog with a palette of
  eight hues in five shades and a row of grays (swatches **14** square in a
  well, the chosen one outlined Black and White), the new color over the old
  in a well, and sliders of red, green and blue.

### 5.24 Sheet, formula bar, find and replace, ruler

- Sheet: cells **72×18**, row numbers **32** wide, in the content frame.
  Headers in the App face on raised Gray2 (White/Gray7, Gray9 lines between
  them), the headers of the selected rows and columns Gray4. Gray3 lines
  between cells; numbers right-aligned **4** from the edge, text left. The
  selected range is shaded the accent's Light (Gray3 unfocused); the cursor
  is a frame two pixels wide of the accent's Dark (Black unfocused).
  Formulas begin with "=" and take numbers, cells, + − × ÷, parentheses and
  SUM, AVERAGE, MIN and MAX of ranges; errors show as #DIV/0!, #CYCLE!,
  #NAME?, #REF! and #SYNTAX!. The arrows move the cursor and Shift with
  them extends the range; typing replaces the cell, Escape cancels the
  edit, Delete clears the range. Return commits and moves down, Tab right;
  Tab from the last column (Shift-Tab from the first) leaves the sheet for
  the next control, so that the keyboard is never trapped in it.
- Formula bar: a band of Gray2 with an etched line under it; the cursor's
  name on a placard **44** wide, and its text in a text field.
- Find and replace: a movable modal dialog **380** wide with Find and
  Replace fields under labels **62** wide, Match case and Whole word, a
  Gray9 note, and a Cancel button at the left of Replace All, Replace and
  Find (default).
- Ruler: **22** tall on Gray2 with a Black line at its bottom; a scale band
  **12** tall, White between the margins and Gray1 outside, ruled Gray7 above
  and below. The scale is in inches at 72 dpi and starts at the left margin:
  ticks from its bottom every **9** (2 tall), every 36 (5 tall) and every 72
  (the band's height), numbered right of the tallest; Black margin markers
  under the band, and on the band's top a Black triangle **5** wide pointing
  down, the first-line indent; all three dragged in steps of 9, the right
  margin and the indent from the left margin. The first line of each
  paragraph starts the indent further in.

### 5.25 Progress dialog, banner, toasts, empty state, paths, pages, card, status bar

- Progress dialog: a movable modal dialog **320** wide: the message, a
  progress bar **290** long, a Gray9 detail, and a Cancel button.
- Banner: the alert frame (5.15) **30** tall, laid out where the app
  puts it — across the top of a window's content, as a rule, the content
  under it moved down — as wide as its parent: a 16×16 icon (small note,
  caution or stop), text in the App face, and a close box (5.13).
- Toasts: MyGo's toasts (`c.Toast`), drawn by `Toasts` as banners as wide
  as their text, centered **12** above the bottom of the window, **6**
  apart, the newest lowest; an action as a push button (5.1) before the
  close box. They show and go at once, without fading.
- Empty state: one line of Gray8 text in the App face, centered in its
  room — no picture.
- Path pop-up: a bevel button with a small folder and the folder's name,
  whose menu lists the folder and those it is in, outward.
- *Extras.* Breadcrumbs and pagination belong to later, web-era designs
  and have no Platinum counterpart; the library draws them in its
  vocabulary for apps that need them, but the gallery and the desktop's
  windows use the path pop-up and scrolling instead. Breadcrumbs: the
  folders' names in the System face, Gray9 but the last (Black), with Gray8
  arrows 3×5 between them; pressed, a name inverts on the accent's Dark.
  Pagination: bevel buttons **22** square with arrows, "Page n of m"
  between them.
- Card: a Black outline round a raised Gray2 face, content inset **10**.
- Status bar: placards (5.19) side by side, the last taking the room left.

### 5.26 Column view, document tabs, font picker, page setup

- Column view: list boxes (5.11) **140** wide side by side, sharing their
  edges; rows that hold others end in a Black arrow 4×7, **11** from the
  list's right edge (White on the focused selection). The keyboard focus
  is one ring, two pixels of Accent.Dark, inside the whole of the columns,
  not around a column. Under them a horizontal scroll bar (5.12) scrolls
  them, its arrows dimmed while they fit; the newest column scrolls into
  view when the path changes.
- Document tabs: tabs (5.8) **4** from the pane's edge, titled as tab
  controls' are; a document closes with Command-W while the tabs have the
  focus, or its tab's context menu, not a box on the tab. The pane's
  content inset **10**; with no documents, an empty state (5.25).
- Font picker: a list of families **140×98**, a size field and list
  **76** wide, the styles as check boxes, and a sample in a well **40**
  tall drawn in the regular face styled as rich text is (5.27); sizes of
  18 and over draw at whole multiples, plain or underlined.
- Page setup: a movable modal dialog **360** wide; the paper pop-up,
  orientation as two icon bevel buttons (portrait, landscape) that toggle,
  scale and margin number fields, and a picture of the page: White with a
  Black outline and a Gray7 shadow 2 down and right, margins dotted Gray5,
  lines of text Gray4.

### 5.27 Rich text and the format bar

- Rich text area: a text area (5.10) whose runs take styles in the same
  regular face, as classic bitmap styled text: **bold** draws each glyph again a
  pixel to its right, a pixel wider; *italic* slants each pair of rows a
  pixel right of the pair below (the top row leaning half the ascent, 4),
  and its run ends with that much more room, so the next run does not touch
  it; underline is a line one pixel under the baseline. East Asian glyphs
  take their bold face and stay upright in italic, as do characters in the
  system's font, at their own size. Each paragraph
  aligns its lines left, centered or right within the wrapping width; a
  return makes a paragraph aligned as the one it splits, and joining two
  keeps the first one's alignment. Text comes in four sizes, 11, 22, 33 and
  44 points — the face at one to four times, each of its pixels drawn k by
  k, as a bitmap font was drawn at a size it had no bitmaps for; a line is
  as tall as its largest text needs (14 plus, for each step, the ink of a
  line: 11, or 13 with East Asian glyphs), and all its runs share that
  text's baseline. Command-B, -I and -U toggle the styles; Command-Shift-.
  and -, step the size; a Size menu, its items named in the app's language
  ("22 Point"), checks the selection's.
- Format bar: a toolbar (5.17) of icon buttons — bold, italic and underline
  toggles that show the selection's style, and three alignment toggles
  that show and set the alignment of the paragraphs the selection touches.

### 5.28 Pop-up window

- A window docked at the bottom of the desktop, as the classic pop-up
  windows: a **tab** on its top edge, **20** tall, its title in the System
  face centered, **16** from either side (80 wide at least), **12** from
  the window's left. The tab is the title bar's Gray2 with the frame's
  bevel, Black outline with the top corners cut by a pixel; held down, its
  face is Gray5 with the bevel reversed. In the background its title is
  Gray8 and its bevel softens to Gray1/Gray5.
- Open, the window rises from the bottom edge under the tab, the tab's face
  running into its frame: the frame of a window (5.13) without a title bar,
  **6** wide, its bottom and the sunken edge of its content running off the
  edge. Closed, the tab alone stands on the bottom edge.
- A click on the tab, or Space or Enter while it has the keyboard focus
  (an Accent.Dark bar under the title), opens and closes the window;
  dragging the tab sideways moves it, and dragging it up more than **24**
  tears the window off, to show as a window again. Command-W closes it.
- A drag that rests on a closed tab for **500 ms** springs the window open,
  for what is dragged to be dropped in it. A window in the background
  takes a drop in its content while a drag is under way, keeping its look;
  only a click there first brings it to the front. The app may take drops
  on the tab itself (Tab), as a classic desktop put an item dropped on a
  closed tab in its folder.

### 5.29 Open and save dialogs

- The classic desktop's standard file dialogs: modal boxes with no title
  bar, in the alert's raised double frame, centered a fifth of the way
  down. At the top, the folder shown as a bevel button with its menu of
  the folders it is in, out to the desktop, with the small icon of what it
  is (the desktop's, a disk's, a folder's).
- Under it a list **220×150** of what the folder holds, each row with its
  **16×16** icon (folder, document or disk) and its name **19** after it,
  and a scroll bar **16** wide at its right always, a flat track when the
  rows fit.
- Beside the list, a column of buttons **100** wide, **12** from it: the
  icon and name of the disk shown; Eject (dimmed: nothing here ejects),
  Desktop, and in the save dialog New Folder; a Gray4 line; and at the
  bottom Cancel and the default button, Open or Save, its face lined up
  with the buttons above and its ring outside the column.
- Opening a folder — a double click, Return, or Open — goes into it, and
  Command-Up goes out of it; Escape cancels. The open dialog gives the
  file opened.
- The save dialog shows its files in Gray8, not to be chosen; under the
  list its prompt, which the app gives ("Save as:"), and the name's field
  **220** wide, the name
  selected and holding the keyboard so that typing replaces it. Saving
  over a file of the name asks first, in a caution alert: Replace existing
  “name”?

### 5.30 Window header

- A band **18** tall across the top of a window's content, telling about
  it (a folder's count of items and the room left on its disk): Gray2,
  with a Gray7 line and a Black one at its bottom; its texts in the App
  face, **8** from the left and **8** from the right, centered in the
  band over the lines.

## 6. Layout

- Controls in a column are **6–8** apart; groups of controls **12–16**.
- Dialog buttons sit at the bottom right, the default rightmost, **10**
  apart; Cancel to the left of the default.
- Labels before a control end with a colon and sit **6** before it.

## 7. Motion

The look is quiet: there are no fades or slides. What moves is functional —
a window opening from an icon is preceded by **zoom rectangles**: four
dotted Gray10 outlines growing from the icon's box to the window's over
**180 ms**, each trailing the one before by **0.12** of the way;
a chosen menu item blinks (150 ms), a keyboard-pressed button flashes
(130 ms), the caret blinks (530 ms), indeterminate progress stripes move
(24 px/s), and arrows of scroll bars repeat. A drag that rests **500 ms**
on a folder, or on a pop-up window's closed tab, springs it open. With
reduced motion asked for, zoom rectangles do not show; the stripes still
move (they show that work goes on).

Nothing lags a frame behind the input that changed it. A control reporting
what the user did — a change, a window activated or closed, a menu item
chosen — has MyGo build the view again in that frame, for the app's answer
to show with it; a control painted at a new size, and a modal layer the
first time it shows, ask for one frame more.

## 8. Accessibility

Every control is built on MyGo's widget bases or gives its element a role,
a name (its label) and its states (checked, expanded, selected, value,
range), so screen readers read it as the native control it looks like.
Controls drawn in one element (lists, menus) also build invisible elements
for their rows and items, over where they show. Menu titles and menu items
take the press action of assistive technology: a title opens its menu as
the keyboard does, an item is chosen.

Every stop of the keyboard focus has a name: a control made of parts (a
number field and its little arrows, a combo box and its button) names each
part, in the window's language (9). Tab goes round every window without
being kept: a control that takes Tab for itself (a sheet, moving between
cells) lets it go on from its last cell, Shift-Tab from its first. News
the keyboard focus does not bring — a banner, the note of find and
replace, what a search found — is read out once. Tests tab round the
gallery and the desktop to hold all of this.

## 9. Language

The controls' own words — Cancel, OK, Open, Save, the names screen readers
read — are in English, Simplified and Traditional Chinese, Japanese and
Korean, in the language of the window they show in (menus, built without a
window: the app's). An app changes them, or adds languages, with keys
`platinum.<key>` in its own catalogs. Dates and times follow the user's
regional formats: the first day of the week, the 12- or 24-hour clock, and
the order of hours, minutes and the half of the day; where the system
says nothing of them, as in tests, those of the language. The look is
written left to right.

## 10. Assets

Every icon and every picture in this repository was drawn for it and is
under its MIT license. The faces are platinum-fonts', under the SIL Open
Font License 1.1: its Latin faces drawn for it, its East Asian glyphs from
the open fonts it lists. No vendor's font, icon or image is copied or
traced; the system's font draws, at run time, only the characters no face
has (3).
