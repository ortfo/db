package ortfodb

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/go-homedir"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

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
	MakeGifs       MakeGIFsConfiguration       `yaml:"make GIFs"`
	MakeThumbnails MakeThumbnailsConfiguration `yaml:"make thumbnails"`
}

type BuildMetadata struct {
	PreviousBuildDate time.Time
}

// Configuration represents what the ortfodb.yaml configuration file describes.
type Configuration struct {
	ExtractColors         ExtractColorsConfiguration  `yaml:"extract colors"`
	MakeGifs              MakeGIFsConfiguration       `yaml:"make GIFs"`
	MakeThumbnails        MakeThumbnailsConfiguration `yaml:"make thumbnails"`
	BuildMetadataFilepath string                      `yaml:"build metadata file"`
	Media                 struct{ At string }         `yaml:"media"`
	ScatteredModeFolder   string                      `yaml:"scattered mode folder"`
	// Signals whether the configuration was instanciated by DefaultConfiguration.
	IsDefault bool `yaml:"-"`
	// Markdown struct {
	// 	Abbreviations      bool                                  `yaml:"abbreviations"`
	// 	DefinitionLists    bool                                  `yaml:"definition lists"`
	// 	Admonitions        bool                                  `yaml:"admonitions"`
	// 	Footnotes          bool                                  `yaml:"footnotes"`
	// 	MarkdownInHTML     bool                                  `yaml:"markdown in html"`
	// 	NewLineToLineBreak bool                                  `yaml:"new-line-to-line-break"`
	// 	SmartyPants        bool                                  `yaml:"smarty pants"`
	// 	AnchoredHeadings   configurationMarkdownAnchoredHeadings `yaml:"anchored headings"`
	// 	CustomSyntaxes     []configurationMarkdownCustomSyntax   `yaml:"custom syntaxes"`
	// }
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
	return nil
}

// NewConfiguration loads a YAML configuration file.
// If filepath is empty, the path defaults to databaseDirectory/ortfodb.yaml.
// This function also validates the configuration and prints any error to the user.
// Use LoadConfiguration for a lower-level function that just loads the YAML file into a struct.
func NewConfiguration(filename string, databaseDirectory string) (Configuration, error) {
	if filename == "" {
		filename = path.Join(databaseDirectory, "ortfodb.yaml")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			defaultConfig, err := yaml.Marshal(DefaultConfiguration())
			if err != nil {
				panic(err)
			}
			os.WriteFile("ortfodb.yaml", []byte(defaultConfig), 0o644)
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

	config := Configuration{}
	err = LoadConfiguration(filename, &config)

	// Set default value for ScatteredModeFolder
	if config.ScatteredModeFolder == "" {
		config.ScatteredModeFolder = ".ortfo"
	}

	// Remove trailing slash(es) from folder name.
	config.ScatteredModeFolder = strings.TrimRight(config.ScatteredModeFolder, "/\\")

	// Expand ~
	config.MakeThumbnails.FileNameTemplate, err = homedir.Expand(config.MakeThumbnails.FileNameTemplate)
	config.Media.At, err = homedir.Expand(config.Media.At)
	config.BuildMetadataFilepath, err = homedir.Expand(config.BuildMetadataFilepath)

	return config, err
}

// ValidateConfiguration uses the JSON configuration schema ConfigurationJSONSchema to validate the configuration file at configFilepath.
// The third return value (of type error) is not nil when the validation process itself fails, not if the validation ran succesfully with a result of "not validated".
func ValidateConfiguration(configFilepath string) (bool, []gojsonschema.ResultError, error) {
	// read file → unmarshal YAML → marshal JSON
	var configuration interface{}
	configContent, err := readFileBytes(configFilepath)
	if err != nil {
		return false, nil, err
	}
	yaml.Unmarshal(configContent, &configuration)
	json := jsoniter.ConfigFastest
	configurationDocument, _ := json.Marshal(configuration)
	return validateWithJSONSchema(string(configurationDocument), configurationJSONSchema)
}

// DefaultConfiguration returns a configuration with sensible defaults.
func DefaultConfiguration() Configuration {
	return Configuration{
		ExtractColors: ExtractColorsConfiguration{
			Enabled: true,
		},
		MakeThumbnails: MakeThumbnailsConfiguration{
			Enabled:          true,
			Sizes:            []int{100, 400, 600, 1200},
			FileNameTemplate: "<media directory>/<work id>/<block id>@<size>.webp",
		},
		Media: struct{ At string }{
			At: "media/",
		},
		BuildMetadataFilepath: ".lastbuild.yaml",
		ScatteredModeFolder:   ".ortfo",
		IsDefault:             true,
	}
}
