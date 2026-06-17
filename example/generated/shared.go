// Generated Code with gen-cobra-flags - Do Not Edit
package generated


var adaptorCustomInt32ToString func(int32) (string, error)

func RegisterAdaptorCustomInt32ToString(f func(int32) (string, error)) {
    adaptorCustomInt32ToString = f
}

var adaptorCustomInt64ToInt32 func(int64) (int32, error)

func RegisterAdaptorCustomInt64ToInt32(f func(int64) (int32, error)) {
    adaptorCustomInt64ToInt32 = f
}

var adaptorStringToMyCustomType func(string) (example.MyCustomType, error)

func RegisterAdaptorStringToMyCustomType(f func(string) (example.MyCustomType, error)) {
    adaptorStringToMyCustomType = f
}

