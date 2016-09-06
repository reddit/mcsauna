package main

import (
	"bytes"
	"testing"
)

type ParseCommandTest struct {
	RawData   []byte
	Cmd       string
	Keys      []string
	Remainder []byte
	CmdErr    int
}

var PARSE_COMMAND_TEST_TABLE = []ParseCommandTest{

	// Single Command Per Packet Tests
	ParseCommandTest{[]byte("set foo 0 0 3\r\nabc\r\n"), "set", []string{"foo"}, []byte{}, ERR_NONE},
	ParseCommandTest{[]byte("get bar\r\n"), "get", []string{"bar"}, []byte{}, ERR_NONE},
	ParseCommandTest{[]byte("\r\n"), "", []string{}, []byte{}, ERR_NO_CMD},
	ParseCommandTest{[]byte("get foo"), "", []string{}, []byte{}, ERR_TRUNCATED},
	ParseCommandTest{[]byte("foo bar\r\n"), "", []string{}, []byte{}, ERR_INVALID_CMD},
	ParseCommandTest{[]byte("get \r\n"), "get", []string{}, []byte{}, ERR_INCOMPLETE_CMD},
	ParseCommandTest{[]byte("get\r\n"), "", []string{}, []byte{}, ERR_NO_CMD},
	ParseCommandTest{[]byte("incr foo 1\r\n"), "incr", []string{"foo"}, []byte{}, ERR_NONE},
	ParseCommandTest{[]byte("decr foo 1\r\n"), "decr", []string{"foo"}, []byte{}, ERR_NONE},

	// Multiple Commands Per Packet Tests
	ParseCommandTest{[]byte("get foo\r\nget bar\r\n"), "get", []string{"foo"}, []byte("get bar\r\n"), ERR_NONE},
	ParseCommandTest{[]byte("set foo 0 0 3\r\nabc\r\nget bar\r\n"), "set", []string{"foo"}, []byte("get bar\r\n"), ERR_NONE},
}

func TestParseCommand(t *testing.T) {
	for test_i, test := range PARSE_COMMAND_TEST_TABLE {
		t.Logf(" -> parseCommand(%q)\n", test.RawData)
		cmd, keys, remainder, cmd_err := parseCommand(test.RawData)
		t.Logf(" <- %v %v %v\n", cmd, keys, cmd_err)

		// Verify Command
		if test.Cmd != cmd {
			t.Errorf("Test %d: expected cmd %s, got %s\n", test_i, test.Cmd, cmd)
		}

		// Verify Keys
		if len(test.Keys) != len(keys) {
			t.Errorf("Test %d: expected keys %v, got %v\n",
				test_i, test.Keys, keys)
		} else {
			for i := range test.Keys {
				if test.Keys[i] != keys[i] {
					t.Errorf("Test %d: expected keys %v, got %v\n",
						test_i, test.Keys, keys)
					break
				}
			}
		}

		// Verify Remainder
		if !bytes.Equal(test.Remainder, remainder) {
			t.Errorf("Test %d: expected remainder %v, got %v\n",
				test_i, test.Remainder, remainder)
		}

		// Verify Error
		if test.CmdErr != cmd_err {
			t.Errorf("Test %d: expected cmd err %d, got %d\n",
				test_i, test.CmdErr, cmd_err)
		}
	}
}
