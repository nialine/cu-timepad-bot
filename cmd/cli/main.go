package main

import (
	"bytes"
	"context"
	"cu-timepad-bot/internal/adapters/timepad"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("expected 'get' or 'reserve' subcommands")
	}

	ctx := context.Background()

	switch args[0] {
	case "get":
		getCmd := flag.NewFlagSet("get", flag.ContinueOnError)
		rawURL := getCmd.String("url", "", "URL of Timepad Event")

		if err := getCmd.Parse(args[1:]); err != nil {
			return err
		}

		if *rawURL == "" {
			return fmt.Errorf("no url provided")
		}
		URL, err := url.Parse(*rawURL)
		if err != nil {
			return fmt.Errorf("invalid url")
		}

		client := timepad.New(nil, &http.Client{
			Timeout: 30 * time.Second,
		})

		raw_json, err := client.GetRawData(ctx, URL.String())

		buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024*8))
		json.Indent(buffer, []byte(raw_json), "", "  ")
		fmt.Print("\n", buffer.String())

		return nil

	case "reserve":
		reserveCmd := flag.NewFlagSet("reserve", flag.ContinueOnError)
		raw_baseURL := reserveCmd.String("baseurl", "https://cu-events.timepad.ru/", "Base URL of Timepad")
		magic_number := reserveCmd.Int64("magic", 0, "Magic number to fix everything")
		eventid := reserveCmd.Int64("eventid", 0, "Event id")
		slotid := reserveCmd.Int64("slotid", 0, "Slot id")
		email := reserveCmd.String("email", "", "Email")
		name := reserveCmd.String("name", "", "Name")
		surname := reserveCmd.String("surname", "", "Surname")

		if err := reserveCmd.Parse(args[1:]); err != nil {
			return err
		}

		if *raw_baseURL == "" {
			return fmt.Errorf("no baseurl provided")
		}
		if *eventid == 0 {
			return fmt.Errorf("no eventid provided")
		}
		if *slotid == 0 {
			return fmt.Errorf("no slotid provided")
		}

		baseURL, err := url.Parse(*raw_baseURL)
		if err != nil {
			return fmt.Errorf("invalid url")
		}

		client := timepad.New(baseURL, &http.Client{
			Timeout: 30 * time.Second,
		})

		resp, err := client.ReserveSlot(ctx, *magic_number, *eventid, *slotid, *email, *surname, *name)
		if err != nil {
			return err
		}

		buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024))
		buffer.ReadFrom(resp.Request.Body)
		fmt.Print("Request:\n", buffer.String())

		buffer_raw := bytes.NewBuffer(make([]byte, 0, 1024*1024*8))
		buffer_raw.ReadFrom(resp.Body)
		json.Indent(buffer, buffer_raw.Bytes(), "", "  ")
		fmt.Print("Response:\n", buffer.String())

		return nil
	}
	return fmt.Errorf("unknown subcommand: %s", args[0])
}
