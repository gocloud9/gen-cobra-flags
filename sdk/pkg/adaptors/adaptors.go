package adaptors

import (
	"fmt"
	"net"
	"regexp"
	"time"
)

func IPToString(ip net.IP) (string, error) {
	return ip.String(), nil
}

func StringToIP(ip string) (net.IP, error) {
	return net.ParseIP(ip), nil
}

func TimeToString(t time.Time) (string, error) {
	return t.Format(time.RFC3339), nil
}

func StringToTime(t string) (time.Time, error) {
	return time.Parse(time.RFC3339, t)
}

func DurationToString(dur time.Duration) (string, error) {
	return dur.String(), nil
}

func StringToDuration(dur string) (time.Duration, error) {
	return time.ParseDuration(dur)
}

func IPNetToString(ipnet *net.IPNet) (string, error) {
	return ipnet.String(), nil
}

func StringToIPNet(ipnet string) (*net.IPNet, error) {
	_, network, err := net.ParseCIDR(ipnet)

	return network, err
}

func StringToBool(s string) (bool, error) {
	switch s {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, nil
	}
}

func BoolToString(b bool) (string, error) {
	if b {
		return "true", nil
	}

	return "false", nil
}

func IntegerToString[IN Integer](i int) (string, error) {
	return fmt.Sprintf("%d", i), nil
}

func StringToInteger[OUT Integer](s string) (OUT, error) {
	var i int64
	o, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0, err
	}

	return OUT(o), nil
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func IntegerToInteger[IN Integer, OUT Integer](i IN) (OUT, error) {
	return OUT(i), nil
}

type Float interface {
	~float32 | ~float64
}

func FloatToFloat[IN Float, OUT Float](f IN) (OUT, error) {
	return OUT(f), nil
}

func FloatToString[IN Float](f IN) (string, error) {
	return fmt.Sprintf("%f", f), nil
}

func StringToFloat[OUT Float](s string) (OUT, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return 0, err
	}

	return OUT(f), nil
}

func FloatToInteger[IN Float, OUT Integer](f IN) (OUT, error) {
	return OUT(f), nil
}

func IntegerToFloat[IN Integer, OUT Float](i IN) (OUT, error) {
	return OUT(i), nil
}

func BoolToInteger[OUT Integer](b bool) (OUT, error) {
	if b {
		return OUT(1), nil
	}

	return OUT(0), nil
}

func IntegerToBool[IN Integer](i IN) (bool, error) {
	if i != 0 {
		return true, nil
	}

	return false, nil
}

func SliceToSlice[IN any, OUT any](f func(IN) (OUT, error), in []IN) ([]OUT, error) {
	out := make([]OUT, len(in))
	for i := range in {
		o, err := f(in[i])
		if err != nil {
			return nil, err
		}

		out = append(out, o)
	}

	return out, nil
}

func StringMapToStringMap[IN any, OUT any](f func(IN) (OUT, error), in map[string]IN) (map[string]OUT, error) {
	out := make(map[string]OUT)
	for i := range in {
		o, err := f(in[i])
		if err != nil {
			return nil, err
		}

		out[i] = o
	}

	return out, nil
}

func ToPtr[T any](t T) (*T, error) {
	return &t, nil
}

func GetFuncNameByTypeNames(typeNameIn, typeNameOut string) string {
	funcMap := map[string]map[string]string{
		"net.IP": {
			"string": "IPToString",
		},
		"string": {
			"net.IP":        "StringToIP",
			"time.Time":     "StringToTime",
			"time.Duration": "StringToDuration",
			"net.IPNet":     "StringToIPNet",
			"bool":          "StringToBool",
			"int*":          "StringToInteger",
			"uint*":         "StringToInteger",
			"float*":        "StringToFloat",
		},
		"time.Time": {
			"string": "TimeToString",
		},
		"time.Duration": {
			"string": "DurationToString",
		},
		"net.IPNet": {
			"string": "IPNetToString",
		},
		"bool": {
			"string": "BoolToString",
			"int*":   "BoolToInteger",
			"uint*":  "BoolToInteger",
		},
		"int*": {
			"string": "IntegerToString",
			"int*":   "IntegerToInteger",
			"uint*":  "IntegerToInteger",
			"float*": "IntegerToFloat",
		},
		"uint*": {
			"string": "IntegerToString",
			"int*":   "IntegerToInteger",
			"uint*":  "IntegerToInteger",
			"float*": "IntegerToFloat",
		},
	}

	inMap, ok := funcMap[convertToWildcardType(typeNameIn)]
	if !ok {
		return ""
	}

	funcName, ok := inMap[convertToWildcardType(typeNameOut)]
	if !ok {
		return ""
	}

	return funcName
}

func convertToWildcardType(typeName string) string {
	if regexp.MustCompile("int[0-9]+").MatchString(typeName) {
		return "int*"
	}

	if regexp.MustCompile("uint[0-9]+").MatchString(typeName) {
		return "uint*"
	}

	if regexp.MustCompile("float[0-9]+").MatchString(typeName) {
		return "float*"
	}

	return typeName
}
