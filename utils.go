package ortfodb

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// readFileBytes reads the content of filename and returns the contents as a byte array.
func readFileBytes(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

// readFile reads the content of filename and returns the contents as a string.
func readFile(filename string) (string, error) {
	content, err := readFileBytes(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// writeFile writes content to file filepath.
func writeFile(filename string, content []byte) error {
	absfilepath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	f, err := os.Create(absfilepath)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

// validateWithJSONSchema checks if the JSON document document conforms to the JSON schema schema.
func validateWithJSONSchema(document string, schema string) (bool, []gojsonschema.ResultError, error) {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(document)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return false, nil, err
	}

	if result.Valid() {
		return true, nil, nil
	}
	return false, result.Errors(), nil
}

// fileExists checks if the given file exists, and returns true if it exists or false otherwise.
func fileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

// regexpMatches checks if s matches the regex regex at least once.
func regexpMatches(regex string, s string) bool {
	p := regexp.MustCompile(regex)
	return p.MatchString(s)
}

// regexpGroups returns all the capture groups' contents from the first match of regex regex in s. The first element [0] is the entire match. [1] is the first capture group's content, et cætera.
func regexpGroups(regex string, s string) []string {
	p := regexp.MustCompile(regex)
	return p.FindStringSubmatch(s)
}

// isValidURL tests a string to determine if it is a well-structured url or not.
func isValidURL(URL string) bool {
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return false
	}

	u, err := url.Parse(URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// stringInSlice checks if needle is in haystack.
func stringInSlice(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// filterSlice returns a slice of strings containing only the elements that return true when called with cond.
func filterSlice(s []string, cond func(string) bool) []string {
	filtered := make([]string, 0)
	for _, item := range s {
		if cond(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// mapKeys returns a slice of strings containing the map's keys.
func mapKeys(m map[string]string) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// filepathBaseNoExt returns the basename of pth with the extension removed.
func filepathBaseNoExt(pth string) string {
	return strings.TrimSuffix(filepath.Base(pth), path.Ext(pth))
}
