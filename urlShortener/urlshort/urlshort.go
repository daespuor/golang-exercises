package urlshort

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"urlShortener/services"

	"gopkg.in/yaml.v2"
)

type URLMapping struct {
	ShortURL string `yaml:"path" json:"path"`
	LongURL  string `yaml:"url" json:"url"`
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortUrl := r.URL.Path
		longUrl, ok := pathsToUrls[shortUrl]
		if ok {
			http.Redirect(w, r, longUrl, http.StatusPermanentRedirect)
		}
		fallback.ServeHTTP(w, r)
	})
}

type FileHandler interface {
	Handle(fallback http.Handler) (http.HandlerFunc, error)
}

type YAMLHandler struct {
	content []byte
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//   - path: /some-path
//     url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func (h YAMLHandler) Handle(fallback http.Handler) (http.HandlerFunc, error) {
	yamlParsed, err := parseYAML(h.content)
	if err != nil {
		return nil, err
	}
	resultMap := buildMap(yamlParsed)
	return MapHandler(resultMap, fallback), nil
}

func parseYAML(yml []byte) ([]URLMapping, error) {
	content := make([]URLMapping, 0)
	err := yaml.Unmarshal(yml, &content)
	if err != nil {
		return nil, fmt.Errorf("error parsing the yaml file %w", err)
	}
	return content, nil
}

func buildMap(content []URLMapping) map[string]string {
	resultMap := make(map[string]string)
	for _, u := range content {
		resultMap[u.ShortURL] = u.LongURL
	}
	return resultMap
}

type JSONHandler struct {
	content []byte
}

func (h JSONHandler) Handle(fallback http.Handler) (http.HandlerFunc, error) {
	content, err := parseJSON(h.content)
	if err != nil {
		return nil, err
	}

	resultMap := buildMap(content)
	return MapHandler(resultMap, fallback), nil
}

func parseJSON(data []byte) ([]URLMapping, error) {
	var mapping []URLMapping
	err := json.Unmarshal(data, &mapping)
	if err != nil {
		return nil, fmt.Errorf("error parsing the json file %w", err)
	}

	return mapping, nil
}

type DBHandler struct {
	s *services.URLService
}

func NewDBHandler(s *services.URLService) DBHandler {
	return DBHandler{s: s}
}

func (h DBHandler) Handle(fallback http.Handler) (http.HandlerFunc, error) {
	content := make([]URLMapping, 0)
	contentDTO, err := h.s.GetAllMappings()
	if err != nil {
		return nil, err
	}
	for _, dto := range contentDTO {
		content = append(content, URLMapping{ShortURL: dto.ShortUrl, LongURL: dto.LongUrl})
	}

	resultMap := buildMap(content)
	return MapHandler(resultMap, fallback), nil
}

func NewHandler(filepath string, content []byte) FileHandler {

	if strings.HasSuffix(filepath, ".yaml") {
		return YAMLHandler{content}
	} else if strings.HasSuffix(filepath, ".json") {
		return JSONHandler{content}
	}

	return nil
}
