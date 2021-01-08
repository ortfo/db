package portfoliodb

import (
	"path"
	"path/filepath"

	"github.com/docopt/docopt-go"
	"github.com/imdatngo/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/xeipuuv/gojsonschema"
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

type checks struct {
	SchemaCompliance     string `yaml:"schema compliance"`
	WorkFolderUniqueness string `yaml:"work folder uniqueness"`
	WorkFolderSafeness   string `yaml:"work folder safeness"`
	YamlHeader           string `yaml:"yaml header"`
	TitlePresence        string `yaml:"title presence"`
	TitleUniqueness      string `yaml:"title uniqueness"`
	WorkingUrls          string `yaml:"working urls"`
}

type replaceMediaSource struct {
	Replace string `yaml:"replace"`
	With    string `yaml:"with"`
}

// Configuration represents what the .portfoliodb.yml configuration file describes
type Configuration struct {
	configurationBuildSteps
	Checks              checks               `yaml:"checks"`
	ReplaceMediaSources []replaceMediaSource `yaml:"replace media sources"`
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

// LoadConfiguration loads the .portfoliodb.yml file in ``databaseFolderPath`` and puts it contents into ``loadInto``.
func LoadConfiguration(filepath string, loadInto *Configuration) error {
	raw, err := ReadFileBytes(filepath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(raw, loadInto)
}

// GetConfiguration  reads from the .portfoliodb.yml file in ``databaseFolderPath``
// and returns a ``Configuration`` struct
func GetConfiguration(filepath string) (Configuration, error) {
	var userConfig Configuration
	defaultConfig := Configuration{}
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
func ValidateConfiguration(configFilepath string) (bool, []gojsonschema.ResultError, error) {
	// read file → unmarshal YAML → marshal JSON
	var configuration interface{}
	configContent, err := ReadFileBytes(configFilepath)
	if err != nil {
		return false, nil, err
	}
	yaml.Unmarshal(configContent, &configuration)
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
		return Configuration{}, nil, err
	}
	validated, validationErrors, err := ValidateConfiguration(configFilepath)
	if err != nil {
		return Configuration{}, nil, err
	}
	if !validated {
		return Configuration{}, validationErrors, nil
	}
	var config Configuration
	if err := LoadConfiguration(configFilepath, &config); err != nil {
		return Configuration{}, make([]gojsonschema.ResultError, 0), err
	}
	return config, make([]gojsonschema.ResultError, 0), nil
}
