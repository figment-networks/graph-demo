package jsRuntime

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

type GQLCaller interface {
	CallGQL(ctx context.Context, name, query string, variables map[string]interface{}, version string) ([]byte, error)

	Subscribe(ctx context.Context, name string, events []structs.Subs) error
	Unsubscribe(ctx context.Context, name string, events []string) error
}

type callback func(info *v8go.FunctionCallbackInfo) *v8go.Value

type Loader struct {
	subgraphs map[string]*Subgraph
	lock      sync.RWMutex
	rqstr     GQLCaller
	stor      store.Storage
	log       *zap.Logger
}

func NewLoader(l *zap.Logger, rqstr GQLCaller, stor store.Storage) *Loader {
	return &Loader{
		subgraphs: make(map[string]*Subgraph),
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
	_, err = s.context.RunScript(e, "main.js")
	return err
}

func (l *Loader) NewEvent(typ string, data map[string]interface{}) error {
	l.log.Debug("Event received ", zap.String("type", typ), zap.Any("data", data))

	l.log.Debug("subgraphs ", zap.Any("data", l.subgraphs))
	l.lock.RLock()
	for name := range l.subgraphs {
		// TODO(l): Add dictionary to map events
		if err := l.CallSubgraphHandler(name, &SubgraphHandler{name: strings.Replace(typ, "new", "handle", -1), values: []interface{}{data}}); err != nil {
			return err
		}
	}
	l.lock.RUnlock()
	return nil
}

func (l *Loader) LoadJS(name string, path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// TODO(l): load it from subgraph.yaml
	l.rqstr.Subscribe(context.Background(), "cosmos",
		[]structs.Subs{
			{"newTransaction", 0},
			{"newBlock", 0}})

	return l.createRunable(name, b)
}

func (l *Loader) createRunable(name string, code []byte) error {

	subgr := NewSubgraph(name, l.rqstr, l.stor)
	iso, _ := v8go.NewIsolate()

	callGQL, _ := v8go.NewFunctionTemplate(iso, subgr.callGQL)
	storeRecord, _ := v8go.NewFunctionTemplate(iso, subgr.storeRecord)

	logDebug, _ := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		fmt.Printf("v8LogDebug:  %+v  \n", info.Args()[0])
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
	l.lock.Unlock()

	l.log.Debug("Loaded subgraph", zap.String("name", name), zap.Any("data", l.subgraphs))

	return nil
}

func cleanJS(code []byte) string {
	b := strings.Builder{}
	for _, l := range strings.Split(string(code), "\n") {
		if !strings.Contains(l, "require(") && !strings.Contains(l, "exports") {
			b.WriteString(l)
			b.WriteString("\n")
		}
	}
	m1 := regexp.MustCompile(`([^=[:space:]\\{]*)graphql.call`)
	res1 := m1.ReplaceAllString(b.String(), " v8Call")

	m3 := regexp.MustCompile(`([^=[:space:]\\{]*)log.debug`)
	res2 := m3.ReplaceAllString(res1, " v8LogDebug")

	m2 := regexp.MustCompile(`([^=[:space:]\\{]*)store.save`)
	a := m2.ReplaceAllString(res2, " v8StoreSave")

	return a

}

type Subgraph struct {
	name string
	body []byte

	callbacks map[string]callback //name - callback
	dependsON []string

	caller  GQLCaller
	stor    store.Storage
	context *v8go.Context
}

func NewSubgraph(name string, caller GQLCaller, stor store.Storage) *Subgraph {
	return &Subgraph{
		name:      name,
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

	if err := s.stor.Store(context.Background(), s.name, `"`+args[0].String()+`"`, a); err != nil {
		return jsonError(info.Context(), err)
	}

	return nil

}

func (s *Subgraph) callGQL(info *v8go.FunctionCallbackInfo) *v8go.Value {
	args := info.Args()

	mj, _ := args[2].MarshalJSON()

	a := map[string]interface{}{}
	_ = json.Unmarshal(mj, &a)
	//	iso, _ := info.Context().Isolate()
	resp, err := s.caller.CallGQL(context.Background(), `"`+args[0].String()+`"`, `"`+args[1].String()+`"`, a, `"`+args[2].String()+`"`)
	fmt.Println(string(resp))
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
