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
	"fmt"
	"os"
	"regexp"
	"strings"

	getopt "github.com/kesselborn/go-getopt"
)

// ReleaseInfo a structure containing details about how this binary has been built.
type ReleaseInfo struct {
	// GitCommit is the git commit hash string used to build this binary.
	GitCommit string

	// BuildTimestamp is the timestamp in a string format when this binary has been built.
	BuildTimestamp string

	// ReleaseVersion is a string defined by the builder of this binary - usually is
	// equivalent to the revision tag released - that represents the release version
	// of the build.
	ReleaseVersion string

	// GoVersion indicates which version of Go has been used to build this binary.
	GoVersion string
}

var (
	// envVarPrefixRegex expression must only allow a prefix with the following rules:
	// 	- All letters must be in uppercase.
	// 	- Must start with a letter.
	// 	- Must contain only letters, numbers or underscores.
	// 	- Must end with a letter or a number.
	envVarPrefixRegex = regexp.MustCompile("\\A[A-Z][A-Z0-9_]*?[A-Z0-9]\\z")

	// placeHolderRegex expression must only allow a placeholder with the following rules:
	// 	- All letters must be in uppercase.
	// 	- Must start with "${" followed by a letter.
	// 	- Must contain only letters, numbers or underscores.
	// 	- Must end with a letter or a number followed by a "}".
	placeHolderRegex = regexp.MustCompile("(?P<PLACEHOLDER>\\$\\{[A-Z][A-Z0-9_]*?[A-Z0-9]\\})")
)

// sanitizePlaceholderToken trims the placeholder token of the "${" prefix and the "}" suffix.
func sanitizePlaceholderToken(envvar string) string {
	return strings.TrimSuffix(strings.TrimPrefix(envvar, "${"), "}")
}

// ProcessCommandLine reads command line arguments and processes them
// leading to one of the following results:
//
//		1. Returns usage or help if either -h or --help flag is specified.
//		2. Returns release information if either -v or --version flag is specified.
//		3. Parses a JSON string specified by -c or --config flags or define in an environment
//		   variable $<envVarPrefix>_CONFIG where <envVarPrefix> is a string passed as parameter
//		   envVarPrefix filling the conf object parameter with the parsed configurations
//		   and then returns an empty string.
//
// The description parameter is shown when displaying help with option --help.
// The info parameter is must not be nil and it has to contain the release information
// which will be displayed with -v/--version option.
// And finally the conf parameter must not be nil, it will carry the application configuration
// parsed from the JSON string passed as an argument along with -c/--config option, or defined
// as environment variable specified $<envVarPrefix>_CONFIG.
func ProcessCommandLine(envVarPrefix, description string, info *ReleaseInfo, conf interface{}) (string, error) {

	var err error

	// make sure that the environment variable prefix format is valid.
	if matches := envVarPrefixRegex.MatchString(envVarPrefix); !matches {
		return "", fmt.Errorf("environment variable prefix [%v] must start with a letter then letters or underscores", envVarPrefix)
	}

	var confRef []byte

	// create an indented JSON string example out of the default configuration
	// to be used as an example in the help/usage output.
	if confRef, err = json.MarshalIndent(conf, "  ", "  "); err != nil {
		return "", err
	}

	// now create the parser with the desired rules for options.
	parser := getopt.Options{
		Description: getopt.Description(description),
		Definitions: []getopt.Option{
			{
				OptionDefinition: fmt.Sprintf("config|c|%v_CONFIG", strings.Trim(strings.ToUpper(envVarPrefix), "_")),
				Description:      fmt.Sprintf("JSON string describing the configuration options, JSON values can be placeholders for environment variables that start with '%v_' e.g '${DOMAIN}' is replaced with the value of environment variable '%v_DOMAIN'.", envVarPrefix, envVarPrefix),
				Flags:            getopt.Optional | getopt.ExampleIsDefault,
				DefaultValue:     string(confRef),
			}, {
				OptionDefinition: "version|v",
				Description:      "Prints the version and exits",
				Flags:            getopt.Flag | getopt.Optional,
				DefaultValue:     false,
			},
		},
	}

	// start parsing command line arguments, given the parser rules and command line input.
	var options, _, _, _ = parser.ParseCommandLine()

	// check on parsed options, if any of the conditions below evaluates to true, then a non-empty string
	// will be returned and the caller of this fuction and the caller should probably output this string
	// to the stdout then exits.
	if help, wantsHelp := options["help"]; wantsHelp && help.String == "usage" {
		// the user has requested to display usage and exit.
		return parser.Usage(), nil
	} else if wantsHelp && help.String == "help" {
		// the user has requested to display help and exit.
		return parser.Help(), nil
	} else if options["version"].Bool {
		// the user has requested to the application version and exit.
		if info == nil {
			info = &ReleaseInfo{}
		}

		return fmt.Sprintf("Release: %v%vCommit: %v%vBuild Time: %v%vBuilt with: %v\n",
			info.ReleaseVersion, fmt.Sprintln(),
			info.GitCommit, fmt.Sprintln(),
			info.BuildTimestamp, fmt.Sprintln(),
			info.GoVersion), nil
	}

	// if this point is reached, it means that user has requested none of the above.
	// so the application is meant to be run and the configuration JSON string must be parsed.
	// the JSON string may contain placeholders e.g. ${PASSWORD} which translates
	// into "I want to inject the value of the environment variable APP_PREFIX_PASSWORD here"
	// so here all the placeholders are being replaced by their real values.
	configJSON := placeHolderRegex.ReplaceAllStringFunc(options["config"].String, func(group string) string {
		// here the placeholder is prefixed with the environment variable prefix to obtain the key,
		// and then the value is being read from os.Getenv by the key.
		return os.Getenv(strings.Join([]string{envVarPrefix, sanitizePlaceholderToken(group)}, "_"))
	})

	// now the JSON string is ready, it needs to be parsed into the supplied configuration structure.
	if err = json.Unmarshal([]byte(strings.TrimSpace(configJSON)), conf); err != nil {
		return "", err
	}

	// a returned empty string means that the caller should not exit the application, instead continue
	// to run with the configuration structure filled.
	return "", nil
}
