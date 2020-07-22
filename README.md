# map

## this is a map extension package ##
 you can use it transfers a map to xmp or to json or to struct, you can use it for transfer map<=>struct<=>xml 
 its implements the sub map value edit , you can use "." to edit the sub map 

```go
    //create a map
    m:=New()
    //set a value to map
    m.Set("key","value")
    //set an array to map
    m.Set("array_key",[]string{"value1","value2"})
    //set a value to key->sub_key
    m.Set("key.sub_key","value")
    

    //get a value from map
    v :=m.Get("key")
    //get an array from map
    v := m.GetStringArray("array_key")
    //get a value from map with default
    v := m.GetD("key","default_value")

    //marshal map to json
    v.ToJSON()
    //marshal map to xml
    v.ToXML()

    //get the map copy with deep copy
    v.Clone()

    //copy the value from m to v
    v.Join(m)

    //transfer the map to map[string]interface
    v.ToGoMap()

    //merge all values from v to m
    Merge(m,v)
    
    //transfer the struct to map
    StructToMap(&strcut{/*your strcut*/})

    type Struct1 struct{
        Value string 
}           
    type Struct2 struct {
        S1 Struct1
}

    exampleWithSubMap := StructToMap(&Struct2{})
    //to get the Value
    exampleWithSubMap.GetString("S1.Value")




```