package schema

import (
	"io/ioutil"
	"path"
	"regexp"

	"github.com/figment-networks/graph-demo/graphcall"
	"github.com/figment-networks/graph-demo/runner/store"

	"gopkg.in/yaml.v2"
)

var (
	entityRegxp = regexp.MustCompile("type\\s+([^\\s]+)\\s+\\@entity\\s+{([^\\}]+)}")
	kvRegxp     = regexp.MustCompile("\\s+([a-zA-Z0-9]+):\\s+([\\[\\]a-zA-Z0-9]+)!?")
)

type Manifest struct {
	Description string         `yaml:"description"`
	Schema      ManifestSchema `yaml:"schema"`
	Sources     []DataSources  `yaml:"dataSources"`
}

type ManifestSchema struct {
	File string `yaml:"file"`
}

type DataSources struct {
	Kind    string             `yaml:"kind"`
	Name    string             `yaml:"name"`
	Network string             `yaml:"network"`
	Mapping DataSourcesMapping `yaml:"mapping"`
}

type DataSourcesMapping struct {
	Kind          string          `yaml:"kind"`
	EventHandlers []EventHandlers `yaml:"eventHandlers"`
}

type EventHandlers struct {
	Event   string `yaml:"event"`
	Handler string `yaml:"handler"`
}

type Schemas struct {
	ss        store.Storage
	Subgraphs map[string]*graphcall.Subgraph
}

func NewSchemas(ss store.Storage) *Schemas {
	return &Schemas{
		ss:        ss,
		Subgraphs: make(map[string]*graphcall.Subgraph),
	}
}

func (s *Schemas) LoadFromSubgraphYaml(fpath string) error {

	f, err := ioutil.ReadFile(path.Join(fpath, "subgraph.yaml"))
	if err != nil {
		return nil
	}
	m := &Manifest{}
	err = yaml.Unmarshal(f, &m)
	if err != nil {
		return nil
	}


	/*
		dir, err := os.ReadDir(path)
		for _, fse := range dir {
			if strings.HasSuffix(fse.Name(), ".graphQL") {

			}
		}
		s.Subgraphs[name] = sg
	*/
	return nil

}

func processSchema(filepath, name string) (*graphcall.Subgraph, error) {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	sg, err := graphcall.ParseSchema(name, f)
	if err != nil {
		return nil, err
	}
	return sg, nil
}
