package main

import (
	"path"

	"github.com/davecgh/go-spew/spew"
	"github.com/docopt/docopt-go"
	"github.com/imdatngo/mergo"
	"gopkg.in/yaml.v2"
)

type ConfigurationBuildStepsExtractColors struct {
	Enabled         bool
	Extract         []string
	DefaultFileName []string `yaml:"default file name"`
}

type ConfigurationBuildStepsMakeGifs struct {
	Enabled          bool
	FileNameTemplate string `yaml:"file name template"`
}

type ConfigurationBuildStepsMakeThumbnails struct {
	Enabled          bool
	Widths           []int
	InputFile        string `yaml:"input file"`
	FileNameTemplate string `yaml:"file name template"`
}

type ConfigurationBuildSteps struct {
	ExtractColors  ConfigurationBuildStepsExtractColors  `yaml:"extract colors"`
	MakeGifs       ConfigurationBuildStepsMakeGifs       `yaml:"make GIFs"`
	MakeThumbnails ConfigurationBuildStepsMakeThumbnails `yaml:"make thumbnails"`
}

type ConfigurationMarkdownAnchoredHeadings struct {
	Enabled bool
	Format  string
}

type ConfigurationMarkdownCustomSyntax struct {
	From string
	To   string
}

// Configuration represents what the .portfoliodb.yml configuration file describes
type Configuration struct {
	BuildSteps ConfigurationBuildSteps `yaml:"build steps"`
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
		AnchoredHeadings   ConfigurationMarkdownAnchoredHeadings `yaml:"anchored headings"`
		CustomSyntaxes     []ConfigurationMarkdownCustomSyntax   `yaml:"custom syntaxes"`
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
	spew.Dump(userConfig, defaultConfig)
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

// GetConfigurationFromCLIArgs gets the configuration by using the CLI arguments
func GetConfigurationFromCLIArgs(args docopt.Opts) (Configuration, error) {
	// Weird bug if args.String("<database>") is used...
	databaseDirectory := args["<database>"].([]string)[0]
	explicitConfigFilepath, _ := args.String("--config")
	configFilepath := ResolveConfigurationPath(databaseDirectory, explicitConfigFilepath)
	println("loading config file at:", configFilepath)
	var config Configuration
	if err := LoadConfiguration(configFilepath, &config); err != nil {
		return Configuration{}, err
	}
	return config, nil
}
