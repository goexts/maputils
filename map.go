package extmap

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

/*CustomHeader xml header*/
const CustomHeader = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>`

// Mapper ...
type Mapper interface {
	Map() Map
}

// XMLer ...
type XMLer interface {
	XML() []byte
}

// JSONer ...
type JSONer interface {
	JSON() []byte
}

// Stringer ...
type Stringer interface {
	String() string
}

/*CDATA xml cdata defines */
type CDATA struct {
	XMLName xml.Name
	Value   string `xml:",cdata"`
}

/*String String */
type String string

// ErrNilMap ...
var ErrNilMap = errors.New("nil map")

func init() {
	structs.DefaultTagName = "map"
}

/*String String */
func (s String) String() string {
	return string(s)
}

/*ToString ToString */
func ToString(s string) String {
	return String(s)
}

// Map ...
type Map map[string]interface{}

/*String transfer map to JSON string */
func (m Map) String() string {
	return string(m.JSON())
}

// StructToMap ...
func StructToMap(s interface{}) Map {
	return ToMap(structs.Map(s))
}

// New ...
func New() Map {
	return make(Map)
}

/*ToMap transfer to map[string]interface{} or MapAble to GMap  */
func ToMap(p interface{}) Map {
	switch v := p.(type) {
	case map[string]interface{}:
		return v
	case Mapper:
		return v.Map()
	}
	return nil
}

// Struct ...
func (m Map) Struct(v interface{}) (e error) {
	return mapstructure.Decode(m, v)
}

// MergeMaps ...
func MergeMaps(target Map, sources ...Map) Map {
	if sources == nil {
		return target
	}

	for _, v := range sources {
		target.join(v, true)
	}
	return target
}

/*Set set interface */
func (m Map) Set(key string, v interface{}) Map {
	return m.SetPath(strings.Split(key, "."), v)
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

/*ReplaceNil replace interface if key is not exist */
func (m Map) ReplaceNil(s string, v interface{}) Map {
	if !m.Has(s) {
		m.Set(s, v)
	}
	return m
}

/*ReplaceExist replace interface if key is exist */
func (m Map) ReplaceExist(s string, v interface{}) Map {
	if m.Has(s) {
		m.Set(s, v)
	}
	return m
}

/*ReplaceMap replace value from source map if key is exist */
func (m Map) ReplaceMap(s string, v Map) Map {
	if m.Has(s) {
		m.Set(s, v[s])
	}
	return m
}

/*Get get interface from map with out default */
func (m Map) Get(key string) interface{} {
	return m.GetD(key, "")
}

/*GetD get interface from map with default */
func (m Map) GetD(s string, d interface{}) interface{} {
	if s == "" {
		return nil
	}
	if v := m.GetPath(strings.Split(s, ".")); v != nil {
		return v
	}
	return d
}

/*GetMap get map from map with out default */
func (m Map) GetMap(s string) Map {
	return m.GetMapD(s, nil)
}

/*GetMapD get map from map with default */
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

/*GetMapArray get map from map with out default */
func (m Map) GetMapArray(s string) []Map {
	return m.GetMapArrayD(s, nil)

}

/*GetMapArrayD get map from map with default */
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
func (m Map) GetArray(s string) []interface{} {
	return m.GetArrayD(s, nil)

}

// GetArrayD ...
func (m Map) GetArrayD(s string, d []interface{}) []interface{} {
	switch v := m.Get(s).(type) {
	case []interface{}:
		return v
	default:
	}
	return d
}

/*GetBool get bool from map with out default */
func (m Map) GetBool(s string) bool {
	return m.GetBoolD(s, false)
}

/*GetBoolD get bool from map with default */
func (m Map) GetBoolD(s string, b bool) bool {
	if v, b := m.Get(s).(bool); b {
		return v
	}
	return b
}

/*GetNumber get float64 from map with out default */
func (m Map) GetNumber(s string) (float64, bool) {
	return ParseNumber(m.Get(s))
}

/*GetNumberD get float64 from map with default */
func (m Map) GetNumberD(s string, d float64) float64 {
	n, b := ParseNumber(m.Get(s))
	if b {
		return n
	}
	return d
}

/*GetInt64 get int64 from map with out default */
func (m Map) GetInt64(s string) (int64, bool) {
	return ParseInt(m.Get(s))
}

/*GetInt64D get int64 from map with default */
func (m Map) GetInt64D(s string, d int64) int64 {
	i, b := ParseInt(m.Get(s))
	if b {
		return i
	}
	return d
}

/*GetString get string from map with out default */
func (m Map) GetString(s string) string {
	return m.GetStringD(s, "")

}

/*GetStringD get string from map with default */
func (m Map) GetStringD(s string, d string) string {
	if v, b := m.Get(s).(string); b {
		return v
	}
	return d
}

/*GetBytes get bytes from map with default */
func (m Map) GetBytes(s string) []byte {
	return m.GetBytesD(s, nil)

}

/*GetBytesD get bytes from map with default */
func (m Map) GetBytesD(s string, d []byte) []byte {
	if v, b := m.Get(s).([]byte); b {
		return v
	}
	return d
}

/*Delete delete if exist */
func (m Map) Delete(key string) bool {
	if key == "" {
		return false
	}
	return m.DeletePath(strings.Split(key, "."))
}

// DeletePath ...
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

/*Has check if key exist */
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
	return subtree[keys[len(keys)-1]]
}

/*SortKeys 排列key */
func (m Map) SortKeys() []string {
	var keys sort.StringSlice
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	return keys
}

/*XML transfer map to XML */
func (m Map) XML() []byte {
	v, e := xml.Marshal(&m)
	if e != nil {
		panic("map to xml error")
	}
	return v
}

/*ParseXML parse XML bytes to map */
func (m Map) ParseXML(b []byte) {
	toMap, err := xmlToMap(b, true)
	if err != nil {
		return
	}
	m.Join(toMap)
}

/*JSON transfer map to JSON */
func (m Map) JSON() []byte {
	v, e := json.Marshal(m)
	if e != nil {
		panic("map to json error")
	}
	return v
}

/*ParseJSON parse JSON bytes to map */
func (m Map) ParseJSON(b []byte) Map {
	tmp := Map{}
	if e := json.Unmarshal(b, &tmp); e == nil {
		m.join(tmp, true)
	}
	return m
}

/*URLEncode transfer map to url encode */
func (m Map) URLEncode() string {
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

func (m Map) join(source Map, replace bool) Map {
	for k, v := range source {
		if replace || !m.Has(k) {
			m.Set(k, v)
		}
	}
	return m
}

// Append ...
func (m Map) Append(p Map) Map {
	for k, v := range p {
		if m.Has(k) {
			m.Set(k, []interface{}{m.Get(k), v})
		} else {
			m.Set(k, v)
		}
	}
	return m
}

/*ReplaceJoin insert map s to m with replace */
func (m Map) ReplaceJoin(s Map) Map {
	return m.join(s, true)
}

/*Join insert map s to m with out replace */
func (m Map) Join(s Map) Map {
	return m.join(s, false)
}

/*Only get map with keys */
func (m Map) Only(keys []string) Map {
	p := Map{}
	size := len(keys)
	for i := 0; i < size; i++ {
		p.Set(keys[i], m.Get(keys[i]))
	}

	return p
}

/*Expect get map expect keys */
func (m Map) Expect(keys []string) Map {
	p := m.Clone()
	size := len(keys)
	for i := 0; i < size; i++ {
		p.Delete(keys[i])
	}

	return p
}

/*Clone copy a map */
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

// GoMap trans return a map[string]interface from Map
func (m Map) GoMap() map[string]interface{} {
	return m
}

// ToMap implements MapAble
func (m Map) ToMap() Map {
	return m
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

func marshalXML(maps Map, e *xml.Encoder, start xml.StartElement) error {
	if maps == nil {
		return errors.New("map is nil")
	}
	err := e.EncodeToken(start)
	if err != nil {
		return err
	}
	for k, v := range maps {
		err := convertXML(k, v, e, xml.StartElement{Name: xml.Name{Local: k}})
		if err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

func unmarshalXML(maps Map, d *xml.Decoder, start xml.StartElement, needCast bool) error {
	current := ""
	var data interface{}
	last := ""
	arrayTmp := make(Map)
	arrayTag := ""
	var ele []string

	for t, err := d.Token(); err == nil; t, err = d.Token() {
		switch token := t.(type) {
		case xml.StartElement:
			if strings.ToLower(token.Name.Local) == "xml" ||
				strings.ToLower(token.Name.Local) == "root" {
				continue
			}
			ele = append(ele, token.Name.Local)
			current = strings.Join(ele, ".")
			if current == last {
				arrayTag = current
				tmp := maps.Get(arrayTag)
				switch tmp.(type) {
				case []interface{}:
					arrayTmp.Set(arrayTag, tmp)
				default:
					arrayTmp.Set(arrayTag, []interface{}{tmp})
				}
				maps.Delete(arrayTag)
			}
		case xml.EndElement:
			name := token.Name.Local
			if strings.ToLower(name) == "xml" ||
				strings.ToLower(name) == "root" {
				break
			}
			last = strings.Join(ele, ".")
			if current == last {
				if data != nil {
					maps.Set(current, data)
				} else {
				}
				data = nil
			}
			if last == arrayTag {
				arr := arrayTmp.GetArray(arrayTag)
				if arr != nil {
					if v := maps.Get(arrayTag); v != nil {
						maps.Set(arrayTag, append(arr, v))
					} else {
						maps.Set(arrayTag, arr)
					}
				} else {
					//exception doing
					maps.Set(arrayTag, []interface{}{maps.Get(arrayTag)})
				}
				arrayTmp.Delete(arrayTag)
				arrayTag = ""
			}

			ele = ele[:len(ele)-1]
		case xml.CharData:
			if needCast {
				data, err = strconv.Atoi(string(token))
				if err == nil {
					continue
				}

				data, err = strconv.ParseFloat(string(token), 64)
				if err == nil {
					continue
				}

				data, err = strconv.ParseBool(string(token))
				if err == nil {
					continue
				}
			}

			data = string(token)
		default:
		}

	}

	return nil
}

func convertXML(k string, v interface{}, e *xml.Encoder, start xml.StartElement) error {
	var err error
	switch v1 := v.(type) {
	case Map:
		return marshalXML(v1, e, xml.StartElement{Name: xml.Name{Local: k}})
	case map[string]interface{}:
		return marshalXML(v1, e, xml.StartElement{Name: xml.Name{Local: k}})
	case string:
		if _, err := strconv.ParseInt(v1, 10, 0); err != nil {
			err = e.EncodeElement(
				CDATA{Value: v1}, xml.StartElement{Name: xml.Name{Local: k}})
			return err
		}
		err = e.EncodeElement(v1, xml.StartElement{Name: xml.Name{Local: k}})
		return err
	case float64:
		if v1 == float64(int64(v1)) {
			err = e.EncodeElement(int64(v1), xml.StartElement{Name: xml.Name{Local: k}})
			return err
		}
		err = e.EncodeElement(v1, xml.StartElement{Name: xml.Name{Local: k}})
		return err
	case bool:
		err = e.EncodeElement(v1, xml.StartElement{Name: xml.Name{Local: k}})
		return err
	case []interface{}:
		size := len(v1)
		for i := 0; i < size; i++ {
			err := convertXML(k, v1[i], e, xml.StartElement{Name: xml.Name{Local: k}})
			if err != nil {
				return err
			}
		}
		//add a null string to []
		if size == 1 {
			return convertXML(k, "", e, xml.StartElement{Name: xml.Name{Local: k}})
		}
	default:
	}
	return nil
}
func mapToXML(maps Map, needHeader bool) ([]byte, error) {
	buff := bytes.NewBuffer([]byte(CustomHeader))
	if needHeader {
		buff.Write([]byte(xml.Header))
	}

	enc := xml.NewEncoder(buff)
	err := marshalXML(maps, enc, xml.StartElement{Name: xml.Name{Local: "xml"}})
	if err != nil {
		return nil, err
	}
	err = enc.Flush()
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
func xmlToMap(contentXML []byte, hasHeader bool) (Map, error) {
	m := make(Map)
	dec := xml.NewDecoder(bytes.NewReader(contentXML))
	err := unmarshalXML(m, dec, xml.StartElement{Name: xml.Name{Local: "xml"}}, true)
	if err != nil {
		return nil, fmt.Errorf("xml to map:%w", err)
	}

	return m, nil
}

/*ParseNumber parse interface to number */
func ParseNumber(v interface{}) (float64, bool) {
	switch v0 := v.(type) {
	case float64:
		return v0, true
	case float32:
		return float64(v0), true
	}
	return 0, false
}

/*ParseInt parse interface to int64 */
func ParseInt(v interface{}) (int64, bool) {
	switch v0 := v.(type) {
	case int:
		return int64(v0), true
	case int32:
		return int64(v0), true
	case int64:
		return int64(v0), true
	case uint:
		return int64(v0), true
	case uint32:
		return int64(v0), true
	case uint64:
		return int64(v0), true
	case float64:
		return int64(v0), true
	case float32:
		return int64(v0), true
	default:
	}
	return 0, false
}

/*ParseString parse interface to string */
func ParseString(v interface{}) (string, bool) {
	switch v0 := v.(type) {
	case string:
		return v0, true
	case []byte:
		return string(v0), true
	case bytes.Buffer:
		return v0.String(), true
	default:
	}
	return "", false
}
