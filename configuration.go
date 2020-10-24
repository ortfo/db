package main

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/imdatngo/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/xeipuuv/gojsonschema"
	"github.com/mitchellh/colorstring"
	"gopkg.in/yaml.v2"
)

type configurationBuildStepsExtractColors struct {
	Enabled         bool
	Extract         []string
	DefaultFileName []string `yaml:"default file name"`
}

type configurationBuildStepsMakeGifs struct {
	Enabled          bool
	FileNameTemplate string `yaml:"file name template"`
}

type configurationBuildStepsMakeThumbnails struct {
	Enabled          bool
	Widths           []int
	InputFile        string `yaml:"input file"`
	FileNameTemplate string `yaml:"file name template"`
}

type configurationBuildSteps struct {
	ExtractColors  configurationBuildStepsExtractColors  `yaml:"extract colors"`
	MakeGifs       configurationBuildStepsMakeGifs       `yaml:"make GIFs"`
	MakeThumbnails configurationBuildStepsMakeThumbnails `yaml:"make thumbnails"`
}

type configurationMarkdownAnchoredHeadings struct {
	Enabled bool
	Format  string
}

type configurationMarkdownCustomSyntax struct {
	From string
	To   string
}

// Configuration represents what the .portfoliodb.yml configuration file describes
type Configuration struct {
	BuildSteps configurationBuildSteps `yaml:"build steps"`
	Features   struct {
		madeWith      bool `yaml:"made with"`
		mediaHoisting bool `yaml:"media hoisting"`
	}
	Validate struct {
		Checks struct {
			SchemaCompliance     string `yaml:"schema compliance"`
			WorkFolderUniqueness string `yaml:"work folder uniqueness"`
			WorkFolderSafeness   string `yaml:"work folder safeness"`
			YamlHeader           string `yaml:"yaml header"`
			TitlePresence        string `yaml:"title presence"`
			TitleUniqueness      string `yaml:"title uniqueness"`
			TagsPresence         string `yaml:"tags presence"`
			TagsKnowledge        string `yaml:"tags knowledge"`
			WorkingMedia         string `yaml:"working media"`
			WorkingUrls          string `yaml:"working urls"`
		}
	}
	Markdown struct {
		Abbreviations      bool                                  `yaml:"abbreviations"`
		DefinitionLists    bool                                  `yaml:"definition lists"`
		Admonitions        bool                                  `yaml:"admonitions"`
		Footnotes          bool                                  `yaml:"footnotes"`
		MarkdownInHTML     bool                                  `yaml:"markdown in html"`
		NewLineToLineBreak bool                                  `yaml:"new-line-to-line-break"`
		SmartyPants        bool                                  `yaml:"smarty pants"`
		AnchoredHeadings   configurationMarkdownAnchoredHeadings `yaml:"anchored headings"`
		CustomSyntaxes     []configurationMarkdownCustomSyntax   `yaml:"custom syntaxes"`
	}
}

// LoadConfiguration loads the .portfoliodb.yml file in ``databaseFolderPath`` and puts it contents into ``loadInto``.
func LoadConfiguration(filepath string, loadInto *Configuration) error {
	raw := ReadFileBytes(filepath)
	return yaml.Unmarshal(raw, loadInto)
}

// LoadDefaultConfiguration gets the default .portfoliodb.yml configuration (at ./.portfoliodb.yml) and puts it contents into ``loadInto``.
func LoadDefaultConfiguration(loadInto *Configuration) error {
	return LoadConfiguration("./.portfoliodb.yml", loadInto)
}

// GetConfiguration  reads from the .portfoliodb.yml file in ``databaseFolderPath``
// and returns a ``Configuration`` struct
func GetConfiguration(filepath string) (Configuration, error) {
	var userConfig Configuration
	var defaultConfig Configuration
	// Load the default configuration
	if err := LoadDefaultConfiguration(&defaultConfig); err != nil {
		return Configuration{}, err
	}
	// Load the user's configuration
	if err := LoadConfiguration(filepath, &userConfig); err != nil {
		return Configuration{}, err
	}
	// Then merge defaultConfig into userConfig, to fill out uninitialized fields
	if err := mergo.Merge(&userConfig, &defaultConfig); err != nil {
		return Configuration{}, err
	}
	return userConfig, nil
}

// ResolveConfigurationPath determines the path of the configuration file to use
func ResolveConfigurationPath(databaseDirectory string, explicitlySpecifiedConfigurationFilepath string) string {
	if explicitlySpecifiedConfigurationFilepath == "" {
		return path.Join(databaseDirectory, ".portfoliodb.yml")
	}
	return explicitlySpecifiedConfigurationFilepath
}

// ValidateConfiguration uses the JSON configuration schema ConfigurationJSONSchema to validate the configuration file at configFilepath
func ValidateConfiguration(configFilepath string) (bool, []gojsonschema.ResultError) {
	// read file → unmarshal YAML → marshal JSON
	var configuration interface{}
	yaml.Unmarshal(ReadFileBytes(configFilepath), &configuration)
	json := jsoniter.ConfigFastest
	configurationDocument, _ := json.Marshal(configuration)
	return ValidateWithJSONSchema(string(configurationDocument), ConfigurationJSONSchema)
}

// GetConfigurationFromCLIArgs gets the configuration by using the CLI arguments
func GetConfigurationFromCLIArgs(args docopt.Opts) (Configuration, []gojsonschema.ResultError, error) {
	// Weird bug if args.String("<database>") is used...
	databaseDirectory := args["<database>"].([]string)[0]
	explicitConfigFilepath, _ := args.String("--config")
	configFilepath := ResolveConfigurationPath(databaseDirectory, explicitConfigFilepath)
	configFilepath, err := filepath.Abs(configFilepath)
	if err != nil {
		panic(err)
	}
	validated, validationErrors := ValidateConfiguration(configFilepath)
	if !validated {
		return Configuration{}, validationErrors, nil
	}
	var config Configuration
	if err := LoadConfiguration(configFilepath, &config); err != nil {
		return Configuration{}, make([]gojsonschema.ResultError, 0), err
	}
	return config, make([]gojsonschema.ResultError, 0), nil
}

// DisplayValidationErrors takes in a slice of json schema validation errors and displays them nicely to in the terminal
func DisplayValidationErrors(errors []gojsonschema.ResultError) {
	println("Your configuration file is invalid. Here are the validation errors:\n")
	for _, err := range errors {
		colorstring.Println("- " + strings.ReplaceAll(err.Field(), ".", "[blue][bold]/[reset]"))
		colorstring.Println("    [red]" + err.Description())
	}
}
