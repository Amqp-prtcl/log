package log

type LogLevel int

func (l LogLevel) String() string {
	switch l {
	case 0:
		return "DEGUB"
	case 1:
		return "INFO"
	case 2:
		return "WARN"
	case 3:
		return "ERROR"
	case 4:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func (l LogLevel) Lower() LogLevel {
	if l > 0 {
		return l - 1
	}
	return l
}

func (l LogLevel) Restrict() LogLevel {
	if l < 3 {
		return l + 1
	}
	return l
}

func (l LogLevel) Permits(level LogLevel) bool {
	return l <= level
}

// OutputType represents a type configuration used by log calls
// to save already compiled buffers
//
// New custom OutpytType can be created by packages to be used by
// custom outputs
type OutputType int

const (
	T_Text OutputType = iota
	T_JSON
)

const (
	L_Debug LogLevel = iota
	L_Info
	L_Warn
	L_Error
	L_Fatal
)

const (
	TimeFieldKey   = "time"
	LevelFieldKey  = "level"
	PrefixFieldKey = "prefix"
)

const (
	// wether to show time (ex: 20/12/2022 12:43); in the case of T_JSON the date
	// is marshaled by calling json.Marshal(time)
	F_Time = 1 << iota

	// adds microseconds to time; ignored by T_JSON, assumes F_Time.
	F_Micro

	// wether to show prefixes (by default all).
	F_Prefix

	// wether to only show last prefix, assumes and overwrites LPrefix.
	F_LastPrefix

	// wether to show log level
	F_Level

	// wether to add '\n' at the end of each line if not already present (or always add it in case of T_JSON)
	F_NewLine

	// wether to add fields as top level fields in marshaled structure; ignored by T_TEXT
	//
	// if multiple F_Fields_* are added the priority is F_Fields_A -> F_Fields -> F_Fields_B
	F_Fields

	// similar to F_Fields but fields are marshaled as JSON object in
	// a single fields named 'fields' instead of top level fields
	//
	// if multiple F_Fields_* are added the priority is F_Fields_A -> F_Fields -> F_Fields_B
	F_Fields_A

	// similar to F_Fields but fields are marshaled as an array of objects in
	// a single field named 'fields' instead of top level fields
	//
	// if multiple F_Fields_* are added the priority is F_Fields_A -> F_Fields -> F_Fields_B
	F_Fields_B

	// if specified, compiled buffer won't be saved in entry.
	// So that even if a later call to GetCompiled() with the same flags and
	// output type is made, it will return false
	F_NotSave

	// flags used by default logger
	F_Std = F_Time | F_Prefix | F_Level | F_NewLine | F_Fields
)

// duplicates duplicates and concatenates multiple slices with a single allocation
//
// NOTE: duplicate only does shallow copies of the content of the slice(s)
func duplicate[T any](slices ...[]T) []T {
	var l int
	for _, v := range slices {
		l += len(v)
	}
	var s = make([]T, 0, l)
	for _, v := range slices {
		s = append(s, v...)
	}
	return s
}

// Must panics if err is non-nil
func Must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
