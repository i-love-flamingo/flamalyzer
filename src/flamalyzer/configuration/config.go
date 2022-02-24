// Package configuration allows external loading of config files and configuration via flags.
// The CoreConfig is reserved for the Core, whereas the AnalyzerConfig should be used when accessed from an Analyzer
package configuration

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"flamingo.me/flamalyzer/src/flamalyzer/log"
	"github.com/spf13/pflag"

	"gopkg.in/yaml.v2"
)

// AnalyzerConfig used by the Analyzers or others
type AnalyzerConfig interface {
	GetProps(name string) interface{}
	IsDebug() bool
}

// CoreConfig used by the Core
type CoreConfig interface {
	AnalyzerConfig
	LoadConfigFromFiles()
}

// Config main struct
type Config struct {
	props            *configProps
	configSuffixFlag *string
	configFolderFlag *string
}

// This struct can be filled by config-files
type configProps struct {
	Debug                  *bool
	AnalyzerConfigurations map[string]interface{} `yaml:",inline"`
}

// GetProps from the loaded config-files by name
func (c *Config) GetProps(name string) interface{} {
	if props, ok := c.props.AnalyzerConfigurations[name]; ok {
		return props
	}
	log.Println("WARNING: Configuration properties `"+name+"` doesn't exist in the Config-Files! Default values of the analyzer may be used!", c.IsDebug())
	return nil
}

// IsDebug determines weather the debug mode `--debugFlamalyzer` is on
func (c *Config) IsDebug() bool {
	return *c.props.Debug
}

// Set Config-Tags which can be used to use custom files with suffixes
func (c *Config) prepareConfigFlags() {
	fset := pflag.NewFlagSet("Flamalyzer", pflag.ContinueOnError)

	// todo maybe move all this to env vars to prevent the double declaration hassle
	configSuffixMsg := "Suffix for Config-Files that should be loaded e.g `.mySuffix`"
	configFolderMsg := "Path to the Config-Folder with Config-Files in it"
	debugMsg := "Enables Flamalyzer debug messages"
	flag.String("configSuffix", "", "use `--configSuffix` instead. "+configSuffixMsg)
	flag.String("configFolder", "", "use `--configFolder` instead. "+configFolderMsg)
	flag.Bool("debugFlamalyzer", true, "use `--debugFlamalyzer` instead. "+debugMsg)
	c.props.Debug = fset.Bool("debugFlamalyzer", false, debugMsg)
	c.configSuffixFlag = fset.String("configSuffix", "", configSuffixMsg)
	c.configFolderFlag = fset.String("configFolder", "", configFolderMsg)

	fset.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}
	_ = fset.Parse(os.Args[1:])
}

// LoadConfigFromFiles loads the data from the config-file and makes them available in the configProps
func (c *Config) LoadConfigFromFiles() {
	// Create default props so if there is no config-file to read
	// the defaultProps configured in the analyzers will be used
	c.props = new(configProps)
	c.prepareConfigFlags()

	if *c.configFolderFlag == "" {
		log.Println("WARNING: No configfolder defined, this means the default settings will be used!", c.IsDebug())
		return
	}

	configFolderPath, err := filepath.Abs(*c.configFolderFlag)
	if err != nil {
		msg := fmt.Sprint("WARNING: Delivered ConfigFolder-Path is not valid: ", err)
		log.Println(msg, c.IsDebug())
		return
	}
	files, err := ioutil.ReadDir(configFolderPath)
	if err != nil {
		msg := fmt.Sprint("WARNING: Bad ConfigFolder-path, folder not found! ", err)
		log.Println(msg, c.IsDebug())
		return
	}

	// Read all config-files into one byte stream
	var inputStream []byte
	var cfgFileFound = false
	for _, file := range files {
		// Name must contain a flag
		if strings.Contains(file.Name(), *c.configSuffixFlag) && strings.HasSuffix(file.Name(), ".yaml") {
			cfgFileFound = true
			msg := "reading Config-File: `" + file.Name() + "`"
			log.Println(msg, c.IsDebug())

			fileContent, err := ioutil.ReadFile(filepath.Join(configFolderPath, file.Name()))
			if err != nil {
				msg := fmt.Sprintf("WARNING: Error reading the %s Config-File! %s", file.Name(), err)
				log.Println(msg, c.IsDebug())
			} else {
				// Add newline after every read File to prevent complications when a file doesn't end with a newline
				inputStream = append(inputStream, "\n"...)
				inputStream = append(inputStream, fileContent...)
			}
		}
	}
	if !cfgFileFound {
		log.Println("No Config-File(s) found.", c.IsDebug())
	}
	// Extract the rawData into the configProps
	err = yaml.Unmarshal(inputStream, c.props)
	if err != nil {
		msg := fmt.Sprint("WARNING: error trying to unpack the read Configurations to the properties data-struture. DEFAULT CONFIGURATION WILL BE USED! ", err)
		log.Println(msg, c.IsDebug())
	}
}
