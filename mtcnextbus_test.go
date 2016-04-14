package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetRouteID(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Express - Target - Hwy 252 and 73rd Av P&R - Mpls", "765"},
		{"METRO Blue Line", "901"},
	}
	for _, c := range cases {
		got, err := getRouteID(c.in)
		if got != c.want {
			t.Errorf("getRouteID(%q) == %q, want %q, error: %v", c.in, got, c.want, err)
		}
	}
}

func TestGetDirectionID(t *testing.T) {
	cases := []struct {
		in, in2, want string
	}{
		{"765", "south", "1"},
		{"901", "north", "4"},
	}
	for _, c := range cases {
		got, err := getDirectionID(c.in, c.in2)
		if got != c.want {
			t.Errorf("getDirectionID(%q, %q) == %q, want %q, error: %v", c.in, c.in2, got, c.want, err)
		}
	}
}

func TestGetStopID(t *testing.T) {
	cases := []struct {
		in, in2, in3, want string
	}{
		{"765", "1", "Target North Campus Building F", "TGBF"},
		{"901", "4", "Target Field Station Platform 1", "TF1I"},
	}
	for _, c := range cases {
		got, err := getStopID(c.in, c.in2, c.in3)
		if got != c.want {
			t.Errorf("getStopID(%q, %q, %q) == %q, want %q, error: %v", c.in, c.in2, c.in3, got, c.want, err)
		}
	}
}

func TestGetDeparture(t *testing.T) {
	cases := []struct {
		in, in2, in3, want, okerr string
	}{
		{"765", "1", "TGBF", "Min", "No upcoming departures found for today."},
		{"901", "4", "TF1I", "Min", "No upcoming departures found for today."},
		{"", "", "", "", "error getting departures: bad response code from API: 404 Not Found"},
	}
	for _, c := range cases {
		got, err := getDeparture(c.in, c.in2, c.in3)
		if strings.Contains(got, c.want) == false && strings.Contains(got, "Due") == false {
			if fmt.Sprintf("%s", err) != c.okerr {
				t.Errorf("getDeparture(%q, %q, %q) == %q, want %q or err %q, error: %v", c.in, c.in2, c.in3, got, c.want, c.okerr, err)
			}
		}
		if err != nil {
			if fmt.Sprintf("%s", err) != c.okerr {
				t.Errorf("getDeparture(%q, %q, %q) == %q, want %q or err %q, error: %v", c.in, c.in2, c.in3, got, c.want, c.okerr, err)
			}
		}
	}
}
