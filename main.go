package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	confluence "github.com/virtomize/confluence-go-api"
)

type Output struct {
	Items         []ScriptFilterItem `json:"items"`
	Variables     map[string]string  `json:"variables,omitzero"`
	Rerun         int                `json:"rerun,omitzero"`
	Cache         Cache              `json:"cache,omitzero"`
	SkipKnowledge bool               `json:"skipknowledge,omitzero"`
}

type Cache struct {
	Seconds     int  `json:"seconds"`
	LooseReload bool `json:"loosereload"`
}

/*
ScriptFilterItem contains the structure for an Alfred script filter item
See https://www.alfredapp.com/help/workflows/inputs/script-filter/json/ for more information
*/
type ScriptFilterItem struct {
	Uid          string            `json:"uid,omitzero"`
	Title        string            `json:"title"`
	Subtitle     string            `json:"subtitle,omitzero"`
	Arg          []string          `json:"arg,omitzero"`
	Icon         Icon              `json:"icon,omitzero"`
	Valid        bool              `json:"valid,omitzero"`
	Match        string            `json:"match,omitzero"`
	Autocomplete string            `json:"autocomplete,omitzero"`
	Type         string            `json:"type,omitzero"`
	Mods         map[string]Mod    `json:"mods,omitzero"`
	Action       Action            `json:"action,omitzero"`
	Text         map[string]string `json:"text,omitzero"`
	QuickLookURL string            `json:"quicklookurl,omitzero"`
}

type Icon struct {
	Type string `json:"type,omitzero"`
	Path string `json:"path"`
}

type Mod struct {
	Valid    bool   `json:"valid"`
	Arg      string `json:"arg"`
	Subtitle string `json:"subtitle"`
}

type Action struct {
	Text string `json:"text"`
	Url  string `json:"url,omitzero"`
	File string `json:"file,omitzero"`
	Auto string `json:"auto,omitzero"`
}

func main() {
	token := flag.String("token", "", "API token (required)")
	baseURL := flag.String("url", "", "Base URL (required)")
	username := flag.String("username", "", "confluence username (required)")
	query := flag.String("query", "", "search query to execute (required)")

	flag.Parse()
	if *token == "" || *baseURL == "" || *username == "" || *query == "" {
		flag.Usage()
		os.Exit(1)
	}

	api, err := confluence.NewAPI(fmt.Sprintf("%s/wiki/rest/api", *baseURL), *username, *token)
	if err != nil {
		fatal("could not connect to api", err)
	}

	cql := confluence.SearchQuery{
		CQL:                   fmt.Sprintf("type=page and title~\"%s\"", *query),
		IncludeArchivedSpaces: false,
		Limit:                 25,
	}

	result, err := api.Search(cql)
	if err != nil {
		fatal("couldn't fetch results", err)
	}

	items := []ScriptFilterItem{}
	for _, v := range result.Results {
		pageUrl := fmt.Sprintf("%s/wiki%s", *baseURL, v.URL)
		page := ScriptFilterItem{
			Uid:          v.ID,
			Title:        v.Content.Title,
			Subtitle:     v.Content.Space.Key,
			Arg:          []string{pageUrl},
			Icon:         Icon{Path: "icon.png"},
			Valid:        true,
			Match:        v.Content.Title,
			Autocomplete: v.Content.Title,
			Type:         "default",
			Text: map[string]string{
				"copy":     pageUrl,
				"lagetype": v.Content.Title,
			},
			QuickLookURL: pageUrl,
		}
		items = append(items, page)
	}
	output := Output{
		Items: items,
		Cache: Cache{
			Seconds:     3600,
			LooseReload: true,
		},
	}
	out(output)
}

func fatal(why string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v", why, err)
	os.Exit(1)
}

func out(output Output) {
	jsonBytes, err := json.MarshalIndent(output, "", " ")
	if err != nil {
		fatal("Error marshalling output to json", err)
	}
	fmt.Println(string(jsonBytes))
}
