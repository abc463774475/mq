// Code generated by "enumer -type=BaseMsgID"; DO NOT EDIT.

package msg

import (
	"fmt"
	"strings"
)

const _BaseMsgIDName = "BaseMsgID_StartBaseMsgID_PlayerLoginBaseMsgID_End"

var _BaseMsgIDIndex = [...]uint8{0, 15, 36, 49}

const _BaseMsgIDLowerName = "basemsgid_startbasemsgid_playerloginbasemsgid_end"

func (i BaseMsgID) String() string {
	i -= 1
	if i < 0 || i >= BaseMsgID(len(_BaseMsgIDIndex)-1) {
		return fmt.Sprintf("BaseMsgID(%d)", i+1)
	}
	return _BaseMsgIDName[_BaseMsgIDIndex[i]:_BaseMsgIDIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _BaseMsgIDNoOp() {
	var x [1]struct{}
	_ = x[BaseMsgID_Start-(1)]
	_ = x[BaseMsgID_PlayerLogin-(2)]
	_ = x[BaseMsgID_End-(3)]
}

var _BaseMsgIDValues = []BaseMsgID{BaseMsgID_Start, BaseMsgID_PlayerLogin, BaseMsgID_End}

var _BaseMsgIDNameToValueMap = map[string]BaseMsgID{
	_BaseMsgIDName[0:15]:       BaseMsgID_Start,
	_BaseMsgIDLowerName[0:15]:  BaseMsgID_Start,
	_BaseMsgIDName[15:36]:      BaseMsgID_PlayerLogin,
	_BaseMsgIDLowerName[15:36]: BaseMsgID_PlayerLogin,
	_BaseMsgIDName[36:49]:      BaseMsgID_End,
	_BaseMsgIDLowerName[36:49]: BaseMsgID_End,
}

var _BaseMsgIDNames = []string{
	_BaseMsgIDName[0:15],
	_BaseMsgIDName[15:36],
	_BaseMsgIDName[36:49],
}

// BaseMsgIDString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func BaseMsgIDString(s string) (BaseMsgID, error) {
	if val, ok := _BaseMsgIDNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _BaseMsgIDNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to BaseMsgID values", s)
}

// BaseMsgIDValues returns all values of the enum
func BaseMsgIDValues() []BaseMsgID {
	return _BaseMsgIDValues
}

// BaseMsgIDStrings returns a slice of all String values of the enum
func BaseMsgIDStrings() []string {
	strs := make([]string, len(_BaseMsgIDNames))
	copy(strs, _BaseMsgIDNames)
	return strs
}

// IsABaseMsgID returns "true" if the value is listed in the enum definition. "false" otherwise
func (i BaseMsgID) IsABaseMsgID() bool {
	for _, v := range _BaseMsgIDValues {
		if i == v {
			return true
		}
	}
	return false
}
