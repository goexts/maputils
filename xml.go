package gomap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//CDATA xml cdata defines
type CDATA struct {
	XMLName xml.Name
	Value   string `xml:",cdata"`
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

func xmlToMap(maps Map, contentXML []byte, hasHeader bool) error {
	dec := xml.NewDecoder(bytes.NewReader(contentXML))
	err := unmarshalXML(maps, dec, xml.StartElement{Name: xml.Name{Local: "xml"}}, true)
	if err != nil {
		return fmt.Errorf("xml to map error:%w", err)
	}
	return nil
}
