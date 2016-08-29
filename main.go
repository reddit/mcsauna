package main

import (
	"container/heap"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"io/ioutil"
	"time"
)

const CAPTURE_SIZE = 9000

// startReportingLoop starts a loop that will periodically output statistics
// on the hottest keys, and optionally, errors that occured in parsing.
func startReportingLoop(config Config, hot_keys *HotKeyPool, errors *HotKeyPool) {
	sleep_duration := time.Duration(config.Interval) * time.Second
	time.Sleep(sleep_duration)
	for {
		st := time.Now()
		rotated_keys := hot_keys.Rotate()
		top_keys := rotated_keys.GetTopKeys()
		rotated_errors := errors.Rotate()
		top_errors := rotated_errors.GetTopKeys()

		// Build output
		output := ""
		/* Show all matching keys if regexps are specified */
		if len(config.Regexps) > 0 {
			for {
				if top_keys.Len() == 0 {
					break
				}
				key := heap.Pop(top_keys)
				output += fmt.Sprintf("mcsauna.keys.%s %d\n", key.(*Key).Name, key.(*Key).Hits)
			}
		} else
		/* Show top N requested keys */
		{
			for i := 0; i < config.NumItemsToReport; i++ {
				if top_keys.Len() == 0 {
					break
				}
				key := heap.Pop(top_keys)
				output += fmt.Sprintf("mcsauna.keys.%s %d\n", key.(*Key).Name, key.(*Key).Hits)
			}
		}
		/* Show errors */
		if config.ShowErrors {
			for top_errors.Len() > 0 {
				err := heap.Pop(top_errors)
				output += fmt.Sprintf(
					"mcsauna.errors.%s %d\n", err.(*Key).Name, err.(*Key).Hits)
			}
		}

		// Write to stdout
		if !config.Quiet {
			fmt.Print(output)
		}

		// Write to file
		if config.OutputFile != "" {
			err := ioutil.WriteFile(config.OutputFile, []byte(output), 0666)
			if err != nil {
				panic(err)
			}
		}

		elapsed := time.Now().Sub(st)
		time.Sleep(sleep_duration - elapsed)
	}
}

func main() {
	config_file := flag.String("c", "", "config file")
	interval := flag.Int("n", 0, "reporting interval (seconds, default 5)")
	network_interface := flag.String("i", "", "capture interface (default any)")
	port := flag.Int("p", 0, "capture port (default 11211)")
	num_items_to_report := flag.Int("r", 0, "number of items to report (default 20)")
	quiet := flag.Bool("q", false, "suppress stdout output (default false)")
	output_file := flag.String("w", "", "file to write output to")
	show_errors := flag.Bool("e", true, "show errors in parsing as a metric")
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
	if *quiet != false {
		config.Quiet = *quiet
	}
	if *output_file != "" {
		config.OutputFile = *output_file
	}
	if *show_errors != true {
		config.ShowErrors = *show_errors
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
	errors := NewHotKeyPool()

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

	go startReportingLoop(config, hot_keys, errors)

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
				match_errors := []string{}
				for _, key := range keys {
					matched_regex, err := regexp_keys.Match(key)
					if err != nil {
						match_errors = append(match_errors, "match_error")

						// The user has requested that we also show keys that
						// weren't matched at all, probably for debugging.
						if config.ShowUnmatched {
							matches = append(matches, key)
						}

					} else {
						matches = append(matches, matched_regex)
					}
				}
				hot_keys.Add(matches)
				errors.Add(match_errors)
			}
		} else if cmd_err == ERR_TRUNCATED_CMD {
			errors.Add([]string{"truncated"})
		}

	}
}
