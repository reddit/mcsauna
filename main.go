package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"io/ioutil"
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

func startReportingLoop(report_interval int, num_items_to_report int, hot_keys *HotKeyPool) {
	sleep_duration := time.Duration(report_interval) * time.Second
	time.Sleep(sleep_duration)
	for {
		st := time.Now()
		rotated_keys := hot_keys.Rotate()
		top_keys := rotated_keys.GetTopKeys()
		for i := 0; i < num_items_to_report; i++ {
			if len(top_keys) <= i {
				break
			}
			key := top_keys[i]
			hits := rotated_keys.GetHits(key)
			fmt.Printf("key: %s, hits: %d\n", key, hits)
		}
		elapsed := time.Now().Sub(st)
		time.Sleep(sleep_duration - elapsed)
	}
}

func main() {
	config_file := flag.String("c", "", "config file")
	interval := flag.Int("n", 0, "reporting interval (seconds)")
	network_interface := flag.String("i", "", "capture interface")
	port := flag.Int("p", 0, "capture port")
	num_items_to_report := flag.Int("r", 0, "number of items to report")
	flag.Parse()

	// Parse Config
	var config Config
	var err error
	if *config_file != "" {
		config_data, _ := ioutil.ReadFile(*config_file)
		config, err = NewConfig(config_data)
		if err != nil {
			panic(err)
		}
	} else {
		config, err = NewConfig([]byte{})
	}

	// Parse CLI Args
	if *interval != 0 {
		config.Interval = *interval
	}
	if *network_interface != "" {
		config.Interface = *network_interface
	}
	if *port != 0 {
		config.Port = *port
	}
	if *num_items_to_report != 0 {
		config.NumItemsToReport = *num_items_to_report
	}

	// Build Regexps
	regexp_keys := NewRegexpKeys()
	for _, re := range config.Regexps {
		regexp_key, err := NewRegexpKey(re.Re, re.Name)
		if err != nil {
			panic(err)
		}
		regexp_keys.Add(regexp_key)
	}

	hot_keys := NewHotKeyPool()

	// Setup pcap
	handle, err := pcap.OpenLive(config.Interface, CAPTURE_SIZE, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	filter := fmt.Sprintf("tcp and dst port %d", config.Port)
	err = handle.SetBPFFilter(filter)
	if err != nil {
		panic(err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	go startReportingLoop(config.Interval, config.NumItemsToReport, hot_keys)

	// Grab a packet
	for packet := range packetSource.Packets() {
		app_data := packet.ApplicationLayer()
		if app_data == nil {
			continue
		}

		// Process data
		_, keys, cmd_err := parseCommand(app_data.Payload())
		if cmd_err == ERR_NONE {

			// Raw key
			if len(config.Regexps) == 0 {
				hot_keys.Add(keys)
			} else {

				// Regex
				matches := []string{}
				for _, key := range keys {
					matched_regex := regexp_keys.Match(key)
					matches = append(matches, matched_regex)
				}
				hot_keys.Add(matches)
			}
		}

	}
}
