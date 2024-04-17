package ortfodb

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// readFileBytes reads the content of filename and returns the contents as a byte array.
func readFileBytes(filename string) ([]byte, error) {
	b, err := os.ReadFile(filename)
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
	defer f.Close()

	_, err = f.Write(content)
	if err != nil {
		return err
	}
	return nil
}

// validateWithJSONSchema checks if the JSON document document conforms to the JSON schema schema.
func validateWithJSONSchema(document string, schema *jsonschema.Schema) (bool, []gojsonschema.ResultError, error) {
	schemaJson, err := schema.MarshalJSON()
	if err != nil {
		panic(err)
	}
	schemaLoader := gojsonschema.NewStringLoader(string(schemaJson))
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
func mapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mapValues[T any](m map[string]T) []T {
	values := make([]T, 0)
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// filepathBaseNoExt returns the basename of pth with the extension removed.
func filepathBaseNoExt(pth string) string {
	return strings.TrimSuffix(filepath.Base(pth), path.Ext(pth))
}

// merge merges the given maps. Conflicting keys are overwritten by the values of the latest map with that key.
func merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// some returns true if predicate evaluates to true on any of the haystack elements
func some[T any](haystack []T, predicate func(T) bool) bool {
	for _, v := range haystack {
		if predicate(v) {
			return true
		}
	}
	return false
}

// all returns true if predicate evaluates to true on all of the haystack elements
func all[T any](haystack []T, predicate func(T) bool) bool {
	for _, v := range haystack {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// noDuplicates removes duplicate elements from the given slice, keeping only the first occurences.
func noDuplicates[T comparable](s []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0)
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func handleControlC(action func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			action()
		}
	}()
}

func stringsLooselyMatch(s string, needles ...string) bool {
	for _, needle := range needles {
		if strings.EqualFold(s, needle) {
			return true
		}
	}
	return false
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return contents, nil
}

func ensureHttpPrefix(url string) string {
	if !isValidURL(url) {
		if strings.HasPrefix(url, "localhost") || strings.HasPrefix(url, "127.0.0.") {
			return "http://" + url
		} else {
			return "https://" + url
		}
	}
	return url
}

func debugging() bool {
	return os.Getenv("DEBUG") == "1" || os.Getenv("ORTFO_DEBUG") == "1" || os.Getenv("ORTFODB_DEBUG") == "1"
}

func cgoEnabled() bool {
	return os.Getenv("CGO_ENABLED") == "1"
}

func writeYAML(v any, filename string) error {
	encoded, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("while encoding to YAML: %w", err)
	}
	err = os.WriteFile(filename, encoded, 0o644)
	if err != nil {
		return fmt.Errorf("while writing encoded yaml contents to %q: %w", filename, err)
	}
	return nil
}
