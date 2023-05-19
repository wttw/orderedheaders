// Code generated by "enumer -json -trimprefix=HeaderType -transform=kebab -type HeaderType"; DO NOT EDIT.

package orderedheaders

import (
	"encoding/json"
	"fmt"
)

const _HeaderTypeName = "unstructuredmailboxmailbox-listdatereceivedmessage-idmessage-id-listphrase-listreturn-pathopaque"

var _HeaderTypeIndex = [...]uint8{0, 12, 19, 31, 35, 43, 53, 68, 79, 90, 96}

func (i HeaderType) String() string {
	if i < 0 || i >= HeaderType(len(_HeaderTypeIndex)-1) {
		return fmt.Sprintf("HeaderType(%d)", i)
	}
	return _HeaderTypeName[_HeaderTypeIndex[i]:_HeaderTypeIndex[i+1]]
}

var _HeaderTypeValues = []HeaderType{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

var _HeaderTypeNameToValueMap = map[string]HeaderType{
	_HeaderTypeName[0:12]:  0,
	_HeaderTypeName[12:19]: 1,
	_HeaderTypeName[19:31]: 2,
	_HeaderTypeName[31:35]: 3,
	_HeaderTypeName[35:43]: 4,
	_HeaderTypeName[43:53]: 5,
	_HeaderTypeName[53:68]: 6,
	_HeaderTypeName[68:79]: 7,
	_HeaderTypeName[79:90]: 8,
	_HeaderTypeName[90:96]: 9,
}

// HeaderTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func HeaderTypeString(s string) (HeaderType, error) {
	if val, ok := _HeaderTypeNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to HeaderType values", s)
}

// HeaderTypeValues returns all values of the enum
func HeaderTypeValues() []HeaderType {
	return _HeaderTypeValues
}

// IsAHeaderType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i HeaderType) IsAHeaderType() bool {
	for _, v := range _HeaderTypeValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for HeaderType
func (i HeaderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for HeaderType
func (i *HeaderType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("HeaderType should be a string, got %s", data)
	}

	var err error
	*i, err = HeaderTypeString(s)
	return err
}
