package main

import (
	"bytes"
	"strings"
)

const (
	ERR_NONE = iota
	ERR_NO_CMD
	ERR_INVALID_CMD
	ERR_TRUNCATED
	ERR_INCOMPLETE_CMD
)

/* A map to translate errors that may arise to the name of the stat that
 * should be reported back when the error occurs. An entry should be added
 * for all non-none errors that can be returned. */
var ERR_TO_STAT = map[int]string{
	ERR_NO_CMD:         "no_cmd",
	ERR_INVALID_CMD:    "invalid_cmd",
	ERR_TRUNCATED:      "truncated",
	ERR_INCOMPLETE_CMD: "incomplete_cmd",
}

var VALID_READ_CMDS = []string{"get", "gets"}
var VALID_WRITE_CMDS = []string{"set", "add", "replace", "append", "prepend", "incr", "decr"}

func itemInArray(item string, array []string) bool {
	for _, i := range array {
		if i == item {
			return true
		}
	}
	return false
}

// parseCommand parses a command and list of keys the command is operating on from
// a sequence of application-level data bytes.
func parseCommand(app_data []byte) (cmd string, keys []string, cmd_err int) {

	// Parse out the command
	space_i := bytes.IndexByte(app_data, byte(' '))
	if space_i == -1 {
		return "", []string{}, ERR_NO_CMD
	}

	// Find the first newline
	newline_i := bytes.Index(app_data, []byte("\r\n"))
	if newline_i == -1 {
		return "", []string{}, ERR_TRUNCATED
	}

	// Validate command
	split_data := strings.Split(string(app_data[:newline_i]), " ")
	is_read_cmd := itemInArray(split_data[0], VALID_READ_CMDS)
	is_write_cmd := itemInArray(split_data[0], VALID_WRITE_CMDS)
	if !is_read_cmd && !is_write_cmd {
		return "", []string{}, ERR_INVALID_CMD
	}

	// Extract command & keys
	if is_read_cmd {
		// get commands can be multiple keys
		cmd, keys = split_data[0], split_data[1:]
	} else {
		// set commands can only be one key
		cmd, keys = split_data[0], split_data[1:2]
	}

	// Validate keys
	if len(keys) == 0 || (len(keys) == 1 && keys[0] == "") {
		/* The command was valid, but we didn't find any keys. */
		return "", []string{}, ERR_INCOMPLETE_CMD
	}

	return cmd, keys, ERR_NONE
}
