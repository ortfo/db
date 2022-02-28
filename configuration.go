package ortfodb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
	"unicode"

	jsoniter "github.com/json-iterator/go"
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
	Sizes            []uint16
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
			fmt.Println("No configuration file found. Using default configuration.")
			defaultConfig, err := yaml.Marshal(DefaultConfiguration())
			if err != nil {
				panic(err)
			}
			ioutil.WriteFile("ortfodb.yaml", []byte(defaultConfig), 0o644)
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
			Sizes:            []uint16{100, 400, 600, 1200},
			FileNameTemplate: "<media directory>/<work id>/<media id>@<size>.webp",
		},
		Media: struct{ At string }{
			At: "media/",
		},
		BuildMetadataFilepath: ".lastbuild.yaml",
		ScatteredModeFolder:   ".ortfo",
	}
}

// setJSONNamingStrategy rename struct fields uniformly.
func setJSONNamingStrategy(translate func(string) string) {
	jsoniter.RegisterExtension(&namingStrategyExtension{jsoniter.DummyExtension{}, translate})
}

type namingStrategyExtension struct {
	jsoniter.DummyExtension
	translate func(string) string
}

func (extension *namingStrategyExtension) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		if unicode.IsLower(rune(binding.Field.Name()[0])) || binding.Field.Name()[0] == '_' {
			continue
		}
		tag, hastag := binding.Field.Tag().Lookup("json")
		if hastag {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" {
				continue // hidden field
			}
			if tagParts[0] != "" {
				continue // field explicitly named
			}
		}
		binding.ToNames = []string{extension.translate(binding.Field.Name())}
		binding.FromNames = []string{extension.translate(binding.Field.Name())}
	}
}

// lowerCaseWithUnderscores one strategy to SetNamingStrategy for. It will change HelloWorld to hello_world.
func lowerCaseWithUnderscores(name string) string {
	// Handle acronyms
	if isAllUpper(name) {
		return strings.ToLower(name)
	}
	newName := []rune{}
	for i, c := range name {
		if i == 0 {
			newName = append(newName, unicode.ToLower(c))
		} else {
			if c == ' ' {
				newName = append(newName, '_')
			} else if unicode.IsUpper(c) {
				newName = append(newName, '_')
				newName = append(newName, unicode.ToLower(c))
			} else {
				newName = append(newName, c)
			}
		}
	}
	return string(newName)
}

func isAllUpper(s string) bool {
	for _, c := range s {
		if !unicode.IsUpper(c) {
			return false
		}
	}
	return true
}
