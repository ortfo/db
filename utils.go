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

// ReadFileBytes reads the content of ``filename`` and returns the contents as a byte array
func ReadFileBytes(filename string) ([]byte, error) {
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

// ReadFile reads the content of ``filename`` and returns the contents as a string
func ReadFile(filename string) (string, error) {
	content, err := ReadFileBytes(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile writes content to file filepath
func WriteFile(filename string, content []byte) error {
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

// ValidateWithJSONSchema checks if the JSON document ``document`` conforms to the JSON schema ``schema``
func ValidateWithJSONSchema(document string, schema string) (bool, []gojsonschema.ResultError, error) {
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

// FileExists checks if the given file exists, and returns ``true`` if it exists or ``false`` otherwise
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

// RegexpMatches checks if s matches the regex regex at least once
func RegexpMatches(regex string, s string) bool {
	p := regexp.MustCompile(regex)
	return p.MatchString(s)
}

// RegexpGroups returns all the capture groups' contents from the first match of regex regex in s. The first element [0] is the entire match. [1] is the first capture group's content, et cætera.
func RegexpGroups(regex string, s string) []string {
	p := regexp.MustCompile(regex)
	return p.FindStringSubmatch(s)
}

// IsValidURL tests a string to determine if it is a well-structured url or not.
func IsValidURL(URL string) bool {
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

// StringInSlice checks if needle is in haystack
func StringInSlice(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// FilterSlice returns a slice of strings containing only the elements that return true when called with cond.
func FilterSlice(s []string, cond func(string) bool) []string {
	filtered := make([]string, 0)
	for _, item := range s {
		if cond(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// MapKeys returns a slice of strings containing the map's keys
func MapKeys(m map[string]string) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// FilepathBaseNoExt returns the basename of pth with the extension removed
func FilepathBaseNoExt(pth string) string {
	return strings.TrimSuffix(filepath.Base(pth), path.Ext(pth))
}
