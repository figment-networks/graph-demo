package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/figment-networks/graph-demo/runner/store"
	"github.com/figment-networks/graph-demo/runner/structs"
	"go.uber.org/zap"

	"rogchap.com/v8go"
)

var (
	regxp1 = regexp.MustCompile(`([^=[:space:]\\{]*)graphql.call`)
	regxp2 = regexp.MustCompile(`([^=[:space:]\\{]*)log.debug`)
	regxp3 = regexp.MustCompile(`([^=[:space:]\\{]*)store.save`)
)

type GQLCaller interface {
	CallGQL(ctx context.Context, name, query string, variables map[string]interface{}, version string) ([]byte, error)

	Subscribe(ctx context.Context, name string, events []structs.Subs) error
	Unsubscribe(ctx context.Context, name string, events []string) error
}

type callback func(info *v8go.FunctionCallbackInfo) *v8go.Value

type Loader struct {
	subgraphs map[string]*Subgraph

	events map[string]map[string][]*Subgraph

	lock  sync.RWMutex
	rqstr GQLCaller
	stor  store.Storage
	log   *zap.Logger

	data map[string]interface{}
}

func NewLoader(l *zap.Logger, rqstr GQLCaller, stor store.Storage) *Loader {
	return &Loader{
		subgraphs: make(map[string]*Subgraph),
		events:    make(map[string]map[string][]*Subgraph),
		rqstr:     rqstr,
		stor:      stor,
		log:       l,
	}
}

func (l *Loader) CallSubgraphHandler(subgraph string, handler *SubgraphHandler) error {

	l.log.Debug("Calling SubgraphHandler ", zap.String("subgraph", subgraph))
	//	l.lock.RLock()
	s, ok := l.subgraphs[subgraph]
	//	l.lock.RUnlock()

	if !ok {
		return errors.New("subgraph not found")
	}

	e, err := handler.EncodeString()
	if err != nil {
		return err
	}
	_, err = s.context.RunScript(e, "mapping.js")
	return err
}

func (l *Loader) NewEvent(typ string, data map[string]interface{}) error {
	l.log.Debug("Event received ", zap.String("type", typ), zap.Any("data", data))
	l.lock.RLock()
	defer l.lock.RUnlock()
	for handler, subgs := range l.events[typ] {
		for _, sgs := range subgs {
			if err := l.CallSubgraphHandler(sgs.Name, &SubgraphHandler{name: handler, values: []interface{}{data}}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Loader) LoadJS(name string, path string, evH map[string]string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return l.createRunable(name, b, evH)
}

func (l *Loader) createRunable(name string, code []byte, evH map[string]string) error {

	subgr := NewSubgraph(name, l.rqstr, l.stor)
	iso, _ := v8go.NewIsolate()

	callGQL, _ := v8go.NewFunctionTemplate(iso, subgr.callGQL)
	storeRecord, _ := v8go.NewFunctionTemplate(iso, subgr.storeRecord)

	logDebug, _ := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		l.log.Debug("v8LogDebug", zap.Any("args", info.Args()))
		return nil
	})

	global, err := v8go.NewObjectTemplate(iso)
	if err != nil {
		return err
	}
	global.Set("v8LogDebug", logDebug)
	global.Set("v8Call", callGQL)
	global.Set("v8StoreSave", storeRecord)

	subgr.context, err = v8go.NewContext(iso, global)
	if err != nil {
		return err
	}
	_, err = subgr.context.RunScript(cleanJS(code), "main.js")
	if err != nil {
		return err
	}

	l.lock.Lock()
	l.subgraphs[name] = subgr
	for event, handler := range evH {
		e, ok := l.events[event]
		if !ok {
			e = make(map[string][]*Subgraph)
		}
		h, ok := e[handler]
		if !ok {
			h = []*Subgraph{}
		}

		h = append(h, subgr)
		e[handler] = h
		l.events[event] = e
	}
	l.lock.Unlock()

	l.log.Debug("Loaded subgraph", zap.String("name", name), zap.Any("data", l.subgraphs))
	return nil
}

// cleanJS - To mimic assemblyscript that currently runs webassembly
// we need to change some function names
func cleanJS(code []byte) string {
	b := strings.Builder{}
	for _, l := range strings.Split(string(code), "\n") {
		if !strings.Contains(l, "require(") && !strings.Contains(l, "exports") {
			b.WriteString(l)
			b.WriteString("\n")
		}
	}
	res1 := regxp1.ReplaceAllString(b.String(), " v8Call")
	res2 := regxp2.ReplaceAllString(res1, " v8LogDebug")
	return regxp3.ReplaceAllString(res2, " v8StoreSave")
}

type Subgraph struct {
	Name string
	body []byte

	callbacks map[string]callback //name - callback

	caller  GQLCaller
	stor    store.Storage
	context *v8go.Context
}

func NewSubgraph(name string, caller GQLCaller, stor store.Storage) *Subgraph {
	return &Subgraph{
		Name:      name,
		caller:    caller,
		stor:      stor,
		callbacks: make(map[string]callback),
	}
}

func (s *Subgraph) storeRecord(info *v8go.FunctionCallbackInfo) *v8go.Value {
	args := info.Args()

	mj, _ := args[1].MarshalJSON()
	a := map[string]interface{}{}
	_ = json.Unmarshal(mj, &a)

	if err := s.stor.Store(context.Background(), s.Name, `"`+args[0].String()+`"`, a); err != nil {
		return jsonError(info.Context(), err)
	}

	return nil

}

func (s *Subgraph) callGQL(info *v8go.FunctionCallbackInfo) *v8go.Value {
	args := info.Args()

	mj, err := json.Marshal(args[2])
	if err != nil {
		log.Println(fmt.Printf("parameters error %v \n", err))
	}

	a := map[string]interface{}{}
	err = json.Unmarshal(mj, &a)
	if err != nil {
		log.Println(fmt.Errorf("marshal error %w \n", err))
	}
	resp, err := s.caller.CallGQL(context.Background(), args[0].String(), args[1].String(), a, args[2].String())

	if err != nil {
		log.Println(fmt.Printf("callGQL error %v \n", err))
		return jsonError(info.Context(), err)
	}

	p, _ := v8go.JSONParse(info.Context(), string(resp))
	return p
}

func jsonError(ctx *v8go.Context, err error) *v8go.Value {
	erro, _ := v8go.JSONParse(ctx, "{\"error\":{\"message\":\""+strings.ReplaceAll(err.Error(), "\"", "\\\"")+"\"}}")
	return erro
}

type SubgraphHandler struct {
	name   string
	values []interface{}
}

func (sh *SubgraphHandler) EncodeString() (string, error) {
	sb := strings.Builder{}
	sb.WriteString(sh.name)
	sb.WriteString("(")
	for i, v := range sh.values {
		if i > 0 {
			sb.WriteString(",")
		}
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		sb.Write(b)
	}
	sb.WriteString(")")
	return sb.String(), nil
}
