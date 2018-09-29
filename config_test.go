/*
Copyright 2018 Ahmed Zaher

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"
)

type input struct {
	args        []string
	prefix      string
	description string
	info        *ReleaseInfo
	conf        interface{}
}

type output struct {
	result string
	err    error
}

type testConf struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Online bool   `json:"online"`
}

var (
	info = &ReleaseInfo{
		BuildTimestamp: time.Now().String(),
		GitCommit:      "0",
		GoVersion:      runtime.Version(),
		ReleaseVersion: "v0.1.0",
	}
	conf = make(map[string]interface{})
)

func withMockedArgs(i *input, fn func(*input) (string, error)) (string, error) {
	args := os.Args
	defer func(args []string) {
		os.Args = args
	}(args)
	os.Args = i.args

	return fn(i)
}

func TestCli(t *testing.T) {
	cases := [][]interface{}{
		{&input{prefix: ""},
			&output{"", errors.New("environment variable prefix [] must start with a letter then letters or underscores")}},
		{&input{prefix: "TEST",
			conf: make(chan int),
			args: []string{""},
		}, &output{"", errors.New("json: unsupported type: chan int")}},
		{&input{prefix: "TEST",
			conf: conf,
			args: []string{""},
		}, &output{"", errors.New("json: Unmarshal(non-pointer map[string]interface {})")}},
		{&input{prefix: "TEST",
			args: []string{"", "-h"},
		}, &output{"Usage: . [-c <config>] [-v]\n\n", nil}},
		{&input{prefix: "TEST",
			args: []string{"", "--help"},
		}, &output{"Usage: . [-c <config>] [-v]\n\nOptions:\n    -c, --config=<config>   JSON string describing the configuration options, JSON values can be placeholders for environment variables that start with 'TEST_' e.g '${DOMAIN}' is replaced with the value of environment variable 'TEST_DOMAIN'. (default: null); setable via $TEST_CONFIG\n    -v, --version           Prints the version and exits (e.g. false)\n    -h, --help              usage (-h) / detailed help text (--help)\n\n", nil}},
		{&input{prefix: "TEST",
			conf: &conf,
			args: []string{"", "-v"},
		}, &output{"Release: \nCommit: \nBuild Time: \nBuilt with: \n", nil}},
		{&input{prefix: "TEST",
			conf: &conf,
			info: info,
			args: []string{"", "-v"},
		}, &output{fmt.Sprintf("Release: %v\nCommit: %v\nBuild Time: %v\nBuilt with: %v\n",
			info.ReleaseVersion, info.GitCommit, info.BuildTimestamp, info.GoVersion), nil}},
		{&input{prefix: "TEST",
			conf: &conf,
			args: []string{""},
		}, &output{"", nil}},
	}

	for _, c := range cases {
		i, o := c[0].(*input), c[1].(*output)

		res, err := withMockedArgs(i, func(in *input) (string, error) {
			return ProcessCommandLine(in.prefix, in.description, in.info, in.conf)
		})

		if o.result != res || (o.err != err && (o.err == nil || err == nil || o.err.Error() != err.Error())) {
			t.Errorf("expected output: (%v, %v), but found: (%v, %v)", o.result, o.err, res, err)
		}
	}
}

func TestCliConfig(t *testing.T) {
	c := &testConf{}

	os.Setenv("TEST_NAME", "Bob")
	i := &input{
		prefix: "TEST",
		conf:   c,
		args:   []string{"", "-c", "{\"id\":1,\"name\":\"${NAME}\",\"online\":true}"},
	}

	res, err := withMockedArgs(i, func(in *input) (string, error) {
		return ProcessCommandLine(in.prefix, in.description, in.info, in.conf)
	})

	if res != "" || err != nil {
		t.Errorf("expected output: (\"\", nil), but found: (%v, %v)", res, err)
	}

	if c.ID != 1 || c.Name != "Bob" || !c.Online {
		j, _ := json.Marshal(c)
		t.Errorf("expected output: %v, but found: %v", i.args[2], string(j))
	}
}
