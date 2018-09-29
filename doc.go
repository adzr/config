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

/*
Package config provides a custom CLI function to interpret JSON based configuration.

Brief

This library provides a custom CLI function to interpret JSON based configuration.

Usage

	$ go get -u bitbucket.org/azaher/config

Then, import the package:

  import (
    "bitbucket.org/azaher/config"
  )

Example

  type testConf struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Online bool   `json:"online"`
  }

  func main() {
    conf := &testConf{}

    info := &config.ReleaseInfo {
      GitCommit: GitCommit,
      BuildTimestamp: BuildTimestamp,
      ReleaseVersion: ReleaseVersion,
      GoVersion: GoVersion,
    }

    if out := ProcessCommandLine("TEST_APP", "Test App", info, conf); out != "" {
      println(out)
      os.Exit(0)
    }

    // do whatever you want with the configuration object.
    runApp(conf)
  }

*/
package config
