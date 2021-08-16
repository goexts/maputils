package extmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

//String String
type String string

// ErrNilMap ...
var ErrNilMap = errors.New("nil map")
var ErrUnsupportedType = errors.New("error unsupported type")

func init() {
	structs.DefaultTagName = "map"
}

//String String
func (s String) String() string {
	return string(s)
}

//ToString ToString
func ToString(s string) String {
	return String(s)
}

type Map interface {
	Option() *Option
	Bind(v interface{}) (e error)
	Set(k interface{}, v interface{}) Map
	Get(s interface{}) interface{}
	Range(f func(key interface{}, value interface{}) bool)
}

// Map ...
type innerMap struct {
	data   map[interface{}]interface{}
	option *Option
}

func (m innerMap) Option() *Option {
	op := *m.option
	return &op
}

func newInnerMap(op *Option) *innerMap {
	return &innerMap{
		data:   make(map[interface{}]interface{}),
		option: op,
	}
}

//String transfer map to JSON string
func (m innerMap) String() string {
	toJSON, err := m.ToJSON()
	if err != nil {
		return ""
	}
	return string(toJSON)
}

// StructToMap ...
func StructToMap(s interface{}) Map {
	return ToExtMap(structs.Map(s))
}

// New create a map interface
func New(fns ...OptionFunc) Map {
	op := defaultOption()
	for _, fn := range fns {
		fn(op)
	}
	return newInnerMap(op)
}

//ToExtMap transfer to map[string]interface{} or MapAble to GMap
func ToExtMap(p interface{}) Map {
	switch v := p.(type) {
	case map[interface{}]interface{}:
		return &innerMap{data: v}
	//todo: add other type process
	default:
		panic(ErrUnsupportedType)
	}
	return nil
}

func (m innerMap) Bind(v interface{}) (e error) {
	switch n := v.(type) {
	case struct{}:
		return m.BindStruct(v)
	default:
		return &mapErr{v: fmt.Sprintf("%v", n)}
	}
}

func (m innerMap) BindStruct(v interface{}) (e error) {
	return nil
}

// ToStruct transfer Map to struct
func (m innerMap) ToStruct(v interface{}) (e error) {
	return mapstructure.Decode(m, v)
}

// Merge marge all maps to target Map, the newer value will replace the older value
func Merge(maps ...Map) Map {
	target := newInnerMap(defaultOption())
	if maps == nil {
		return target
	}
	for _, v := range maps {
		target.join(v, true)
	}
	return target
}

//Set set interface
func (m *innerMap) Set(key interface{}, v interface{}) Map {
	if m.option.Split {
		switch k := key.(type) {
		case string:
			return m.setString(k, v)
		}
	}
	return m.set(key, v)
}

func (m *innerMap) set(key interface{}, val interface{}) Map {
	m.data[key] = val
	return m
}
func (m *innerMap) setString(key interface{}, val interface{}) Map {
	return m.SetPath(strings.Split(key.(string), "."), val)
}

// SetPath is the same as SetPath, but allows you to provide comment
// information to the key, that will be reused by Marshal().
func (m *innerMap) SetPath(keys []string, v interface{}) Map {
	subtree := m
	for _, intermediateKey := range keys[:len(keys)-1] {
		nextTree, exists := m.data[intermediateKey]
		if !exists {
			nextTree = New()
			m.data[intermediateKey] = nextTree // add newStruct element here
		}
		switch node := nextTree.(type) {
		case *innerMap:
			subtree = node
		case []*innerMap:
			// go to most recent element
			if len(node) == 0 {
				// create element if it does not exist
				subtree.data[intermediateKey] = append(node, newInnerMap(nil))
			}
			subtree = node[len(node)-1]
		}
	}
	subtree.data[keys[len(keys)-1]] = v
	return m
}

//SetNil set value, if the key is not exist
func (m *innerMap) SetNil(s string, v interface{}) Map {
	if !m.Has(s) {
		m.Set(s, v)
	}
	return m
}

//Replace replace will set value, if the key is exist
func (m *innerMap) Replace(s string, v interface{}) Map {
	if m.Has(s) {
		m.Set(s, v)
	}
	return m
}

//ReplaceFromMap replace will set value from other map, if the key is exist from the both map
func (m *innerMap) ReplaceFromMap(s string, v Map) Map {
	if m.Has(s) {
		m.Set(s, v.Get(s))
	}
	return m
}

//Get get interface from map with out default
func (m innerMap) Get(key interface{}) interface{} {
	switch k := key.(type) {
	case string:
		return m.getString(k)
	default:
		return m.get(key)
	}
}

func (m innerMap) get(k interface{}) interface{} {
	return m.data[k]
}

func (m innerMap) getString(k string) interface{} {
	if v := m.GetPath(strings.Split(k, ".")); v != nil {
		return v
	}
	return nil
}

//GetD get interface from map with default
func (m *innerMap) GetD(s string, d interface{}) interface{} {
	if s == "" {
		return nil
	}
	if v := m.GetPath(strings.Split(s, ".")); v != nil {
		return v
	}
	return d
}

//GetMap get map from map with out default
func (m *innerMap) GetMap(s string) Map {
	return m.GetMapD(s, nil)
}

//GetMapD get map from map with default
func (m *innerMap) GetMapD(s string, d Map) Map {
	switch v := m.Get(s).(type) {
	case map[interface{}]interface{}:
		m := newInnerMap(nil)
		m.data = v
		return m
	case Map:
		return v
	default:
	}
	return d
}

//GetMapArray get map from map with out default
func (m *innerMap) GetMapArray(s string) []Map {
	return m.GetMapArrayD(s, nil)

}

//GetMapArrayD get map from map with default
func (m *innerMap) GetMapArrayD(s string, d []Map) []Map {
	switch v := m.Get(s).(type) {
	case []Map:
		return v
	case []map[interface{}]interface{}:
		var sub []Map
		for _, mp := range v {
			m := newInnerMap(nil)
			m.data = mp
			sub = append(sub, m)
		}
		return sub
	default:
	}
	return d
}

// GetArray get []interface value with out default
func (m *innerMap) GetArray(s string) []interface{} {
	return m.GetArrayD(s, nil)

}

// GetArrayD get []interface value with default
func (m *innerMap) GetArrayD(s string, d []interface{}) []interface{} {
	switch v := m.Get(s).(type) {
	case []interface{}:
		return v
	default:
	}
	return d
}

//GetBool get bool from map with out default
func (m *innerMap) GetBool(s string) bool {
	return m.GetBoolD(s, false)
}

//GetBoolD get bool from map with default
func (m *innerMap) GetBoolD(s string, b bool) bool {
	if v, b := m.Get(s).(bool); b {
		return v
	}
	return b
}

//GetNumber get float64 from map with out default
func (m *innerMap) GetNumber(s string) (float64, bool) {
	return ParseNumber(m.Get(s))
}

//GetNumberD get float64 from map with default
func (m *innerMap) GetNumberD(s string, d float64) float64 {
	n, b := ParseNumber(m.Get(s))
	if b {
		return n
	}
	return d
}

//GetInt64 get int64 from map with out default
func (m *innerMap) GetInt64(s string) (int64, bool) {
	return ParseInt(m.Get(s))
}

//GetInt64D get int64 from map with default
func (m *innerMap) GetInt64D(s string, d int64) int64 {
	i, b := ParseInt(m.Get(s))
	if b {
		return i
	}
	return d
}

//GetString get string from map with out default
func (m *innerMap) GetString(s string) string {
	return m.GetStringD(s, "")
}

//GetStringD get string from map with default
func (m *innerMap) GetStringD(s string, d string) string {
	if v, b := m.Get(s).(string); b {
		return v
	}
	return d
}

//GetStringArray get string from map with out default
func (m *innerMap) GetStringArray(s string) []string {
	return m.GetStringArrayD(s, []string{})
}

//GetStringD get string from map with default
func (m *innerMap) GetStringArrayD(s string, d []string) []string {
	if v, b := m.Get(s).([]string); b {
		return v
	}
	return d
}

//GetBytes get bytes from map with default
func (m *innerMap) GetBytes(s string) []byte {
	return m.GetBytesD(s, nil)

}

//GetBytesD get bytes from map with default
func (m *innerMap) GetBytesD(s string, d []byte) []byte {
	if v, b := m.Get(s).([]byte); b {
		return v
	}
	return d
}

//Delete delete key value if key is exist
func (m *innerMap) Delete(key string) bool {
	if key == "" {
		return false
	}
	return m.DeletePath(strings.Split(key, "."))
}

// DeletePath delete keys value if keys is exist
func (m *innerMap) DeletePath(keys []string) bool {
	panic("todo")
	//if len(keys) == 0 {
	//	return false
	//}
	//subtree := m
	//for _, intermediateKey := range keys[:len(keys)-1] {
	//	value, exists := subtree[intermediateKey]
	//	if !exists {
	//		return false
	//	}
	//	switch node := value.(type) {
	//	case Map:
	//		subtree = node
	//	case []Map:
	//		if len(node) == 0 {
	//			return false
	//		}
	//		subtree = node[len(node)-1]
	//	default:
	//		return false // cannot navigate through other node types
	//	}
	//}
	//// branch based on final node type
	//if _, b := subtree[keys[len(keys)-1]]; !b {
	//	return false
	//}
	//delete(subtree, keys[len(keys)-1])
	return true
}

//Has check if key exist
func (m *innerMap) Has(key interface{}) bool {
	switch k := key.(type) {
	case string:
		return m.hasString(k)
	default:
		return m.has(key)
	}
}
func (m innerMap) has(key interface{}) bool {
	_, b := m.data[key]
	return b
}

func (m innerMap) hasString(key string) bool {
	if key == "" {
		return false
	}
	return m.HasPath(strings.Split(key, "."))
}

// HasPath returns true if the given path of keys exists, false otherwise.
func (m *innerMap) HasPath(keys []string) bool {
	return m.GetPath(keys) != nil
}

// GetPath returns the element in the tree indicated by 'keys'.
// If keys is of length zero, the current tree is returned.
func (m *innerMap) GetPath(keys []string) interface{} {
	panic("todo")
	//if len(keys) == 0 {
	//	return nil
	//}
	//subtree := m
	//for _, intermediateKey := range keys[:len(keys)-1] {
	//	value, exists := subtree[intermediateKey]
	//	if !exists {
	//		return nil
	//	}
	//	switch node := value.(type) {
	//	case Map:
	//		subtree = node
	//	case []Map:
	//		if len(node) == 0 {
	//			return nil
	//		}
	//		subtree = node[len(node)-1]
	//	default:
	//		return nil // cannot navigate through other node types
	//	}
	//}
	//// branch based on final node type
	//v, b := subtree[keys[len(keys)-1]]
	//if b {
	//	return v
	//}
	//return nil
}

//SortKeys 排列key
func (m *innerMap) SortKeys() []string {
	var keys sort.StringSlice
	for k := range m.data {
		keys = append(keys, k.(string))
	}
	sort.Sort(keys)
	return keys
}

//ToJSON transfer map to JSON
func (m *innerMap) ToJSON() (v []byte, err error) {
	v, err = json.Marshal(m)
	return
}

//ParseJSON parse JSON bytes to map
func (m *innerMap) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &m)
}

// Append append source map to target map;
// if the key value is exist and it is a []interface value, this will append into it
// otherwise, it will replace the value
func (m *innerMap) Append(p Map) Map {
	panic("todo")
	//for k, v := range p {
	//	if vget := m.Get(k); vget != nil {
	//		if vvget, b := vget.([]interface{}); b {
	//			vvget = append(vvget, v)
	//			m.Set(k, vvget)
	//			continue
	//		}
	//	}
	//	m.Set(k, v)
	//}
	//return m
}

func (m *innerMap) join(source Map, replace bool) Map {
	s := source.(*innerMap)
	for k, v := range s.data {
		if replace || !m.Has(k) {
			m.Set(k, v)
		}
	}
	return m
}

//ReplaceJoin insert map s to m with replace
func (m *innerMap) ReplaceJoin(s Map) Map {
	return m.join(s, true)
}

//Join insert map s to m with out replace
func (m *innerMap) Join(s Map) Map {
	return m.join(s, false)
}

//Only get map with keys
func (m *innerMap) Only(keys []interface{}) Map {
	p := New()
	size := len(keys)
	for i := 0; i < size; i++ {
		p.Set(keys[i], m.Get(keys[i]))
	}

	return p
}

//Expect get map expect keys
func (m *innerMap) Expect(keys []string) Map {
	panic("todo")
	//p := m.Clone()
	//size := len(keys)
	//for i := 0; i < size; i++ {
	//	p.Delete(keys[i])
	//}
	//return p
}

//Clone copy a map
func (m *innerMap) Clone() Map {
	v := deepCopy(m)
	return (v).(Map)
}

func deepCopy(value interface{}) interface{} {
	panic("todo")
	//if valueMap, ok := value.(Map); ok {
	//	newMap := make(Map)
	//	for k, v := range valueMap {
	//		newMap[k] = deepCopy(v)
	//	}
	//	return newMap
	//} else if valueSlice, ok := value.([]Map); ok {
	//	newSlice := make([]interface{}, len(valueSlice))
	//	for k, v := range valueSlice {
	//		newSlice[k] = deepCopy(v)
	//	}
	//	return newSlice
	//}
	//
	//return value
}

//Range range all maps
func (m *innerMap) Range(f func(key interface{}, value interface{}) bool) {
	for k, v := range m.data {
		if !f(k, v) {
			return
		}
	}
}

//Check check all input keys
//return -1 if all is exist
//return index when not found
func (m *innerMap) Check(s ...interface{}) int {
	size := len(s)
	for i := 0; i < size; i++ {
		if !m.Has(s[i]) {
			return i
		}
	}
	return -1
}

// ToGoMap trans return a map[string]interface from Map
func (m *innerMap) ToGoMap() map[interface{}]interface{} {
	return m.data
}

// ToMap implements MapAble by self
func (m *innerMap) ToMap() Map {
	return m
}

//ToEncodeURL transfer map to url encode
func (m *innerMap) ToEncodeURL() string {
	var buf strings.Builder
	keys := m.SortKeys()
	size := len(keys)
	var tmp interface{}
	for i := 0; i < size; i++ {
		tmp = m.get(keys[i])
		keyEscaped := url.QueryEscape(keys[i])
		switch v := tmp.(type) {
		case string:
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
		case []string:
			for _, v := range v {
				if buf.Len() > 0 {
					buf.WriteByte('&')
				}
				buf.WriteString(keyEscaped)
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(v))
			}
		}
	}

	return buf.String()
}

var _ Map = (*innerMap)(nil)
