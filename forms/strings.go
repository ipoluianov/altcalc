package forms

import (
	"errors"

	"github.com/ipoluianov/altcalc/calc"
	"github.com/ipoluianov/nui/ui"
	"github.com/ipoluianov/nui/ui/i18n"
)

// Strings are all the texts of the application. A misspelled field is a
// compile error; a field a language leaves empty falls back to English
// (strings_test.go checks that none is left, the maps included).
type Strings struct {
	Error string

	// Menus; & marks the letter of Alt+letter
	MenuEdit   string
	MenuView   string
	MenuMemory string
	MenuHelp   string

	// Commands
	CmdUndo            string
	CmdRedo            string
	CmdCut             string
	CmdCopy            string
	CmdCopyCalculation string
	CmdPaste           string
	CmdSelectAll       string
	CmdClear           string
	CmdClearHistory    string
	CmdMemoryClear     string
	CmdMemoryRecall    string
	CmdMemoryAdd       string
	CmdMemorySubtract  string
	CmdMemoryStore     string
	CmdStandard        string
	CmdScientific      string
	CmdHistory         string
	CmdAlwaysOnTop     string
	CmdNextAngle       string
	CmdReference       string
	Angles             [3]string // by calc.AngleMode

	// The display
	DisplayHint string
	Copied      string

	// The errors of the expressions
	ErrSyntax      string
	ErrUnknownName func(name string) string
	ErrDivByZero   string
	ErrDomain      string
	ErrOverflow    string
	ErrArgCount    func(name string) string
	ErrNotInteger  string
	ErrReadOnly    func(name string) string
	ErrUnmatched   string

	// Hints of the keys, by the key id
	KeyHints map[string]string

	// The history panel
	PanelHistory     string
	PanelVariables   string
	HistoryEmpty     string
	ClearLink        string
	InsertResult     string
	InsertExpression string
	CopyResult       string
	Delete           string
	ClearHistoryAsk  string

	// The status bar
	StatusMemory func(value string) string

	// Settings, about, install
	Settings         string
	Help             string
	About            string
	Language         string
	LanguageSystem   string
	Theme            string
	ThemeDark        string
	ThemeLight       string
	DecimalSeparator string
	DecimalPoint     string
	DecimalComma     string
	Grouping         string
	AboutTitle       func(app string) string
	Version          string
	Author           string
	License          string
	VisitWebsite     string
	Close            string
	Install          string
	Update           string
	Uninstall        string
	InstallAsk       func(dir string) string
	InstallFailed    func(err string) string
	Installed        string
	UninstallAsk     func(dir string) string
	Uninstalled      string
	UninstallRunning string

	// The quick reference: what can be typed
	ReferenceTitle string
	Reference      string
}

// ErrorText is the message of an error of an expression
func (s *Strings) ErrorText(err error) string {
	var e *calc.Error
	if !errors.As(err, &e) {
		return err.Error()
	}
	switch e.Code {
	case calc.ErrUnknownName:
		return s.ErrUnknownName(e.Name)
	case calc.ErrDivByZero:
		return s.ErrDivByZero
	case calc.ErrDomain:
		return s.ErrDomain
	case calc.ErrOverflow:
		return s.ErrOverflow
	case calc.ErrArgCount:
		return s.ErrArgCount(e.Name)
	case calc.ErrNotInteger:
		return s.ErrNotInteger
	case calc.ErrReadOnly:
		return s.ErrReadOnly(e.Name)
	case calc.ErrUnmatched:
		return s.ErrUnmatched
	}
	return s.ErrSyntax
}

// KeyHint returns the hint of the key, "" when it has none
func (s *Strings) KeyHint(id string) string {
	if v, ok := s.KeyHints[id]; ok {
		return v
	}
	return en.KeyHints[id]
}

var languages = []struct{ tag, name string }{
	{"en", "English"},
	{"ru", "Русский"},
}

var en = Strings{
	Error: "Error",

	MenuEdit:   "&Edit",
	MenuView:   "&View",
	MenuMemory: "&Memory",
	MenuHelp:   "&Help",

	CmdUndo:            "Undo",
	CmdRedo:            "Redo",
	CmdCut:             "Cut",
	CmdCopy:            "Copy",
	CmdCopyCalculation: "Copy Calculation",
	CmdPaste:           "Paste",
	CmdSelectAll:       "Select All",
	CmdClear:           "Clear",
	CmdClearHistory:    "Clear History",
	CmdMemoryClear:     "Memory Clear",
	CmdMemoryRecall:    "Memory Recall",
	CmdMemoryAdd:       "Memory Add",
	CmdMemorySubtract:  "Memory Subtract",
	CmdMemoryStore:     "Memory Store",
	CmdStandard:        "Standard",
	CmdScientific:      "Scientific",
	CmdHistory:         "History",
	CmdAlwaysOnTop:     "Always on Top",
	CmdNextAngle:       "Next Angle Unit",
	CmdReference:       "Quick Reference",
	Angles:             [3]string{"Degrees", "Radians", "Gradians"},

	DisplayHint: "Type an expression",
	Copied:      "Copied",

	ErrSyntax:      "Syntax error",
	ErrUnknownName: func(name string) string { return "Unknown name: " + name },
	ErrDivByZero:   "Division by zero",
	ErrDomain:      "Not defined for this value",
	ErrOverflow:    "The number is too large",
	ErrArgCount:    func(name string) string { return "Wrong number of arguments of " + name },
	ErrNotInteger:  "Integers only",
	ErrReadOnly:    func(name string) string { return name + " cannot be assigned" },
	ErrUnmatched:   "Unmatched parenthesis",

	KeyHints: map[string]string{
		"2nd":   "The inverse functions on the keys",
		"angle": "The unit of the angles: degrees, radians, gradians (F2)",
		"deg":   "Degrees in any unit: sin(90°)",
		"mod":   "Remainder: 7 mod 3 = 1",
		"ee":    "Exponent: 1.5e3 = 1500",
		"ans":   "The last result",
		"rand":  "Random number from 0 to 1",
		"abs":   "Absolute value",
		"inv":   "1/x of the expression",
		"neg":   "Change the sign (F9)",
		"fact":  "Factorial: 5! = 120",
		"comma": "Separates the arguments: max(1, 2)",
		"left":  "Cursor left",
		"right": "Cursor right",
		"mc":    "Memory clear (Ctrl+L)",
		"mr":    "Memory recall (Ctrl+R)",
		"m+":    "Add to memory (Ctrl+P)",
		"m-":    "Subtract from memory (Ctrl+Q)",
		"clear": "Clear (Esc)",
		"back":  "Backspace",
		"pct":   "Percent: 200 + 15% = 230",
	},

	PanelHistory:     "History",
	PanelVariables:   "Variables",
	HistoryEmpty:     "The calculations will be here",
	ClearLink:        "Clear",
	InsertResult:     "Insert Result",
	InsertExpression: "Edit Expression",
	CopyResult:       "Copy Result",
	Delete:           "Delete",
	ClearHistoryAsk:  "Clear the history of the calculations?",

	StatusMemory: func(value string) string { return "M = " + value },

	Settings:         "Settings",
	Help:             "Help",
	About:            "About",
	Language:         "Language:",
	LanguageSystem:   "System",
	Theme:            "Theme:",
	ThemeDark:        "Dark",
	ThemeLight:       "Light",
	DecimalSeparator: "Decimal point:",
	DecimalPoint:     "Point: 1 234.5",
	DecimalComma:     "Comma: 1 234,5",
	Grouping:         "Separate the thousands",
	AboutTitle:       func(app string) string { return "About " + app },
	Version:          "Version",
	Author:           "Author",
	License:          "License",
	VisitWebsite:     "Visit Website",
	Close:            "Close",
	Install:          "Install",
	Update:           "Update",
	Uninstall:        "Uninstall",
	InstallAsk: func(dir string) string {
		return "Install AltCalc into " + dir + "?\n\nIt is added to the Start menu, the desktop and the list of installed apps, and opens in place of this copy."
	},
	InstallFailed: func(err string) string { return "Installation failed: " + err },
	Installed:     "AltCalc is installed",
	UninstallAsk: func(dir string) string {
		return "Uninstall AltCalc?\n\nThe settings and the history stay in " + dir + "."
	},
	Uninstalled:      "AltCalc is uninstalled",
	UninstallRunning: "AltCalc is running. Close it and try again.",

	ReferenceTitle: "Quick Reference",
	Reference: `Type the expression, the result is shown as you type. Enter or = puts it
in the history; an operator typed next goes on with the result.

OPERATORS
  + − × ÷        2 + 3 × 4 = 14     * and / work too
  ^  ²  ³        2^10 = 1024        ** works too
  √              √16 = 4
  !              5! = 120
  %              200 + 15% = 230,  200 × 15% = 30,  50% = 0.5
  mod            7 mod 3 = 1        (7 % 3 too)
  °              sin(90°) = 1 in any angle unit
  ( )            the missing ones at the end are added
  2π, 3(4+5)     a number before a name or a parenthesis multiplies

NUMBERS
  1.5  1,5       the comma is the decimal point outside of a function call
  1 000 000      spaces and _ separate the thousands
  1.5e3          = 1500
  0xFF 0b101 0o17   hex, binary, octal

FUNCTIONS
  sin cos tan cot sec csc   asin acos atan acot atan2(y, x)
  sinh cosh tanh asinh acosh atanh
  sqrt cbrt root(x, n)   exp ln log lg log2 log(x, base)
  abs sign round(x[, digits]) floor ceil trunc frac
  min max sum avg median (any number of arguments)
  gcd lcm fact ncr(n, k) npr(n, k) hypot
  rand() rand(n) rand(a, b)   deg(x) rad(x)
  Functions of one argument work without parentheses: sin 30, ln 2

BITS (integers)
  & | xor ~ << >>    and, or, not, shl, shr too
  The integer results are shown in hex, octal and binary

CONSTANTS AND VARIABLES
  pi π  e  tau τ  phi φ
  ans            the last result
  x = 5          makes a variable; it is kept between the starts

KEYS
  Enter, =       calculate             Esc    clear
  Up, Down       the previous expressions
  F2             degrees / radians / gradians
  F9             change the sign
  Ctrl+C         copy the result       Ctrl+V  paste
  Ctrl+L R P Q M memory: clear, recall, add, subtract, store
  Ctrl+H         history               Ctrl+2  scientific keys
  Ctrl+T         always on top`,
}

var ru = Strings{
	Error: "Ошибка",

	MenuEdit:   "&Правка",
	MenuView:   "&Вид",
	MenuMemory: "П&амять",
	MenuHelp:   "&Справка",

	CmdUndo:            "Отменить",
	CmdRedo:            "Повторить",
	CmdCut:             "Вырезать",
	CmdCopy:            "Копировать",
	CmdCopyCalculation: "Копировать вычисление",
	CmdPaste:           "Вставить",
	CmdSelectAll:       "Выделить всё",
	CmdClear:           "Очистить",
	CmdClearHistory:    "Очистить журнал",
	CmdMemoryClear:     "Очистить память",
	CmdMemoryRecall:    "Вызвать из памяти",
	CmdMemoryAdd:       "Прибавить к памяти",
	CmdMemorySubtract:  "Вычесть из памяти",
	CmdMemoryStore:     "Сохранить в память",
	CmdStandard:        "Обычный",
	CmdScientific:      "Инженерный",
	CmdHistory:         "Журнал",
	CmdAlwaysOnTop:     "Поверх всех окон",
	CmdNextAngle:       "Следующая единица углов",
	CmdReference:       "Краткая справка",
	Angles:             [3]string{"Градусы", "Радианы", "Грады"},

	DisplayHint: "Введите выражение",
	Copied:      "Скопировано",

	ErrSyntax:      "Синтаксическая ошибка",
	ErrUnknownName: func(name string) string { return "Неизвестное имя: " + name },
	ErrDivByZero:   "Деление на ноль",
	ErrDomain:      "Не определено для этого значения",
	ErrOverflow:    "Слишком большое число",
	ErrArgCount:    func(name string) string { return "Неверное число аргументов " + name },
	ErrNotInteger:  "Только для целых чисел",
	ErrReadOnly:    func(name string) string { return name + " нельзя присвоить" },
	ErrUnmatched:   "Лишняя скобка",

	KeyHints: map[string]string{
		"2nd":   "Обратные функции на клавишах",
		"angle": "Единица углов: градусы, радианы, грады (F2)",
		"deg":   "Градусы в любой единице: sin(90°)",
		"mod":   "Остаток: 7 mod 3 = 1",
		"ee":    "Порядок: 1.5e3 = 1500",
		"ans":   "Последний результат",
		"rand":  "Случайное число от 0 до 1",
		"abs":   "Модуль",
		"inv":   "1/x от выражения",
		"neg":   "Сменить знак (F9)",
		"fact":  "Факториал: 5! = 120",
		"comma": "Разделяет аргументы: max(1, 2)",
		"left":  "Курсор влево",
		"right": "Курсор вправо",
		"mc":    "Очистить память (Ctrl+L)",
		"mr":    "Вызвать из памяти (Ctrl+R)",
		"m+":    "Прибавить к памяти (Ctrl+P)",
		"m-":    "Вычесть из памяти (Ctrl+Q)",
		"clear": "Очистить (Esc)",
		"back":  "Удалить символ",
		"pct":   "Процент: 200 + 15% = 230",
	},

	PanelHistory:     "Журнал",
	PanelVariables:   "Переменные",
	HistoryEmpty:     "Здесь будут вычисления",
	ClearLink:        "Очистить",
	InsertResult:     "Вставить результат",
	InsertExpression: "Изменить выражение",
	CopyResult:       "Копировать результат",
	Delete:           "Удалить",
	ClearHistoryAsk:  "Очистить журнал вычислений?",

	StatusMemory: func(value string) string { return "M = " + value },

	Settings:         "Настройки",
	Help:             "Справка",
	About:            "О программе",
	Language:         "Язык:",
	LanguageSystem:   "Системный",
	Theme:            "Тема:",
	ThemeDark:        "Тёмная",
	ThemeLight:       "Светлая",
	DecimalSeparator: "Десятичный знак:",
	DecimalPoint:     "Точка: 1 234.5",
	DecimalComma:     "Запятая: 1 234,5",
	Grouping:         "Разделять разряды",
	AboutTitle:       func(app string) string { return "О программе " + app },
	Version:          "Версия",
	Author:           "Автор",
	License:          "Лицензия",
	VisitWebsite:     "Открыть сайт",
	Close:            "Закрыть",
	Install:          "Установить",
	Update:           "Обновить",
	Uninstall:        "Удалить",
	InstallAsk: func(dir string) string {
		return "Установить AltCalc в " + dir + "?\n\nПрограмма появится в меню «Пуск», на рабочем столе и в списке установленных приложений и откроется вместо этой копии."
	},
	InstallFailed: func(err string) string { return "Не удалось установить: " + err },
	Installed:     "AltCalc установлен",
	UninstallAsk: func(dir string) string {
		return "Удалить AltCalc?\n\nНастройки и журнал останутся в " + dir + "."
	},
	Uninstalled:      "AltCalc удалён",
	UninstallRunning: "AltCalc запущен. Закройте его и повторите.",

	ReferenceTitle: "Краткая справка",
	Reference: `Введите выражение, результат виден сразу. Enter или = заносит его в журнал;
оператор, набранный следом, продолжает вычисление с результатом.

ОПЕРАТОРЫ
  + − × ÷        2 + 3 × 4 = 14     можно * и /
  ^  ²  ³        2^10 = 1024        можно **
  √              √16 = 4
  !              5! = 120
  %              200 + 15% = 230,  200 × 15% = 30,  50% = 0.5
  mod            7 mod 3 = 1        (и 7 % 3)
  °              sin(90°) = 1 при любой единице углов
  ( )            недостающие в конце добавляются
  2π, 3(4+5)     число перед именем или скобкой умножает

ЧИСЛА
  1.5  1,5       запятая - десятичный знак вне вызова функции
  1 000 000      пробелы и _ разделяют разряды
  1.5e3          = 1500
  0xFF 0b101 0o17   шестнадцатеричные, двоичные, восьмеричные

ФУНКЦИИ
  sin cos tan cot sec csc   asin acos atan acot atan2(y, x)
  sinh cosh tanh asinh acosh atanh
  sqrt cbrt root(x, n)   exp ln log lg log2 log(x, основание)
  abs sign round(x[, знаков]) floor ceil trunc frac
  min max sum avg median (сколько угодно аргументов)
  gcd lcm fact ncr(n, k) npr(n, k) hypot
  rand() rand(n) rand(a, b)   deg(x) rad(x)
  Функции одного аргумента работают без скобок: sin 30, ln 2

БИТЫ (целые)
  & | xor ~ << >>    и and, or, not, shl, shr
  Целые результаты показываются в 16-, 8- и 2-ичной системе

КОНСТАНТЫ И ПЕРЕМЕННЫЕ
  pi π  e  tau τ  phi φ
  ans            последний результат
  x = 5          создаёт переменную; она сохраняется между запусками

КЛАВИШИ
  Enter, =       вычислить             Esc    очистить
  Вверх, Вниз    прежние выражения
  F2             градусы / радианы / грады
  F9             сменить знак
  Ctrl+C         копировать результат  Ctrl+V  вставить
  Ctrl+L R P Q M память: очистить, вызвать, прибавить, вычесть, сохранить
  Ctrl+H         журнал                Ctrl+2  инженерные клавиши
  Ctrl+T         поверх всех окон`,
}

var catalog = i18n.NewCatalog(en, map[string]Strings{
	"ru": ru,
})

// T returns the texts in the language of the application
func T() *Strings {
	return catalog.Get(ui.Language())
}

// SetLanguage switches the application to the language of the settings, "" - the system's
func SetLanguage(lang string) {
	if lang == "" {
		lang = ui.SystemLanguage()
	}
	ui.SetLanguage(lang)
}
