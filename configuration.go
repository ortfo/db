package ortfodb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ll "github.com/ewen-lbh/label-logger-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/go-homedir"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

const DefaultConfigurationFilename = "ortfodb.yaml"
const DefaultScatteredModeFolder = ".ortfo"

type ExtractColorsConfiguration struct {
	Enabled      bool
	Extract      []string
	DefaultFiles []string `yaml:"default files"`
}

type MakeGIFsConfiguration struct {
	Enabled          bool
	FileNameTemplate string `yaml:"file name template"`
}

type MakeThumbnailsConfiguration struct {
	Enabled          bool
	Sizes            []int
	InputFile        string `yaml:"input file"`
	FileNameTemplate string `yaml:"file name template"`
}

type BuildSteps struct {
	ExtractColors  ExtractColorsConfiguration  `yaml:"extract colors"`
	MakeGifs       MakeGIFsConfiguration       `yaml:"make gifs"`
	MakeThumbnails MakeThumbnailsConfiguration `yaml:"make thumbnails"`
}

type TagsConfiguration struct {
	// Path to file describing all tags.
	Repository string
}

type TechnologiesConfiguration struct {
	// Path to file describing all technologies.
	Repository string
}

type MediaConfiguration struct {
	// Path to the media directory.
	At string
}

// Configuration represents what the ortfodb.yaml configuration file describes.
type Configuration struct {
	// Signals whether the configuration was instanciated by DefaultConfiguration.
	IsDefault bool `yaml:"-"`

	ExtractColors       ExtractColorsConfiguration  `yaml:"extract colors,omitempty"`
	MakeGifs            MakeGIFsConfiguration       `yaml:"make gifs,omitempty"`
	MakeThumbnails      MakeThumbnailsConfiguration `yaml:"make thumbnails,omitempty"`
	Media               MediaConfiguration          `yaml:"media,omitempty"`
	ScatteredModeFolder string                      `yaml:"scattered mode folder"`
	Tags                TagsConfiguration           `yaml:"tags,omitempty"`
	Technologies        TechnologiesConfiguration   `yaml:"technologies,omitempty"`

	// Path to the directory containing all projects. Must be absolute.
	ProjectsDirectory string `yaml:"projects at"`

	// Exporter-specific configuration. Maps exporter names to their configuration.
	Exporters map[string]map[string]interface{} `yaml:"exporters,omitempty"`

	// Where was the configuration loaded from
	source string
}

// LoadConfiguration loads the given configuration YAML file and puts it contents into loadInto.
func LoadConfiguration(filename string, loadInto *Configuration) error {
	raw, err := readFileBytes(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(raw, loadInto)
	if err != nil {
		return err
	}
	ll.Debug("Loaded configuration from %s to %#v", filename, loadInto)
	return nil
}

// NewConfiguration loads a YAML configuration file.
// This function also validates the configuration and prints any error to the user.
// Use LoadConfiguration for a lower-level function that just loads the YAML file into a struct.
func NewConfiguration(filename string) (Configuration, error) {
	if filename == DefaultConfigurationFilename {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			ll.Log("Writing", "yellow", "default configuration file at %s", filename)
			defaultConfig := DefaultConfiguration()
			err := writeYAML(defaultConfig, filename)
			if err != nil {
				return Configuration{}, fmt.Errorf("while writing default configuration: %w", err)
			}

			ll.Warn("default configuration assumes that your projects live in %s. Change this with [bold]projects at[reset] in the generated configuration file", defaultConfig.ProjectsDirectory)
			return DefaultConfiguration(), nil
		}
	}

	validated, validationErrors, err := ValidateConfiguration(filename)
	if err != nil {
		return Configuration{}, fmt.Errorf("while validating configuration %s: %v", filename, err.Error())
	}
	if !validated {
		DisplayValidationErrors(validationErrors, filename)
		return Configuration{}, fmt.Errorf("the configuration file is invalid. See validation errors above")
	}

	config := Configuration{source: filename}
	err = LoadConfiguration(filename, &config)
	if err != nil {
		return Configuration{}, fmt.Errorf("while loading configuration file at %s: %w", filename, err)
	}

	config.ProjectsDirectory, err = homedir.Expand(config.ProjectsDirectory)
	if err != nil {
		return Configuration{}, fmt.Errorf("while expanding home symbol for project at: %w", err)
	}

	config.Tags.Repository, err = homedir.Expand(config.Tags.Repository)
	if err != nil {
		return Configuration{}, fmt.Errorf("while expanding home symbol for tags repository at: %w", err)
	}

	config.Technologies.Repository, err = homedir.Expand(config.Technologies.Repository)
	if err != nil {
		return Configuration{}, fmt.Errorf("while expanding home symbol for technologies repository at: %w", err)
	}

	// Make sure the project directory exists, is a directory and is absolute.
	err = checkProjectsDirectory(config)
	if err != nil {
		return Configuration{}, err
	}

	// Set default value for ScatteredModeFolder
	if config.ScatteredModeFolder == "" {
		config.ScatteredModeFolder = ".ortfo"
	}

	// Remove trailing slash(es) from folder name.
	config.ScatteredModeFolder = strings.TrimRight(config.ScatteredModeFolder, "/\\")

	// Expand ~
	config.MakeThumbnails.FileNameTemplate, err = homedir.Expand(config.MakeThumbnails.FileNameTemplate)
	if err != nil {
		return Configuration{}, fmt.Errorf("could not expand home directory symbol of make thumbnails.file name template: %w", err)
	}

	config.Media.At, err = homedir.Expand(config.Media.At)
	if err != nil {
		return Configuration{}, fmt.Errorf("could not expand home directory symbol of media.at: %w", err)
	}

	return config, nil
}

func checkProjectsDirectory(config Configuration) error {
	stat, err := os.Stat(config.ProjectsDirectory)
	if os.IsNotExist(err) {
		return fmt.Errorf("projects directory %s does not exist", config.ProjectsDirectory)
	} else if err != nil {
		return fmt.Errorf("while checking projects directory at %s: %w", config.ProjectsDirectory, err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("projects directory %s is not a directory", config.ProjectsDirectory)
	}
	return nil
}

// ValidateConfiguration uses the JSON configuration schema ConfigurationJSONSchema to validate the configuration file at configFilepath.
// The third return value (of type error) is not nil when the validation process itself fails, not if the validation ran succesfully with a result of "not validated".
func ValidateConfiguration(configFilepath string) (valid bool, validationErrors []gojsonschema.ResultError, err error) {
	// read file → unmarshal YAML → marshal JSON
	var configuration interface{}
	configContent, err := readFileBytes(configFilepath)
	if err != nil {
		return false, nil, err
	}
	yaml.Unmarshal(configContent, &configuration)
	json := jsoniter.ConfigFastest
	configurationDocument, _ := json.Marshal(configuration)
	valid, validationErrors, err = validateWithJSONSchema(string(configurationDocument), ConfigurationJSONSchema())
	return
}

// DefaultConfiguration returns a configuration with sensible defaults.

func DefaultConfiguration() Configuration {
	absoluteFilepathToHere, err := filepath.Abs(".")
	if err != nil {
		panic(fmt.Errorf("cannot get absolute path to current directory: %w", err))
	}

	return Configuration{
		ExtractColors: ExtractColorsConfiguration{
			Enabled: true,
		},
		MakeThumbnails: MakeThumbnailsConfiguration{
			Enabled:          true,
			Sizes:            []int{100, 400, 600, 1200},
			FileNameTemplate: "<work id>/<block id>@<size>.webp",
		},
		Media: struct{ At string }{
			At: "media/",
		},
		ScatteredModeFolder: DefaultScatteredModeFolder,
		IsDefault:           true,
		ProjectsDirectory:   absoluteFilepathToHere,
	}
}
