package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const Version = "v0.1.0"

type config struct {
	Rule string `json:"rule"`
	PathPattern string `json:"path"`
	Age int `json:"age"`
	Format string `json:"format"`
	ExpireSec int `json:"expire_sec"`
}

func readConfigList(confDir string) []config {
	files, err := ioutil.ReadDir(confDir)
	if err != nil {
		log.Fatal(err)
	}

	configs := make([]config, 0)

	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".json")) {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(confDir, file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		var c config
		if err := json.Unmarshal(content, &c); err != nil {
			log.Fatal(err)
		}
		configs = append(configs, c)
	}

	return configs
}

func globMatchPart(s string, p string) string {
	parts := strings.SplitN(p, "*", 2)
	prefixLength := len(parts[0])
	suffixLength := len(parts[1])
	if !strings.HasPrefix(s, parts[0]) || !strings.HasSuffix(s, parts[1]) {
		log.Fatalf("Path(%v) doesn't match to path pattern:%v", s, p)
	}
	return s[prefixLength:(len(s) - suffixLength)]
}

func convertTimestampFormat(format string) (string, bool) {
	timeFormatMapping := [][]string{
		{"%d", "02"},
		{"%b", "Jan"},
		{"%m", "01"},
		{"%y", "06"},
		{"%Y", "2006"},
		{"%H", "15"},
		{"%M", "04"},
		{"%S", "05"},
		{"%z", "-0700"},
		{"%Z", "MST"},
		{"%%", "%"},
	}
	/*
Supported specifiers:
%d  Day of the month as a zero-padded decimal number.
%b  Month as locale’s abbreviated name.
%m  Month as a zero-padded decimal number.
%y  Year without century as a zero-padded decimal number.
%Y  Year with century as a decimal number.
%H  Hour (24-hour clock) as a zero-padded decimal number.
%M  Minute as a zero-padded decimal number.
%S  Second as a zero-padded decimal number.
%z  UTC offset in the form +HHMM or -HHMM.
%Z  Time zone name. UTC, EST, CST
%%  A literal '%' character.

Golang time format template: Mon Jan 2 15:04:05 -0700 MST 2006
    */
	containsTimezone := false
	if strings.Index(format, "%z") > 0 || strings.Index(format, "%Z") > 0 {
		containsTimezone = true
	}
	result := format
	for _, pair := range timeFormatMapping {
		result = strings.Replace(result, pair[0], pair[1], -1)
	}
	return result, containsTimezone
}

func findTargetsWithAge(pathPattern string, paths []string, age int) []string {
	targets := make([]string, 0)
	for _, p := range paths {
		part := globMatchPart(p, pathPattern)
		if i, err := strconv.Atoi(part); err == nil {
			if i >= age {
				targets = append(targets, p)
			}
		}
	}
	return targets
}

func findTargetsWithTimestamp(now time.Time, pathPattern string, paths []string, strptimeFormat string, expireSec int) []string {
	format, containsTimezone := convertTimestampFormat(strptimeFormat)

	expired := now.Add(time.Duration(expireSec) * time.Second * -1)
	targets := make([]string, 0)
	var t time.Time
	var err error
	for _, p := range paths {
		part := globMatchPart(p, pathPattern)
		if containsTimezone {
			t, err = time.Parse(format, part)
			if err != nil {
				continue
			}
		} else {
			t, err = time.ParseInLocation(format, part, time.Local)
			if err != nil {
				continue
			}
		}
		if t.Before(expired) || t.Equal(expired) {
			targets = append(targets, p)
		}
	}
	return targets
}

func findTargetsWithMtime(now time.Time, paths []string, expireSec int) []string {
	expired := now.Add(time.Duration(expireSec) * time.Second * -1)
	targets := make([]string, 0)
	var finfo os.FileInfo
	var t time.Time
	var err error
	for _, p := range paths {
		if finfo, err = os.Stat(p); err == nil {
			t = finfo.ModTime()
			if t.Before(expired) || t.Equal(expired) {
				targets = append(targets, p)
			}
		}
	}
	return targets
}

func findTargets(c config, now time.Time) []string {
	if strings.Index(c.PathPattern, "*") == -1 {
		log.Fatal("Path in configuration must have a glob(*), but not:" + c.PathPattern)
	}
	if strings.Index(c.PathPattern, "*") != strings.LastIndex(c.PathPattern, "*") {
		log.Fatal("Path in configuration must have just one glob(*), but more:" + c.PathPattern)
	}

	paths, err := filepath.Glob(c.PathPattern);
	if err != nil {
		log.Fatal(err)
	}

	var targets []string

	switch c.Rule {
	case "age":
		targets = findTargetsWithAge(c.PathPattern, paths, c.Age)
	case "timestamp":
		targets = findTargetsWithTimestamp(now, c.PathPattern, paths, c.Format, c.ExpireSec)
	case "mtime":
		targets = findTargetsWithMtime(now, paths, c.ExpireSec)
	default:
		log.Fatal("Unknown rule:" + c.Rule)
	}
	return targets
}

func run(confDir string, now time.Time) {
	configs := readConfigList(confDir)
	for _, c := range configs {
		targets := findTargets(c, now)
		for _, target := range targets {
			err := os.Remove(target)
			if err != nil {
				log.Print("Failed to remove a file:" + target)
			}
		}
	}
}

func main() {
	confDir := os.Args[1]
	if confDir == "--version" {
		fmt.Printf("unlinker %s\n", Version)
		return
	}
	run(confDir, time.Now())
}
