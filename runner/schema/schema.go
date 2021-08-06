package schema

import (
	"context"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/figment-networks/graph-demo/graphcall"
	"github.com/figment-networks/graph-demo/runner/store"
	"github.com/figment-networks/graph-demo/runner/structs"

	"gopkg.in/yaml.v2"
)

var (
	entityRegxp = regexp.MustCompile("type\\s+([^\\s]+)\\s+\\@entity\\s+{([^\\}]+)}")
	kvRegxp     = regexp.MustCompile("\\s+([a-zA-Z0-9]+):\\s+([\\[\\]a-zA-Z0-9]+)!?")
)

type GQLCaller interface {
	Subscribe(ctx context.Context, name string, events []structs.Subs) error
}

type JSLoader interface {
	LoadJS(name string, path string, ehs map[string]string) error
}

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
	File    string             `yaml:"file"`
	Network string             `yaml:"network"`
	Mapping DataSourcesMapping `yaml:"mapping"`
	Source  DataSourcesSource  `yaml:"source"`
}

type DataSourcesMapping struct {
	Kind          string          `yaml:"kind"`
	EventHandlers []EventHandlers `yaml:"eventHandlers"`
}

type DataSourcesSource struct {
	StartBlock uint64 `yaml:"startBlock"`
}

type EventHandlers struct {
	Event   string `yaml:"event"`
	Handler string `yaml:"handler"`
}

type Schemas struct {
	ss     store.Storage
	rqstr  GQLCaller
	loader JSLoader
}

func NewSchemas(ss store.Storage, loader JSLoader, rqstr GQLCaller) *Schemas {
	return &Schemas{
		ss:     ss,
		loader: loader,
		rqstr:  rqstr,
	}
}

func (s *Schemas) LoadFromSubgraphYaml(fpath string) error {

	f, err := ioutil.ReadFile(path.Join(fpath, "subgraph.yaml"))
	if err != nil {
		return err
	}

	m := &Manifest{}
	if err = yaml.Unmarshal(f, &m); err != nil {
		return err
	}

	paths := strings.Split(fpath, "/")
	name := paths[len(paths)-1]
	subg, err := processSchema(path.Join(fpath, m.Schema.File), name)
	if err != nil {
		return err
	}

	for _, ent := range subg.Entities {
		indexed := []store.NT{}
		for k, v := range ent.Fields {
			indexed = append(indexed, store.NT{Name: k, Type: v.Type})
		}
		s.ss.NewStore(name, ent.Name, indexed)
	}

	for _, sourc := range m.Sources {
		subs := []structs.Subs{}
		ms := make(map[string]string)

		for _, evh := range sourc.Mapping.EventHandlers {
			subs = append(subs, structs.Subs{Name: evh.Event, StartingHeight: sourc.Source.StartBlock})
			ms[evh.Event] = evh.Handler
		}

		ctxWithTimeout, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer ctxCancel()

		if err := s.rqstr.Subscribe(ctxWithTimeout, sourc.Network, subs); err != nil {
			return err
		}

		if err := s.loader.LoadJS(name, path.Join(fpath, sourc.File), ms); err != nil {
			return err
		}

	}

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
