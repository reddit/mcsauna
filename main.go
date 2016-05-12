package main

import (
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"strings"
	"time"
)

const (
	ERR_NONE = iota
	ERR_NO_CMD
	ERR_INVALID_CMD
	ERR_TRUNCATED_CMD
	ERR_INCOMPLETE_CMD
)

// How many keys we must capture before the hot key pool is rotated and
// reported on.
const ROTATE_THRESHOLD = 10000

const NUM_KEYS_TO_REPORT = 10

const REPORT_INTERVAL = 10 // Seconds

// We only need to capture the first N bytes of a packet to get the command
// and key name.  Commands can be as long as 7 characters, keys can be 250
// characters, and we need to take into account the space in between.
const CAPTURE_SIZE = 7 + 1 + 250

var VALID_READ_CMDS = []string{"get", "gets"}
var VALID_WRITE_CMDS = []string{"set", "add", "replace", "append", "prepend"}

func itemInArray(item string, array []string) bool {
	for _, i := range array {
		if i == item {
			return true
		}
	}
	return false
}

// parseCommand parses a command and list of keys the command is operating on from
// a sequence of application data bytes.
func parseCommand(app_data []byte) (cmd string, keys []string, cmd_err int) {

	// Parse out the command
	space_i := bytes.IndexByte(app_data, byte(' '))
	if space_i == -1 {
		return "", []string{}, ERR_NO_CMD
	}

	// Find the first newline
	newline_i := bytes.Index(app_data, []byte("\r\n"))
	if newline_i == -1 {
		return "", []string{}, ERR_TRUNCATED_CMD
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
		return "", []string{}, ERR_INCOMPLETE_CMD
	}

	return cmd, keys, ERR_NONE
}

func startReportingLoop(hot_keys *HotKeyPool) {
	sleep_duration := REPORT_INTERVAL * time.Second
	time.Sleep(sleep_duration)
	for {
		st := time.Now()
		rotated_keys := hot_keys.Rotate()
		top_keys := rotated_keys.GetTopKeys()
		for i := 0; i < NUM_KEYS_TO_REPORT; i++ {
			key := top_keys[i]
			hits := rotated_keys.GetHits(key)
			fmt.Printf("key: %s, hits: %d\n", key, hits)
		}
		elapsed := time.Now().Sub(st)
		time.Sleep(sleep_duration - elapsed)
	}
}

func main() {
	hot_keys := NewHotKeyPool()

	// TODO: Flag for interface, port
	handle, err := pcap.OpenLive("lo", CAPTURE_SIZE, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	err = handle.SetBPFFilter("tcp and dst port 11211")
	if err != nil {
		panic(err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	go startReportingLoop(hot_keys)
	for packet := range packetSource.Packets() {
		app_data := packet.ApplicationLayer()
		if app_data == nil {
			continue
		}

		// Process data
		cmd, keys, cmd_err := parseCommand(app_data.Payload())
		if cmd_err == ERR_NONE {
			hot_keys.Add(keys)
		}
		_ = cmd

	}
}
