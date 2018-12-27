package tiny

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

const (
	Required    = "required"
	NotRequired = "not required"
)

// User is the object that holds the data for user
type User struct {
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	MiddleName  string `json:"middleName,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Email       string `json:"email,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	CityOfBirth string `json:"cityOfBirth,omitempty"`
}

func DeEval(jsonString string) uint64 {
	u := User{}
	var result uint64
	err := json.Unmarshal([]byte(jsonString), &u)
	if err != nil {
		return result
	}

	elem := reflect.ValueOf(&u).Elem()
	typeOfUser := elem.Type()

	sets := GetSettings()
	for _, set := range sets {

		for i := 0; i < elem.NumField(); i++ {
			f := elem.Field(i)
			value := f.Interface().(string)
			if value == "" {
				continue
			}

			result += set[typeOfUser.Field(i).Name+value]
		}
	}

	return result
}

func NewConfig(keys ...string) string {
	sets := GetSettings()
	results := ""
	var result uint64
	for _, set := range sets {
		for _, key := range keys {
			result += set[key]
		}
		results += strconv.FormatUint(result, 10) + ","
	}
	if results[len(results)-1:len(results)] == "," {
		results = results[:len(results)-1]
	}

	return results
}

func NewUser() User {
	u := User{}

	elem := reflect.ValueOf(&u).Elem()
	typeOfUser := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		f.SetString(typeOfUser.Field(i).Name)
	}
	return u
}

func setIfOneTrue(conf, a, b uint64) string {
	if conf&b != 0 {
		return Required
	}
	if conf&a != 0 {
		return NotRequired
	}
	return ""
}

func parseConfigs(confs string) ([]uint64, string) {
	result := []uint64{}
	sconf := strings.Split(confs, ",")
	if len(sconf) == 0 {
		return result, `{"error": "cannot parse config: ` + confs + `"}`
	}
	for _, val := range sconf {
		val = strings.TrimSpace(val)
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return []uint64{}, `{"error": "cannot parse value of config: ` + val + `"}`
		}
		result = append(result, n)
	}

	return result, ""
}

type settings map[string]uint64

// var rw sync.Mutex

func GetSettings() []settings {
	u := User{}

	elem := reflect.ValueOf(&u).Elem()
	typeOfUser := elem.Type()
	cfgs := make([]settings, cLen(elem.NumField()))

	for index, cfg := range cfgs {
		cindex := index
		cfg = make(settings)
		cindex = cindex * 63
		if cindex != 0 {
			cindex++
		}
		elementNumber := 0
		for i := 0; i <= 63; i++ {
			if i >= elem.NumField()*2 {
				continue
			}
			value := NotRequired
			if i >= elem.NumField() {
				value = Required
			}
			if elementNumber >= elem.NumField() {
				elementNumber = 0
			}
			cfg[typeOfUser.Field(elementNumber+cindex).Name+value] = 1 << uint64(i)
			elementNumber++
		}
		cfgs[index] = cfg
	}

	return cfgs
}

func cLen(cfg int) int {
	var sum int
	for {
		sum++
		cfg -= 63
		if cfg <= 0 {
			return sum
		}
	}
	return 0
}

func Eval(confs string) string {
	cfgs, errString := parseConfigs(confs)
	if errString != "" {
		return errString
	}

	sets := GetSettings()
	// if len(sets) != len(cfgs) {
	// 	return `{"error": "configs length shoud be equal to length of settings:. configs length:` + string(len(cfgs)) + `, settings length:` + string(len(sets)) + `"}`
	// }
	u := User{}

	setUser(&u, cfgs, sets)

	data, err := json.Marshal(u)
	if err != nil {
		return `{"error": "cannot unmarshal object: ` + err.Error() + `"}`
	}
	return string(data)
}

func setUser(u *User, cfgs []uint64, sets []settings) {

	elem := reflect.ValueOf(u).Elem()
	typeOfUser := elem.Type()

	for index, set := range sets {
		cindex := index
		cindex = cindex * 63
		if cindex != 0 {
			cindex++
		}
		for i := 0; i <= 63; i++ {
			if i >= elem.NumField() {
				return
			}
			f := elem.Field(i + cindex)
			if len(cfgs) < index+1 {
				break
			}
			value := setIfOneTrue(cfgs[index], set[typeOfUser.Field(i+cindex).Name+NotRequired], set[typeOfUser.Field(i+cindex).Name+Required])
			f.SetString(value)
		}
	}
}
