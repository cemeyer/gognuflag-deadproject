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

package gnuflag_test

import (
	. "gnuflag"
	"os"
	"testing"
)

var (
	test_bool    = Bool("test_bool", "", false, "bool value")
	test_int     = Int("test_int", "", 0, "int value")
	test_int64   = Int64("test_int64", "", 0, "int64 value")
	test_uint    = Uint("test_uint", "", 0, "uint value")
	test_uint64  = Uint64("test_uint64", "", 0, "uint64 value")
	test_string  = String("test_string", "", "0", "string value")
	test_float   = Float("test_float", "", 0, "float value")
	test_float64 = Float("test_float64", "", 0, "float64 value")
)

func boolString(s string) string {
	if s == "0" {
		return "false"
	}
	return "true"
}

func TestEverything(t *testing.T) {
	m := make(map[string]*Flag)
	desired := "0"
	visitor := func(f *Flag) {
		if len(f.Name) > 5 && f.Name[0:5] == "test_" {
			m[f.Name] = f
			ok := false
			switch {
			case f.Value.String() == desired:
				ok = true
			case f.Name == "test_bool" && f.Value.String() == boolString(desired):
				ok = true
			}
			if !ok {
				t.Error("Visit: bad value", f.Value.String(), "for", f.Name)
			}
		}
	}
	VisitAll(visitor)
	if len(m) != 8 {
		t.Error("VisitAll misses some flags")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	m = make(map[string]*Flag)
	Visit(visitor)
	if len(m) != 0 {
		t.Errorf("Visit sees unset flags")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	// Now set all flags
	Set("test_bool", "true")
	Set("test_int", "1")
	Set("test_int64", "1")
	Set("test_uint", "1")
	Set("test_uint64", "1")
	Set("test_string", "1")
	Set("test_float", "1")
	Set("test_float64", "1")
	desired = "1"
	Visit(visitor)
	if len(m) != 8 {
		t.Error("Visit fails after set")
		for k, v := range m {
			t.Log(k, *v)
		}
	}

	// Test GNU-style arguments:
	Reset()
	os.Args = []string{"-abc", "-dtest", "-e", "100", "--f", "--g=1.5", "--h", "15"}
	Bool("along", "a", false, "")
	Bool("blong", "b", false, "")
	Bool("clong", "c", false, "")
	String("dlong", "d", "aa", "")
	Int("elong", "e", 0, "")
	Bool("f", "f", false, "")
	Float("g", "g", 0.5, "")
	Int("h", "h", 0, "")
	Parse()
}
