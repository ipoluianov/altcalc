package forms

import (
	"unicode"
)

// EditLine is the text of the expression being typed, with the cursor, the
// selection and the undo. It has no UI, the Display shows it.
type EditLine struct {
	text []rune
	// cur is the cursor, anchor the other end of the selection (= cur when there is none)
	cur, anchor int

	undo, redo []editState
}

type editState struct {
	text        []rune
	cur, anchor int
}

// undoLimit is how many changes Ctrl+Z goes back
const undoLimit = 200

func (e *EditLine) Text() string { return string(e.text) }
func (e *EditLine) Len() int     { return len(e.text) }
func (e *EditLine) Cursor() int  { return e.cur }

// Selection returns the selected range, from <= to; from == to when there is none
func (e *EditLine) Selection() (from, to int) {
	return min(e.cur, e.anchor), max(e.cur, e.anchor)
}

func (e *EditLine) HasSelection() bool { return e.cur != e.anchor }

func (e *EditLine) SelectedText() string {
	from, to := e.Selection()
	return string(e.text[from:to])
}

func (e *EditLine) state() editState {
	return editState{append([]rune(nil), e.text...), e.cur, e.anchor}
}

// change remembers the text for the undo before it is changed
func (e *EditLine) change() {
	e.undo = append(e.undo, e.state())
	if len(e.undo) > undoLimit {
		e.undo = e.undo[1:]
	}
	e.redo = nil
}

func (e *EditLine) Undo() bool {
	if len(e.undo) == 0 {
		return false
	}
	e.redo = append(e.redo, e.state())
	s := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.text, e.cur, e.anchor = s.text, s.cur, s.anchor
	return true
}

func (e *EditLine) Redo() bool {
	if len(e.redo) == 0 {
		return false
	}
	e.undo = append(e.undo, e.state())
	s := e.redo[len(e.redo)-1]
	e.redo = e.redo[:len(e.redo)-1]
	e.text, e.cur, e.anchor = s.text, s.cur, s.anchor
	return true
}

// SetText replaces the text, the cursor at its end; it can be undone
func (e *EditLine) SetText(s string) {
	if s == string(e.text) {
		e.cur, e.anchor = len(e.text), len(e.text)
		return
	}
	e.change()
	e.text = []rune(s)
	e.cur, e.anchor = len(e.text), len(e.text)
}

// Insert puts s in place of the selection, the cursor after it
func (e *EditLine) Insert(s string) {
	e.change()
	from, to := e.Selection()
	rs := []rune(s)
	text := make([]rune, 0, len(e.text)-(to-from)+len(rs))
	text = append(text, e.text[:from]...)
	text = append(text, rs...)
	text = append(text, e.text[to:]...)
	e.text = text
	e.cur = from + len(rs)
	e.anchor = e.cur
}

// Wrap puts before and after around the selection, or around the whole
// text when nothing is selected: "1/(" + text + ")"
func (e *EditLine) Wrap(before, after string) {
	if !e.HasSelection() {
		e.cur, e.anchor = len(e.text), 0
	}
	sel := e.SelectedText()
	e.Insert(before + sel + after)
}

// Backspace deletes the selection or the rune before the cursor; word: the word before it
func (e *EditLine) Backspace(word bool) {
	if !e.HasSelection() {
		if e.cur == 0 {
			return
		}
		e.anchor = e.cur
		if word {
			e.cur = e.wordLeft(e.cur)
		} else {
			e.cur--
		}
	}
	e.Insert("")
}

// Delete deletes the selection or the rune after the cursor; word: the word after it
func (e *EditLine) Delete(word bool) {
	if !e.HasSelection() {
		if e.cur == len(e.text) {
			return
		}
		e.anchor = e.cur
		if word {
			e.cur = e.wordRight(e.cur)
		} else {
			e.cur++
		}
	}
	e.Insert("")
}

// MoveTo puts the cursor at pos; selecting keeps the other end of the selection
func (e *EditLine) MoveTo(pos int, selecting bool) {
	e.cur = max(0, min(len(e.text), pos))
	if !selecting {
		e.anchor = e.cur
	}
}

// Move moves the cursor by a rune or a word to the left (dir < 0) or the
// right. Without selecting, a selection is dropped to its side.
func (e *EditLine) Move(dir int, word, selecting bool) {
	if e.HasSelection() && !selecting && !word {
		from, to := e.Selection()
		if dir < 0 {
			e.MoveTo(from, false)
		} else {
			e.MoveTo(to, false)
		}
		return
	}
	pos := e.cur + dir
	if word {
		if dir < 0 {
			pos = e.wordLeft(e.cur)
		} else {
			pos = e.wordRight(e.cur)
		}
	}
	e.MoveTo(pos, selecting)
}

func (e *EditLine) SelectAll() {
	e.anchor, e.cur = 0, len(e.text)
}

// SelectWordAt selects the number or the name at pos
func (e *EditLine) SelectWordAt(pos int) {
	pos = max(0, min(len(e.text), pos))
	from, to := pos, pos
	for from > 0 && isWordRune(e.text[from-1]) {
		from--
	}
	for to < len(e.text) && isWordRune(e.text[to]) {
		to++
	}
	if from == to && to < len(e.text) {
		to++
	}
	e.anchor, e.cur = from, to
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_'
}

// wordLeft is where the word before pos begins: the spaces, then the
// letters and digits or a single other rune
func (e *EditLine) wordLeft(pos int) int {
	for pos > 0 && unicode.IsSpace(e.text[pos-1]) {
		pos--
	}
	if pos > 0 && !isWordRune(e.text[pos-1]) {
		return pos - 1
	}
	for pos > 0 && isWordRune(e.text[pos-1]) {
		pos--
	}
	return pos
}

func (e *EditLine) wordRight(pos int) int {
	n := len(e.text)
	if pos < n && !isWordRune(e.text[pos]) && !unicode.IsSpace(e.text[pos]) {
		pos++
	} else {
		for pos < n && isWordRune(e.text[pos]) {
			pos++
		}
	}
	for pos < n && unicode.IsSpace(e.text[pos]) {
		pos++
	}
	return pos
}
