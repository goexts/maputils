package extmap

import (
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strings"

	"github.com/fatih/structs"
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

type Map struct {
	setting *Setting
	m       map[string]any
}

func (m Map) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &m.m)
}

func (m Map) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.m)
}

func (m Map) String() string {
	v, _ := m.MarshalJSON()
	return string(v)
}

// New create a map interface
func New(ss ...SettingOption) *Map {
	setting := defaultSetting()
	for i := range ss {
		ss[i](setting)
	}
	return &Map{
		setting: setting,
		m:       make(map[string]any),
	}
}

func newWithSetting(setting *Setting) *Map {
	return &Map{
		setting: setting,
		m:       make(map[string]any),
	}
}

//ToMap transfer to map[comparable]any to Map
func ToMap(p any) *Map {
	switch m := p.(type) {
	case map[string]any:
		return &Map{m: m}
	//todo: add other type process
	default:
		panic(ErrUnsupportedType)
	}
	return nil
}

// Merge marge all maps to target Map, the newer value will replace the older value
func (m *Map) Merge(maps ...*Map) *Map {
	for _, v := range maps {
		m.join(v, true)
	}
	return m
}

//Set is a set interface
func (m *Map) Set(key string, v any) *Map {
	if m.setting.Split {
		//switch k := key.(type) {
		//case string:
		return m.setString(key, v)
		//}
	}
	return m.set(key, v)
}

func (m *Map) set(key string, val any) *Map {
	m.m[key] = val
	return m
}
func (m *Map) setString(key string, val any) *Map {
	return m.SetPath(strings.Split(key, "."), val)
}

func (m *Map) Query(key string) (any, error) {
	return nil, nil
}

// SetPath is the same as SetPath, but allows you to provide comment
// information to the key, that will be reused by Marshal().
func (m *Map) SetPath(keys []string, v any) *Map {
	subtree := m
	for _, intermediateKey := range keys[:len(keys)-1] {
		nextTree, exists := m.m[intermediateKey]
		if !exists {
			nextTree = New()
			m.m[intermediateKey] = nextTree // add newStruct element here
		}
		switch node := nextTree.(type) {
		case *Map:
			subtree = node
		case []*Map:
			// go to most recent element
			if len(node) == 0 {
				// create element if it does not exist
				subtree.m[intermediateKey] = append(node, New(func(op *Setting) {
					op = m.setting
				}))
			}
			subtree = node[len(node)-1]
		}
	}
	subtree.m[keys[len(keys)-1]] = v
	return m
}

//SetNil set value, if the key is not exist
func (m *Map) SetNil(key string, val any) *Map {
	if !m.Has(key) {
		m.Set(key, val)
	}
	return m
}

//Replace replace will set value, if the key is exist
func (m *Map) Replace(s string, v any) *Map {
	if m.Has(s) {
		m.Set(s, v)
	}
	return m
}

//ReplaceFromMap replace will set value from other map, if the key is exist from the both map
func (m *Map) ReplaceFromMap(key string, v *Map) *Map {
	if m.Has(key) {
		m.Set(key, v.Get(key))
	}
	return m
}

//Get get interface from map with out default
func (m Map) Get(key string) any {
	return m.getString(key)

}

func (m Map) get(k string) any {
	return m.m[k]
}

func (m Map) getString(k string) any {
	if v := m.GetPath(strings.Split(k, ".")); v != nil {
		return v
	}
	return nil
}

//GetD get interface from map with default
func (m *Map) GetD(s string, d any) any {
	if s == "" {
		return nil
	}
	if v := m.GetPath(strings.Split(s, ".")); v != nil {
		return v
	}
	return d
}

//GetMap get map from map with out default
func (m *Map) GetMap(s string) *Map {
	return m.GetMapD(s, nil)
}

//GetMapD get map from map with default
func (m *Map) GetMapD(s string, d *Map) *Map {
	switch v := m.Get(s).(type) {
	case map[string]any:
		m := newWithSetting(m.setting)
		m.m = v
		return m
	case *Map:
		return v
	default:
	}
	return d
}

//GetMapArray get map from map with out default
func (m *Map) GetMapArray(s string) []*Map {
	return m.GetMapArrayD(s, nil)

}

//GetMapArrayD get map from map with default
func (m *Map) GetMapArrayD(s string, d []*Map) []*Map {
	switch v := m.Get(s).(type) {
	case []*Map:
		return v
	case []map[string]any:
		var sub []*Map
		for _, mp := range v {
			_map := newWithSetting(m.setting)
			_map.m = mp
			sub = append(sub, m)
		}
		return sub
	default:
	}
	return d
}

// GetArray get []interface value with out default
func (m *Map) GetArray(s string) []any {
	return m.GetArrayD(s, nil)

}

// GetArrayD get []interface value with default
func (m *Map) GetArrayD(s string, d []any) []any {
	switch v := m.Get(s).(type) {
	case []any:
		return v
	default:
	}
	return d
}

//GetBool get bool from map with out default
func (m *Map) GetBool(s string) bool {
	return m.GetBoolD(s, false)
}

//GetBoolD get bool from map with default
func (m *Map) GetBoolD(s string, b bool) bool {
	if v, b := m.Get(s).(bool); b {
		return v
	}
	return b
}

//GetNumber get float64 from map with out default
func (m *Map) GetNumber(s string) (float64, bool) {
	return ParseNumber(m.Get(s))
}

//GetNumberD get float64 from map with default
func (m *Map) GetNumberD(s string, d float64) float64 {
	n, b := ParseNumber(m.Get(s))
	if b {
		return n
	}
	return d
}

//GetInt64 get int64 from map with out default
func (m *Map) GetInt64(s string) (int64, bool) {
	return ParseInt(m.Get(s))
}

//GetInt64D get int64 from map with default
func (m *Map) GetInt64D(s string, d int64) int64 {
	i, b := ParseInt(m.Get(s))
	if b {
		return i
	}
	return d
}

//GetString get string from map with out default
func (m *Map) GetString(s string) string {
	return m.GetStringD(s, "")
}

//GetStringD get string from map with default
func (m *Map) GetStringD(s string, d string) string {
	if v, b := m.Get(s).(string); b {
		return v
	}
	return d
}

//GetStringArray get string from map with out default
func (m *Map) GetStringArray(s string) []string {
	return m.GetStringArrayD(s, []string{})
}

//GetStringD get string from map with default
func (m *Map) GetStringArrayD(s string, d []string) []string {
	if v, b := m.Get(s).([]string); b {
		return v
	}
	return d
}

//GetBytes get bytes from map with default
func (m *Map) GetBytes(s string) []byte {
	return m.GetBytesD(s, nil)

}

//GetBytesD get bytes from map with default
func (m *Map) GetBytesD(s string, d []byte) []byte {
	if v, b := m.Get(s).([]byte); b {
		return v
	}
	return d
}

//Delete delete key value if key is exist
func (m *Map) Delete(key string) bool {
	if key == "" {
		return false
	}
	return m.DeletePath(strings.Split(key, "."))
}

// DeletePath delete keys value if keys is exist
func (m *Map) DeletePath(keys []string) bool {
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
func (m *Map) Has(key string) bool {
	//switch k := key.(type) {
	//case string:
	return m.hasString(key)
	//default:
	//	return m.has(key)
	//}
}
func (m Map) has(key string) bool {
	_, b := m.m[key]
	return b
}

func (m Map) hasString(key string) bool {
	if key == "" {
		return false
	}
	return m.HasPath(strings.Split(key, "."))
}

// HasPath returns true if the given path of keys exists, false otherwise.
func (m *Map) HasPath(keys []string) bool {
	return m.GetPath(keys) != nil
}

// GetPath returns the element in the tree indicated by 'keys'.
// If keys is of length zero, the current tree is returned.
func (m *Map) GetPath(keys []string) any {
	if len(keys) == 0 {
		return nil
	}
	subtree := m
	for _, intermediateKey := range keys[:len(keys)-1] {
		value, exists := subtree.m[intermediateKey]
		if !exists {
			return nil
		}
		switch node := value.(type) {
		case *Map:
			subtree = node
		case []*Map:
			if len(node) == 0 {
				return nil
			}
			subtree = node[len(node)-1]
		default:
			return nil // cannot navigate through other node types
		}
	}
	// branch based on final node type
	v, b := subtree.m[keys[len(keys)-1]]
	if b {
		return v
	}
	return nil
}

//SortKeys 排列key
func (m *Map) SortKeys() []string {
	var keys sort.StringSlice
	for k := range m.m {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	return keys
}

//ToJSON transfer map to JSON
func (m *Map) ToJSON() (v []byte, err error) {
	v, err = json.Marshal(m)
	return
}

//ParseJSON parse JSON bytes to map
func (m *Map) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &m)
}

// Append append source map to target map;
// if the key value is exist and it is a []interface value, this will append into it
// otherwise, it will replace the value
func (m *Map) Append(source *Map) *Map {
	for k, v := range source.m {
		if vget := m.Get(k); vget != nil {
			if vvget, b := vget.([]any); b {
				vvget = append(vvget, v)
				m.Set(k, vvget)
				continue
			}
		}
		m.Set(k, v)
	}
	return m
}

func (m *Map) join(source *Map, replace bool) *Map {
	for k, v := range source.m {
		if replace || !m.Has(k) {
			m.Set(k, v)
		}
	}
	return m
}

//ReplaceJoin insert map s to m with replace
func (m *Map) ReplaceJoin(source *Map) *Map {
	return m.join(source, true)
}

//Join insert map s to m with out replace
func (m *Map) Join(source *Map) *Map {
	return m.join(source, false)
}

//Only get map with keys
func (m *Map) Only(keys []string) *Map {
	_map := New(func(op *Setting) {
		op = m.setting
	})
	size := len(keys)
	for i := 0; i < size; i++ {
		_map.Set(keys[i], m.Get(keys[i]))
	}

	return _map
}

//Expect get map expect keys
func (m *Map) Expect(keys []string) Map {
	p := m.Clone()
	size := len(keys)
	for i := 0; i < size; i++ {
		p.Delete(keys[i])
	}
	return p
}

//Clone copy a map
func (m *Map) Clone() Map {
	v := deepCopy(m)
	return (v).(Map)
}

func deepCopy(value any) any {
	if valueMap, ok := value.(*Map); ok {
		newMap := newWithSetting(valueMap.setting)
		for k, v := range valueMap.m {
			newMap.m[k] = deepCopy(v)
		}
		return newMap
	} else if valueSlice, ok := value.([]*Map); ok {
		newSlice := make([]any, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = deepCopy(v)
		}
		return newSlice
	}

	return value
}

//Range range all maps
func (m *Map) Range(f func(key string, value any) bool) {
	for k, v := range m.m {
		if !f(k, v) {
			return
		}
	}
}

//Check check all input keys
//return -1 if all is exist
//return index when not found
func (m *Map) Check(s ...string) int {
	size := len(s)
	for i := 0; i < size; i++ {
		if !m.Has(s[i]) {
			return i
		}
	}
	return -1
}

// GoMap trans return a map[string]interface from Map
func (m *Map) GoMap() map[string]any {
	return m.m
}

// Map implements MapAble by self
func (m *Map) Map() *Map {
	return m
}

//ToEncodeURL transfer map to url encode
func (m *Map) ToEncodeURL() string {
	var buf strings.Builder
	keys := m.SortKeys()
	size := len(keys)
	var tmp any
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
