/*
	gnuflag - A 'flag'-like package for handling GNU-style program options
	Copyright (C) 2010  Conrad Meyer <cemeyer@cs.washington.edu>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
	The gnuflag package implements command-line flag parsing.

	Usage:

	1) Define flags using flag.String(), Bool(), Int(), etc. Example:
		import "gnuflag"
		var ip *int = gnuflag.Int("f", "flagname", 1234, "help message for flagname")
	If you like, you can bind the flag to a variable using the Var() functions.
		var flagvar int
		func init() {
			gnuflag.IntVar(&flagvar, "f", "flagname", 1234, "help message for flagname")
		}

	2) After all flags are defined, call
		gnuflag.Parse()
	to parse the command line into the defined flags.

	3) Flags may then be used directly. If you're using the flags themselves,
	they are all pointers; if you bind to variables, they're values.
		fmt.Println("ip has value ", *ip);
		fmt.Println("flagvar has value ", flagvar);

	4) After parsing, flag.Arg(i) is the i'th argument, excluding flags.
	Args are indexed from 0 up to flag.NArg(). The flag.Args() function
	provides a slice of all remaining arguments.

	Command line flag syntax:
		-f
		-fargument
		-f argument
		--flag
		--flag=argument
		--flag argument

	Two minus signs must be used for the long-name options; a single
	minus sign indicates a short-name option.

	Boolean flags can be composed (setting a value of true), e.g.:
		-pzF
	is identical to:
		-p -z -F

	Other-valued flags may not be composed. They look like:
		-x foo
	or:
		-xfoo

	The only way to set a boolean flag to false is to use the long form
	and 'flag=value' notation, i.e.:
		--foobar=false

	Flag parsing stops after the terminator "--".

	Integer flags accept 1234, 0664, 0x1234 and may be negative.
	Boolean flags may be 1, 0, t, f, true, false, TRUE, FALSE, True, False.
*/
package gnuflag

import (
	"container/vector"
	"fmt"
	"os"
	"strconv"
	"utf8"
)

// TODO(r): BUG: atob belongs elsewhere
func atob(str string) (value bool, ok bool) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true, true
	case "0", "f", "F", "false", "FALSE", "False":
		return false, true
	}
	return false, false
}

// -- Bool Value
type boolValue struct {
	p *bool
}

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return &boolValue{p}
}

func (b *boolValue) set(s string) bool {
	v, ok := atob(s)
	*b.p = v
	return ok
}

func (b *boolValue) String() string { return fmt.Sprintf("%v", *b.p) }

// -- Int Value
type intValue struct {
	p *int
}

func newIntValue(val int, p *int) *intValue {
	*p = val
	return &intValue{p}
}

func (i *intValue) set(s string) bool {
	v, err := strconv.Atoi(s)
	*i.p = int(v)
	return err == nil
}

func (i *intValue) String() string { return fmt.Sprintf("%v", *i.p) }

// -- Int64 Value
type int64Value struct {
	p *int64
}

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return &int64Value{p}
}

func (i *int64Value) set(s string) bool {
	v, err := strconv.Atoi64(s)
	*i.p = v
	return err == nil
}

func (i *int64Value) String() string { return fmt.Sprintf("%v", *i.p) }

// -- Uint Value
type uintValue struct {
	p *uint
}

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return &uintValue{p}
}

func (i *uintValue) set(s string) bool {
	v, err := strconv.Atoui(s)
	*i.p = uint(v)
	return err == nil
}

func (i *uintValue) String() string { return fmt.Sprintf("%v", *i.p) }

// -- uint64 Value
type uint64Value struct {
	p *uint64
}

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return &uint64Value{p}
}

func (i *uint64Value) set(s string) bool {
	v, err := strconv.Atoui64(s)
	*i.p = uint64(v)
	return err == nil
}

func (i *uint64Value) String() string { return fmt.Sprintf("%v", *i.p) }

// -- string Value
type stringValue struct {
	p *string
}

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return &stringValue{p}
}

func (s *stringValue) set(val string) bool {
	*s.p = val
	return true
}

func (s *stringValue) String() string { return fmt.Sprintf("%s", *s.p) }

// -- Float Value
type floatValue struct {
	p *float
}

func newFloatValue(val float, p *float) *floatValue {
	*p = val
	return &floatValue{p}
}

func (f *floatValue) set(s string) bool {
	v, err := strconv.Atof(s)
	*f.p = v
	return err == nil
}

func (f *floatValue) String() string { return fmt.Sprintf("%v", *f.p) }

// -- Float64 Value
type float64Value struct {
	p *float64
}

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return &float64Value{p}
}

func (f *float64Value) set(s string) bool {
	v, err := strconv.Atof64(s)
	*f.p = v
	return err == nil
}

func (f *float64Value) String() string { return fmt.Sprintf("%v", *f.p) }

// FlagValue is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
type FlagValue interface {
	String() string
	set(string) bool
}

// A Flag represents the state of a flag.
type Flag struct {
	Name      string    // name as it appears on command line
	ShortName string    // shortname (optional)
	Usage     string    // help message
	Value     FlagValue // value as set
	DefValue  string    // default value (as text); for usage message
}

type allFlags struct {
	actual map[string]*Flag
	formal map[string]*Flag
	snames map[int]string
	args   *vector.StringVector
}

var flags *allFlags = &allFlags{make(map[string]*Flag), make(map[string]*Flag), make(map[int]string), new([]string)}

// VisitAll visits the flags, calling fn for each. It visits all flags, even those not set.
func VisitAll(fn func(*Flag)) {
	for _, f := range flags.formal {
		fn(f)
	}
}

// Visit visits the flags, calling fn for each. It visits only those flags that have been set.
func Visit(fn func(*Flag)) {
	for _, f := range flags.actual {
		fn(f)
	}
}

// Lookup returns the Flag structure of the named flag, returning nil if none exists.
func Lookup(name string) *Flag {
	f, ok := flags.formal[name]
	if !ok {
		return nil
	}
	return f
}

// Set sets the value of the named flag.  It returns true if the set succeeded; false if
// there is no such flag defined, or if the value is not acceptable for the flag.
func Set(name, value string) bool {
	f, ok := flags.formal[name]
	if !ok {
		return false
	}
	ok = f.Value.set(value)
	if !ok {
		return false
	}
	flags.actual[name] = f
	return true
}

// Reset prepares gnuflag to parse the arg list again. It is mostly for testing
// purposes.
func Reset() {
	flags = &allFlags{make(map[string]*Flag), make(map[string]*Flag), make(map[int]string), new([]string)}
}

// PrintDefaults prints to standard error the default values of all defined flags.
func PrintDefaults() {
	VisitAll(func(f *Flag) {
		var format string
		if _, ok := f.Value.(*stringValue); ok {
			// put quotes on the value
			format = "--%s=%q: %s\n"
		} else {
			format = "--%s=%s: %s\n"
		}
		if f.ShortName != "" {
			fmt.Fprintf(os.Stderr, "  -%s, "+format, f.ShortName, f.Name, f.DefValue, f.Usage)
		} else {
			fmt.Fprintf(os.Stderr, "      "+format, f.Name, f.DefValue, f.Usage)
		}
	})
}

// UsageTemplate is a string formatting template that can be overridden to provide
// more useful usage messages. The %s argument is the program name.
var UsageTemplate = "Usage: %s [OPTION]... [ARGS]\n"

// Usage prints to standard error a default usage message documenting all defined flags.
// The function is a variable that may be changed to point to a custom function.
var Usage = func() {
	fmt.Fprintf(os.Stderr, UsageTemplate, os.Args[0])
	PrintDefaults()
}

// NFlag is the number of actual flags processed.
func NFlag() int { return len(flags.actual) }

// Arg returns the i'th command-line argument.  Arg(0) is the first remaining argument
// after flags have been processed.
func Arg(i int) string {
	if i < 0 || i >= flags.args.Len() {
		return ""
	}
	return flags.args.At(i)
}

// NArg is the number of arguments remaining after flags have been processed.
func NArg() int { return flags.args.Len() }

// Args returns the non-flag command-line arguments.
func Args() []string { return flags.args.Data() }

func add(name string, shortName string, value FlagValue, usage string) {
	// Remember the default value as a string; it won't change.
	f := &Flag{name, shortName, usage, value, value.String()}
	_, alreadythere := flags.formal[name]
	if alreadythere {
		fmt.Fprintln(os.Stderr, "flag redefined:", name)
		panic("flag redefinition") // Happens only if flags are declared with identical names
	}
	// Verify that shortName is the empty string, or a single UTF-8 character.
	if shortName == "" {
		goto noShortName
	}
	r, n := utf8.DecodeRuneInString(shortName)
	if r == utf8.RuneError || n < len(shortName) {
		fmt.Fprintln(os.Stderr, "flag shortname invalid:", name)
		panic("flag shortname invalid")
	}
	flags.snames[r] = name
noShortName:
	flags.formal[name] = f
}

// BoolVar defines a bool flag with specified name, short name, default value, and
// usage string. The argument p points to a bool variable in which to store the value
// of the flag.
func BoolVar(p *bool, name, shortName string, value bool, usage string) {
	add(name, shortName, newBoolValue(value, p), usage)
}

// Bool defines a bool flag with specified name, short name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func Bool(name, shortName string, value bool, usage string) *bool {
	p := new(bool)
	BoolVar(p, name, shortName, value, usage)
	return p
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name, shortName string, value int, usage string) {
	add(name, shortName, newIntValue(value, p), usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name, shortName string, value int, usage string) *int {
	p := new(int)
	IntVar(p, name, shortName, value, usage)
	return p
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64Var(p *int64, name, shortName string, value int64, usage string) {
	add(name, shortName, newInt64Value(value, p), usage)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64(name, shortName string, value int64, usage string) *int64 {
	p := new(int64)
	Int64Var(p, name, shortName, value, usage)
	return p
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func UintVar(p *uint, name, shortName string, value uint, usage string) {
	add(name, shortName, newUintValue(value, p), usage)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func Uint(name, shortName string, value uint, usage string) *uint {
	p := new(uint)
	UintVar(p, name, shortName, value, usage)
	return p
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64Var(p *uint64, name, shortName string, value uint64, usage string) {
	add(name, shortName, newUint64Value(value, p), usage)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64(name, shortName string, value uint64, usage string) *uint64 {
	p := new(uint64)
	Uint64Var(p, name, shortName, value, usage)
	return p
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name, shortName, value string, usage string) {
	add(name, shortName, newStringValue(value, p), usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name, shortName, value string, usage string) *string {
	p := new(string)
	StringVar(p, name, shortName, value, usage)
	return p
}

// FloatVar defines a float flag with specified name, default value, and usage string.
// The argument p points to a float variable in which to store the value of the flag.
func FloatVar(p *float, name, shortName string, value float, usage string) {
	add(name, shortName, newFloatValue(value, p), usage)
}

// Float defines a float flag with specified name, default value, and usage string.
// The return value is the address of a float variable that stores the value of the flag.
func Float(name, shortName string, value float, usage string) *float {
	p := new(float)
	FloatVar(p, name, shortName, value, usage)
	return p
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name, shortName string, value float64, usage string) {
	add(name, shortName, newFloat64Value(value, p), usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64(name, shortName string, value float64, usage string) *float64 {
	p := new(float64)
	Float64Var(p, name, shortName, value, usage)
	return p
}


func (f *allFlags) parseOne(index int) (ok bool, next int) {
	s := os.Args[index]
	// Take care of non-flag arguments.
	if len(s) == 0 || s[0] != '-' || s == "-" {
		f.args.Push(s)
		return true, index + 1
	}
	if s == "--" {
		v := vector.StringVector(os.Args[index+1:])
		f.args.AppendVector(&v)
		return false, -1
	}
	var errorStr string
	// Sort out flag arguments.
	if s[1] != '-' {
		for {
			// Deal with shortname flags
			sname, sz := utf8.DecodeRuneInString(s[1:])
			if sname == utf8.RuneError {
				errorStr = "invalid UTF-8 character"
				goto argError
			}
			name, ok := f.snames[sname]
			if !ok {
				errorStr = fmt.Sprintf("flag provided but not defined: -%s\n", string(sname))
				goto argError
			}
			rest := s[1+sz:]
			// Check for (bad) extraneous flags
			if _, ok := f.actual[name]; ok {
				errorStr = fmt.Sprintf("flag specified twice: -%s\n", string(sname))
				goto argError
			}
			flag, ok := f.formal[name]
			if !ok {
				errorStr = fmt.Sprintf("flag provided but not defined: -%s\n", string(sname))
				goto argError
			}
			// Try and understand the value of the flag
			if f, ok := flag.Value.(*boolValue); ok { // special case: doesn't need an arg
				f.set("true")
				s = "-" + rest
				continue
			}
			has_value := false
			if rest != "" {
				has_value = true
			}
			if !has_value && index < len(os.Args)-1 {
				has_value = true
				index++
				rest = os.Args[index]
			}
			if !has_value {
				errorStr = fmt.Sprintf("flag needs an argument: -%s\n", string(sname))
				goto argError
			}
			if ok = flag.Value.set(rest); !ok {
				errorStr = fmt.Sprintf("invalid value %s for flag: -%s\n", rest, string(sname))
				goto argError
			}
			break
		}
	} else {
		// Long name flags
		name := s[2:]
		if name[0] == '-' || name[0] == '=' {
			errorStr = fmt.Sprintln("bad flag syntax:", s)
			goto argError
		}
		has_value := false
		value := ""
		for i, rune := range name {
			if rune == '=' {
				value = name[i+1:] // the '=' rune has len 1
				has_value = true
				name = name[0:i]
				break
			}
		}
		// Check for (bad) extraneous flags
		if _, ok := f.actual[name]; ok {
			errorStr = fmt.Sprintf("flag specified twice: -%s\n", name)
			goto argError
		}
		flag, ok := f.formal[name]
		if !ok {
			errorStr = fmt.Sprintf("flag provided but not defined: -%s\n", name)
			goto argError
		}
		// Try and understand the value of the flag
		if f, ok := flag.Value.(*boolValue); ok { // special case: doesn't need an arg
			if has_value {
				if !f.set(value) {
					errorStr = fmt.Sprintf("invalid boolean value %t for flag: -%s\n", value, name)
					goto argError
				}
			} else {
				f.set("true")
			}
		} else {
			// It must have a value, which might be the next argument.
			if !has_value && index < len(os.Args)-1 {
				// value is the next arg
				has_value = true
				index++
				value = os.Args[index]
			}
			if !has_value {
				errorStr = fmt.Sprintf("flag needs an argument: -%s\n", name)
				goto argError
			}
			if ok = flag.Value.set(value); !ok {
				errorStr = fmt.Sprintf("invalid value %s for flag: -%s\n", value, name)
				goto argError
			}
		}
		f.actual[name] = flag
	}
	return true, index + 1
argError:
	fmt.Fprint(os.Stderr, errorStr)
	Usage()
	os.Exit(2)
	return false, -1
}

// Parse parses the command-line flags.  Must be called after all flags are defined
// and before any are accessed by the program.
func Parse() {
	var ok bool
	for i := 1; i < len(os.Args); {
		if ok, i = flags.parseOne(i); !ok {
			break
		}
	}
}
