package main

import (
	"testing"
)

type ParseCommandTest struct {
	RawData []byte
	Cmd     string
	Keys    []string
	CmdErr  int
}

var PARSE_COMMAND_TEST_TABLE = []ParseCommandTest{
	ParseCommandTest{[]byte("set foo 0 0 3\r\nabc\r\n"), "set", []string{"foo"}, ERR_NONE},
	ParseCommandTest{[]byte("get bar\r\n"), "get", []string{"bar"}, ERR_NONE},
	ParseCommandTest{[]byte("\r\n"), "", []string{}, ERR_NO_CMD},
	ParseCommandTest{[]byte("get foo"), "", []string{}, ERR_TRUNCATED_CMD},
	ParseCommandTest{[]byte("foo bar\r\n"), "", []string{}, ERR_INVALID_CMD},
	ParseCommandTest{[]byte("get \r\n"), "", []string{}, ERR_INCOMPLETE_CMD},
	ParseCommandTest{[]byte("get\r\n"), "", []string{}, ERR_NO_CMD},
	ParseCommandTest{[]byte("incr foo 1\r\n"), "incr", []string{"foo"}, ERR_NONE},
	ParseCommandTest{[]byte("decr foo 1\r\n"), "decr", []string{"foo"}, ERR_NONE},
}

func TestParseCommand(t *testing.T) {
	for _, test := range PARSE_COMMAND_TEST_TABLE {
		t.Logf("-> parseCommand(%q)\n", test.RawData)
		cmd, keys, cmd_err := parseCommand(test.RawData)
		t.Logf("<- %v %v %v\n", cmd, keys, cmd_err)

		// Verify Command
		if test.Cmd != cmd {
			t.Errorf("Expected cmd %s, got %s\n", test.Cmd, cmd)
		}

		// Verify Keys
		if len(test.Keys) != len(keys) {
			t.Errorf("Expected keys %v, got %v\n", test.Keys, keys)
		} else {
			for i := range test.Keys {
				if test.Keys[i] != keys[i] {
					t.Errorf("Expected keys %v, got %v\n", test.Keys, keys)
					break
				}
			}
		}

		// Verify Error
		if test.CmdErr != cmd_err {
			t.Errorf("Expected cmd err %d, got %d\n", test.CmdErr, cmd_err)
		}
	}
}
