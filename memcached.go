package main

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	ERR_NONE = iota
	ERR_NO_CMD
	ERR_INVALID_CMD
	ERR_TRUNCATED
	ERR_INCOMPLETE_CMD
	ERR_BAD_BYTES
)

/* A map to translate errors that may arise to the name of the stat that
 * should be reported back when the error occurs. An entry should be added
 * for all non-none errors that can be returned. */
var ERR_TO_STAT = map[int]string{
	ERR_NO_CMD:         "no_cmd",
	ERR_INVALID_CMD:    "invalid_cmd",
	ERR_TRUNCATED:      "truncated",
	ERR_INCOMPLETE_CMD: "incomplete_cmd",
	ERR_BAD_BYTES:      "bad_bytes",
}

// processSingleKeyNoData processes a "get", "incr", or "decr" command, all
// of which only allow for a single key to be passed and have no value field.
//
// On the wire, "get" looks like:
//
//     get key\r\n
//
// And "incr" and "decr" look like:
//
//     cmd key value [noreply]\r\n
//
// Where "noreply" is an optional field that indicates whether the server
// should return a response.
func processSingleKeyNoData(first_line string, remainder []byte) (keys []string, processed_remainder []byte, cmd_err int) {

	// Get the key
	// ... the command should at least consist of "cmd foo", where "foo" is the key
	split_data := strings.Split(first_line, " ")
	if len(split_data) <= 1 {
		return []string{}, remainder, ERR_INCOMPLETE_CMD
	}
	key := split_data[1]
	if key == "" {
		return []string{}, remainder, ERR_INCOMPLETE_CMD
	}

	// Return parsed data
	return []string{key}, remainder, ERR_NONE
}

// processSingleKeyWithData processes a "set", "add", "replace", "append", or
// "prepend" command, all of which only allow for a single key and have a
// corresponding value field.
//
// On the wire, these commands look like:
//
//     cmd key flags exptime bytes [noreply]\r\n
//     <data block of `bytes` length>\r\n
//
// Where "noreply" is an optional field that indicates whether the server
// should return a response.
func processSingleKeyWithData(first_line string, remainder []byte) (keys []string, processed_remainder []byte, cmd_err int) {

	// Get the key
	split_data := strings.Split(first_line, " ")
	if len(split_data) != 5 && len(split_data) != 6 {
		return []string{}, remainder, ERR_INCOMPLETE_CMD
	}
	key, bytes_str := split_data[1], split_data[4]

	// Parse length of stored value
	base := 10
	// ... the max memcached object size is 1MB, so a 32 bit int will suffice
	bitSize := 32
	bytes, err := strconv.ParseInt(bytes_str, base, bitSize)
	if err != nil {
		return []string{}, []byte{}, ERR_INVALID_CMD
	}

	// Make sure we got a full command
	// ... bytes + 2 to account for trailing "\r\n"
	next_command_idx := bytes + 2
	if int64(len(remainder)) < next_command_idx {
		return []string{}, []byte{}, ERR_TRUNCATED
	}

	// Return parsed data
	return []string{key}, remainder[next_command_idx:], ERR_NONE

}

// processMultiKeyNoData processes a "gets" command, which allows for
// multiple keys and has no value field.
//
// On the wire, "gets" looks like:
//
//     gets key1 key2 key3\r\n
func processMultiKeyNoData(first_line string, remainder []byte) (keys []string, processed_remainder []byte, cmd_err int) {

	// Get the key(s)
	// ... the command should at least consist of "cmd foo", where "foo" is the key
	split_data := strings.Split(first_line, " ")
	if len(split_data) <= 1 {
		return []string{}, remainder, ERR_INCOMPLETE_CMD
	}
	keys = split_data[1:]

	// Return parsed data
	return keys, remainder, ERR_NONE
}

var CMD_PROCESSORS = map[string]func(first_line string, remainder []byte) (keys []string, processed_remainder []byte, cmd_err int){
	"get":     processSingleKeyNoData,
	"gets":    processMultiKeyNoData,
	"set":     processSingleKeyWithData,
	"add":     processSingleKeyWithData,
	"replace": processSingleKeyWithData,
	"append":  processSingleKeyWithData,
	"prepend": processSingleKeyWithData,
	"incr":    processSingleKeyNoData,
	"decr":    processSingleKeyNoData,
}

// parseCommand parses a command and list of keys the command is operating on from
// a sequence of application-level data bytes.
func parseCommand(app_data []byte) (cmd string, keys []string, remainder []byte, cmd_err int) {

	// Parse out the command
	space_i := bytes.IndexByte(app_data, byte(' '))
	if space_i == -1 {
		return "", []string{}, []byte{}, ERR_NO_CMD
	}

	// Find the first newline
	newline_i := bytes.Index(app_data, []byte("\r\n"))
	if newline_i == -1 {
		return "", []string{}, []byte{}, ERR_TRUNCATED
	}

	// Validate command
	first_line := string(app_data[:newline_i])
	split_data := strings.Split(first_line, " ")
	cmd = split_data[0]
	if fn, ok := CMD_PROCESSORS[cmd]; ok {
		keys, remainder, cmd_err = fn(first_line, app_data[newline_i+2:])
	} else {
		return "", []string{}, []byte{}, ERR_INVALID_CMD
	}

	return cmd, keys, remainder, cmd_err
}
