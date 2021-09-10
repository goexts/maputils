package gomap

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

//CustomHeader xml header
const CustomHeader = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>`

//String String
type String string

// ErrNilMap ...
var ErrNilMap = errors.New("nil map")

//String String
func (s String) String() string {
	return string(s)
}

// ToString ...
// @Description:
// @param string
// @return fmt.Stringer
func ToString(s string) fmt.Stringer {
	return String(s)
}

// Map ...
type Map map[string]interface{}

//String transfer map to JSON string
func (m Map) String() string {
	toJSON, err := m.ToJSON()
	if err != nil {
		return ""
	}
	return string(toJSON)
}

// StructToMap ...
func StructToMap(s interface{}) Map {
	return ToMap(structs.Map(s))
}

// New ...
func New() Map {
	return make(Map)
}

//ToMap transfer to map[string]interface{} or MapAble to GMap
func ToMap(p interface{}) Map {
	switch v := p.(type) {
	case map[string]interface{}:
		return v
	case Mapper:
		return v.ToMap()
	}
	return nil
}

// ToStruct transfer Map to struct
func (m Map) ToStruct(v interface{}) (e error) {
	return mapstructure.Decode(m, v)
}

// Merge marge all maps to target Map, the newer value will replace the older value
func Merge(maps ...Map) Map {
	target := New()
	if maps == nil {
		return target
	}
	for _, v := range maps {
		target.join(v, true)
	}
	return target
}

// Set ...
// @Description: set interface
// @receiver Map
// @param string
// @param interface{}
// @return Map
func (m Map) Set(key string, v interface{}) Map {
	return m.SetPath(strings.Split(key, "."), v)
}

// AppendArray ...
// @Description: append array interface to key
// @receiver Map
// @param string
// @param ...interface{}
// @return Map
func (m Map) AppendArray(key string, v ...interface{}) Map {
	source := m.GetArray(key)
	return m.Set(key, append(source, v...))
}

// SetPath is the same as SetPath, but allows you to provide comment
// information to the key, that will be reused by Marshal().
func (m Map) SetPath(keys []string, v interface{}) Map {
	subtree := m
	for _, intermediateKey := range keys[:len(keys)-1] {
		nextTree, exists := subtree[intermediateKey]
		if !exists {
			nextTree = make(Map)
			subtree[intermediateKey] = nextTree // add new element here
		}
		switch node := nextTree.(type) {
		case Map:
			subtree = node
		case []Map:
			// go to most recent element
			if len(node) == 0 {
				// create element if it does not exist
				subtree[intermediateKey] = append(node, make(Map))
			}
			subtree = node[len(node)-1]
		}
	}
	subtree[keys[len(keys)-1]] = v
	return m
}

//SetNil set value, if the key is not exist
func (m Map) SetNil(s string, v interface{}) Map {
	if !m.Has(s) {
		m.Set(s, v)
	}
	return m
}

//Replace replace will set value, if the key is exist
func (m Map) Replace(s string, v interface{}) Map {
	if m.Has(s) {
		m.Set(s, v)
	}
	return m
}

//ReplaceFromMap replace will set value from other map, if the key is exist from the both map
func (m Map) ReplaceFromMap(s string, v Map) Map {
	if m.Has(s) {
		m.Set(s, v[s])
	}
	return m
}

// Get ...
// @Description: get interface from map without default
// @receiver Map
// @param string
// @return interface{}
func (m Map) Get(s string) interface{} {
	if s == "" {
		return nil
	}
	if v := m.GetPath(strings.Split(s, ".")); v != nil {
		return v
	}
	return nil
}

// GetD ...
// @Description: get interface from map with default
// @receiver Map
// @param string
// @param interface{}
// @return interface{}
func (m Map) GetD(s string, d interface{}) interface{} {
	if s == "" {
		return nil
	}
	if v := m.GetPath(strings.Split(s, ".")); v != nil {
		return v
	}
	return d
}

// GetMap ...
// @Description: get map from map without default
// @receiver Map
// @param string
// @return Map
func (m Map) GetMap(s string) Map {
	return m.GetMapD(s, nil)
}

// GetMapD ...
// @Description: get map from map with default
// @receiver Map
// @param string
// @param Map
// @return Map
func (m Map) GetMapD(s string, d Map) Map {
	switch v := m.Get(s).(type) {
	case map[string]interface{}:
		return v
	case Map:
		return v
	default:
	}
	return d
}

//GetMapArray get map from map with out default
func (m Map) GetMapArray(s string) []Map {
	return m.GetMapArrayD(s, nil)

}

// GetMapArrayD ...
// @Description: get gomap from gomap with default
// @receiver Map
// @param string
// @param []Map
// @return []Map
func (m Map) GetMapArrayD(s string, d []Map) []Map {
	switch v := m.Get(s).(type) {
	case []Map:
		return v
	case []map[string]interface{}:
		var sub []Map
		for _, mp := range v {
			sub = append(sub, mp)
		}
		return sub
	default:
	}
	return d
}

// GetArray ...
// @Description:  get []interface value without default
// @receiver Map
// @param string
// @return []interface{}
func (m Map) GetArray(s string) []interface{} {
	return m.GetArrayD(s, nil)

}

// GetArrayD ...
// @Description: get []interface value with default
// @receiver Map
// @param string
// @param []interface{}
// @return []interface{}
func (m Map) GetArrayD(s string, d []interface{}) []interface{} {
	switch v := m.Get(s).(type) {
	case []interface{}:
		return v
	default:
	}
	return d
}

// GetBool ...
// @Description:  get bool from map with out default
// @receiver Map
// @param string
// @return bool
func (m Map) GetBool(s string) bool {
	return m.GetBoolD(s, false)
}

// GetBoolD ...
// @Description: get bool from map with default
// @receiver Map
// @param string
// @param bool
// @return bool
func (m Map) GetBoolD(s string, b bool) bool {
	if v, b := m.Get(s).(bool); b {
		return v
	}
	return b
}

// GetNumber ...
// @Description: get float64 from map with out default
// @receiver Map
// @param string
// @return float64
// @return bool
func (m Map) GetNumber(s string) (float64, bool) {
	return ParseNumber(m.Get(s))
}

// GetNumberD ...
// @Description: get float64 from map with default
// @receiver Map
// @param string
// @param float64
// @return float64
func (m Map) GetNumberD(s string, d float64) float64 {
	n, b := ParseNumber(m.Get(s))
	if b {
		return n
	}
	return d
}

// GetInt64 ...
// @Description: get int64 from map with out default
// @receiver Map
// @param string
// @return int64
// @return bool
func (m Map) GetInt64(s string) (int64, bool) {
	return ParseInt(m.Get(s))
}

// GetInt64D ...
// @Description: get int64 from map with default
// @receiver Map
// @param string
// @param int64
// @return int64
func (m Map) GetInt64D(s string, d int64) int64 {
	i, b := ParseInt(m.Get(s))
	if b {
		return i
	}
	return d
}

// GetString ...
// @Description: get string from map with out default
// @receiver Map
// @param string
// @return string
func (m Map) GetString(s string) string {
	return m.GetStringD(s, "")
}

// GetStringD ...
// @Description: get string from map with default
// @receiver Map
// @param string
// @param string
// @return string
func (m Map) GetStringD(s string, d string) string {
	if v, b := m.Get(s).(string); b {
		return v
	}
	return d
}

// GetStringArray ...
// @Description: get string from map without default
// @receiver Map
// @param string
// @return []string
func (m Map) GetStringArray(s string) []string {
	return m.GetStringArrayD(s, []string{})
}

// GetStringArrayD ...
// @Description: get string from map with default
// @receiver Map
// @param string
// @param []string
// @return []string
func (m Map) GetStringArrayD(s string, d []string) []string {
	if v, b := m.Get(s).([]string); b {
		return v
	}
	return d
}

// GetBytes ...
// @Description: get bytes from map with default
// @receiver Map
// @param string
// @return []byte
func (m Map) GetBytes(s string) []byte {
	return m.GetBytesD(s, nil)

}

// GetBytesD ...
// @Description: get bytes from map with default
// @receiver Map
// @param string
// @param []byte
// @return []byte
func (m Map) GetBytesD(s string, d []byte) []byte {
	if v, b := m.Get(s).([]byte); b {
		return v
	}
	return d
}

// Delete ...
// @Description:  delete key value if key is exist
// @receiver Map
// @param string
// @return bool
func (m Map) Delete(key string) bool {
	if key == "" {
		return false
	}
	return m.DeletePath(strings.Split(key, "."))
}

// DeletePath delete keys value if keys is exist
func (m Map) DeletePath(keys []string) bool {
	if len(keys) == 0 {
		return false
	}
	subtree := m
	for _, intermediateKey := range keys[:len(keys)-1] {
		value, exists := subtree[intermediateKey]
		if !exists {
			return false
		}
		switch node := value.(type) {
		case Map:
			subtree = node
		case []Map:
			if len(node) == 0 {
				return false
			}
			subtree = node[len(node)-1]
		default:
			return false // cannot navigate through other node types
		}
	}
	// branch based on final node type
	if _, b := subtree[keys[len(keys)-1]]; !b {
		return false
	}
	delete(subtree, keys[len(keys)-1])
	return true
}

//Has check if key exist
func (m Map) Has(key string) bool {
	if key == "" {
		return false
	}
	return m.HasPath(strings.Split(key, "."))

}

// HasPath returns true if the given path of keys exists, false otherwise.
func (m Map) HasPath(keys []string) bool {
	return m.GetPath(keys) != nil
}

// GetPath returns the element in the tree indicated by 'keys'.
// If keys is of length zero, the current tree is returned.
func (m Map) GetPath(keys []string) interface{} {
	if len(keys) == 0 {
		return m
	}
	subtree := m
	for _, intermediateKey := range keys[:len(keys)-1] {
		value, exists := subtree[intermediateKey]
		if !exists {
			return nil
		}
		switch node := value.(type) {
		case Map:
			subtree = node
		case []Map:
			if len(node) == 0 {
				return nil
			}
			subtree = node[len(node)-1]
		default:
			return nil // cannot navigate through other node types
		}
	}
	// branch based on final node type
	v, b := subtree[keys[len(keys)-1]]
	if b {
		return v
	}
	return nil
}

//SortKeys 排列key
func (m Map) SortKeys() []string {
	var keys sort.StringSlice
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	return keys
}

//ToXML transfer map to XML
func (m Map) ToXML() ([]byte, error) {
	return mapToXML(m, true)
}

//ParseXML parse XML bytes to map
func (m Map) ParseXML(b []byte) error {
	return xmlToMap(m, b, true)
}

//ToJSON transfer map to JSON
func (m Map) ToJSON() (v []byte, err error) {
	v, err = json.Marshal(m)
	return
}

//ParseJSON parse JSON bytes to map
func (m Map) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &m)
}

// Append append source map to target map;
// if the key value is exist and it is a []interface value, this will append into it
// otherwise, it will replace the value
func (m Map) Append(p Map) Map {
	for k, v := range p {
		if vget := m.Get(k); vget != nil {
			if vvget, b := vget.([]interface{}); b {
				vvget = append(vvget, v)
				m.Set(k, vvget)
				continue
			}
		}
		m.Set(k, v)
	}
	return m
}

func (m Map) join(source Map, replace bool) Map {
	for k, v := range source {
		if replace || !m.Has(k) {
			m.Set(k, v)
		}
	}
	return m
}

//ReplaceJoin insert map s to m with replace
func (m Map) ReplaceJoin(s Map) Map {
	return m.join(s, true)
}

//Join insert map s to m with out replace
func (m Map) Join(s Map) Map {
	return m.join(s, false)
}

//Only get map with keys
func (m Map) Only(keys []string) Map {
	p := Map{}
	size := len(keys)
	for i := 0; i < size; i++ {
		p.Set(keys[i], m.Get(keys[i]))
	}

	return p
}

//Expect get map expect keys
func (m Map) Expect(keys []string) Map {
	p := m.Clone()
	size := len(keys)
	for i := 0; i < size; i++ {
		p.Delete(keys[i])
	}
	return p
}

//Clone copy a map
func (m Map) Clone() Map {
	v := deepCopy(m)
	return (v).(Map)
}

func deepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(Map); ok {
		newMap := make(Map)
		for k, v := range valueMap {
			newMap[k] = deepCopy(v)
		}
		return newMap
	} else if valueSlice, ok := value.([]Map); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = deepCopy(v)
		}
		return newSlice
	}

	return value
}

//Range range all maps
func (m Map) Range(f func(key string, value interface{}) bool) {
	for k, v := range m {
		if !f(k, v) {
			return
		}
	}
}

//Check check all input keys
//return -1 if all is exist
//return index when not found
func (m Map) Check(s ...string) int {
	size := len(s)
	for i := 0; i < size; i++ {
		if !m.Has(s[i]) {
			return i
		}
	}
	return -1
}

// ToGoMap trans return a map[string]interface from Map
func (m Map) ToGoMap() map[string]interface{} {
	return m
}

// ToMap implements MapAble by self
func (m Map) ToMap() Map {
	return m
}

//ToEncodeURL transfer map to url encode
func (m Map) ToEncodeURL() string {
	var buf strings.Builder
	keys := m.SortKeys()
	size := len(keys)
	for i := 0; i < size; i++ {
		vs := m[keys[i]]
		keyEscaped := url.QueryEscape(keys[i])
		switch val := vs.(type) {
		case string:
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(val))
		case []string:
			for _, v := range val {
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

// MarshalXML ...
func (m Map) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m) == 0 {
		return ErrNilMap
	}
	if start.Name.Local == "root" {
		return marshalXML(m, e, xml.StartElement{Name: xml.Name{Local: "root"}})
	}
	return marshalXML(m, e, xml.StartElement{Name: xml.Name{Local: "xml"}})
}

// UnmarshalXML ...
func (m Map) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local == "root" {
		return unmarshalXML(m, d, xml.StartElement{Name: xml.Name{Local: "root"}}, false)
	}
	return unmarshalXML(m, d, xml.StartElement{Name: xml.Name{Local: "xml"}}, false)
}
