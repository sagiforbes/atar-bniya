package ottoutils

import (
	"encoding/json"

	"github.com/robertkrimen/otto"
)

//Val2Struct convert otto value to goplang struct
func Val2Struct(v otto.Value, i interface{}) error {
	asI, e := v.Export()
	if e != nil {
		return e
	}
	b, e := json.Marshal(asI)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, i)
}
