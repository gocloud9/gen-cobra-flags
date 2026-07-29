package adaptors_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/gocloud9/gen-cobra-flags/sdk/pkg/adaptors"
)

func TestResolveConfigInput(t *testing.T) {
	t.Run("empty returns nil without error", func(t *testing.T) {
		got, err := adaptors.ResolveConfigInput("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("got %q, want nil", string(got))
		}
	})

	t.Run("absolute path reads file contents", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config.json")
		content := `{"name":"foo","count":3}`
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatalf("writing temp file: %v", err)
		}
		got, err := adaptors.ResolveConfigInput(p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != content {
			t.Errorf("got %q, want %q", string(got), content)
		}
	})

	t.Run("yaml file contents", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config.yaml")
		content := "name: bar\ncount: 7\n"
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatalf("writing temp file: %v", err)
		}
		got, err := adaptors.ResolveConfigInput(p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != content {
			t.Errorf("got %q, want %q", string(got), content)
		}
	})

	t.Run("relative path reads file contents", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"name":"rel","count":1}`
		if err := os.WriteFile(filepath.Join(dir, "rel.json"), []byte(content), 0o600); err != nil {
			t.Fatalf("writing temp file: %v", err)
		}
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		t.Cleanup(func() { _ = os.Chdir(wd) })
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		got, err := adaptors.ResolveConfigInput("rel.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != content {
			t.Errorf("got %q, want %q", string(got), content)
		}
	})

	t.Run("inline json content when not a path", func(t *testing.T) {
		content := `{"name":"inline","count":9}`
		got, err := adaptors.ResolveConfigInput(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != content {
			t.Errorf("got %q, want %q", string(got), content)
		}
	})

	t.Run("directory is treated as inline content", func(t *testing.T) {
		dir := t.TempDir()
		got, err := adaptors.ResolveConfigInput(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != dir {
			t.Errorf("got %q, want %q", string(got), dir)
		}
	})
}

func TestJsonOrYamlToStruct(t *testing.T) {
	type payload struct {
		Name  string `json:"name" yaml:"name"`
		Count int    `json:"count" yaml:"count"`
	}

	t.Run("json", func(t *testing.T) {
		got, err := adaptors.JsonOrYamlToStruct[payload]([]byte(`{"name":"foo","count":3}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := payload{Name: "foo", Count: 3}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("yaml", func(t *testing.T) {
		got, err := adaptors.JsonOrYamlToStruct[payload]([]byte("name: bar\ncount: 7\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := payload{Name: "bar", Count: 7}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("nil yields zero value without error", func(t *testing.T) {
		got, err := adaptors.JsonOrYamlToStruct[payload](nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != (payload{}) {
			t.Errorf("got %+v, want zero value", got)
		}
	})

	t.Run("empty bytes yield zero value without error", func(t *testing.T) {
		got, err := adaptors.JsonOrYamlToStruct[payload]([]byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != (payload{}) {
			t.Errorf("got %+v, want zero value", got)
		}
	})

	t.Run("invalid yaml returns error", func(t *testing.T) {
		_, err := adaptors.JsonOrYamlToStruct[payload]([]byte("name: : :\n  - bad"))
		if err == nil {
			t.Error("expected error for invalid input, got nil")
		}
	})
}

func TestStringToInteger(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{name: "positive", in: "42", want: 42},
		{name: "zero", in: "0", want: 0},
		{name: "negative", in: "-17", want: -17},
		{name: "large", in: "1000000", want: 1000000},
		{name: "not a number", in: "abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adaptors.StringToInteger[int](tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got value %d", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("StringToInteger(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestStringToInteger_TypeWidths(t *testing.T) {
	i32, err := adaptors.StringToInteger[int32]("2147483647")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if i32 != 2147483647 {
		t.Errorf("int32 = %d, want 2147483647", i32)
	}

	u, err := adaptors.StringToInteger[uint]("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != 123 {
		t.Errorf("uint = %d, want 123", u)
	}
}

func TestIntegerToString(t *testing.T) {
	got, err := adaptors.IntegerToString[int](42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "42" {
		t.Errorf("IntegerToString(42) = %q, want %q", got, "42")
	}

	got, err = adaptors.IntegerToString[int](-5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "-5" {
		t.Errorf("IntegerToString(-5) = %q, want %q", got, "-5")
	}
}

func TestStringToFloat(t *testing.T) {
	got, err := adaptors.StringToFloat[float64]("3.14")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3.14 {
		t.Errorf("StringToFloat(\"3.14\") = %v, want 3.14", got)
	}

	if _, err := adaptors.StringToFloat[float64]("not-a-float"); err == nil {
		t.Error("expected error for invalid float, got nil")
	}
}

func TestSliceToSlice(t *testing.T) {
	double := func(i int) (int, error) { return i * 2, nil }

	t.Run("maps each element with no padding", func(t *testing.T) {
		got, err := adaptors.SliceToSlice(double, []int{1, 2, 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []int{2, 4, 6}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SliceToSlice = %v, want %v", got, want)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		got, err := adaptors.SliceToSlice(double, []int{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("SliceToSlice([]) length = %d, want 0", len(got))
		}
	})

	t.Run("propagates errors", func(t *testing.T) {
		fail := func(string) (int, error) { return 0, strconv.ErrSyntax }
		_, err := adaptors.SliceToSlice(fail, []string{"x"})
		if err == nil {
			t.Error("expected error to propagate, got nil")
		}
	})
}

func TestStringMapToStringMap(t *testing.T) {
	toLen := func(s string) (int, error) { return len(s), nil }
	got, err := adaptors.StringMapToStringMap(toLen, map[string]string{"a": "xy", "b": "zzz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]int{"a": 2, "b": 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("StringMapToStringMap = %v, want %v", got, want)
	}
}

func TestToPtr(t *testing.T) {
	got, err := adaptors.ToPtr(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != 42 {
		t.Errorf("ToPtr(42) = %v, want pointer to 42", got)
	}
}

func TestIPRoundTrip(t *testing.T) {
	s, err := adaptors.StringToIP("192.168.1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := adaptors.IPToString(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "192.168.1.1" {
		t.Errorf("IP round-trip = %q, want %q", out, "192.168.1.1")
	}
}

func TestTimeRoundTrip(t *testing.T) {
	in := "2024-01-02T15:04:05Z"
	tm, err := adaptors.StringToTime(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := adaptors.TimeToString(tm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != in {
		t.Errorf("time round-trip = %q, want %q", out, in)
	}
}

func TestDurationRoundTrip(t *testing.T) {
	d, err := adaptors.StringToDuration("1h30m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 90*time.Minute {
		t.Errorf("StringToDuration(\"1h30m\") = %v, want 1h30m", d)
	}
	s, err := adaptors.DurationToString(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "1h30m0s" {
		t.Errorf("DurationToString = %q, want %q", s, "1h30m0s")
	}
}

func TestIPNetRoundTrip(t *testing.T) {
	n, err := adaptors.StringToIPNet("10.0.0.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := adaptors.IPNetToString(n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "10.0.0.0/24" {
		t.Errorf("IPNet round-trip = %q, want %q", out, "10.0.0.0/24")
	}
}

func TestStringToBool(t *testing.T) {
	truthy := []string{"true", "1", "yes", "on"}
	for _, s := range truthy {
		b, err := adaptors.StringToBool(s)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", s, err)
		}
		if !b {
			t.Errorf("StringToBool(%q) = false, want true", s)
		}
	}
	falsy := []string{"false", "0", "no", "off"}
	for _, s := range falsy {
		b, err := adaptors.StringToBool(s)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", s, err)
		}
		if b {
			t.Errorf("StringToBool(%q) = true, want false", s)
		}
	}
}

func TestBoolToString(t *testing.T) {
	if s, _ := adaptors.BoolToString(true); s != "true" {
		t.Errorf("BoolToString(true) = %q, want \"true\"", s)
	}
	if s, _ := adaptors.BoolToString(false); s != "false" {
		t.Errorf("BoolToString(false) = %q, want \"false\"", s)
	}
}

func TestNumericConversions(t *testing.T) {
	if v, _ := adaptors.IntegerToInteger[int, int64](5); v != 5 {
		t.Errorf("IntegerToInteger = %d, want 5", v)
	}
	if v, _ := adaptors.FloatToFloat[float32, float64](1.5); v != 1.5 {
		t.Errorf("FloatToFloat = %v, want 1.5", v)
	}
	if v, _ := adaptors.FloatToInteger[float64, int](3.9); v != 3 {
		t.Errorf("FloatToInteger(3.9) = %d, want 3", v)
	}
	if v, _ := adaptors.IntegerToFloat[int, float64](7); v != 7 {
		t.Errorf("IntegerToFloat = %v, want 7", v)
	}
	if v, _ := adaptors.BoolToInteger[int](true); v != 1 {
		t.Errorf("BoolToInteger(true) = %d, want 1", v)
	}
	if v, _ := adaptors.IntegerToBool[int](0); v {
		t.Errorf("IntegerToBool(0) = true, want false")
	}
}
