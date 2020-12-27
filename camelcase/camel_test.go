package camelcase

import (
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestToUpperCamel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Empty string", "", ""},
		{"Single word", "hello", "Hello"},
		{"Snake case", "hello_world", "HelloWorld"},
		{"Kebab case", "hello-world", "HelloWorld"},
		{"All caps", "HTTP_SERVER", "HTTPServer"},
		{"Mixed case", "mySQL_Query", "MySQLQuery"},
		{"With numbers", "user_id_2", "UserID2"},
		{"Abbreviations", "http_request", "HTTPRequest"},
		{"Unicode", "こんにちは_世界", "こんにちは世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToUpperCamel(tt.input); got != tt.want {
				t.Errorf("ToUpperCamel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToLowerCamel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Empty string", "", ""},
		{"Single word", "Hello", "hello"},
		{"Snake case", "hello_world", "helloWorld"},
		{"All caps", "HTTP_SERVER", "httpServer"},
		{"With numbers", "USER_ID_2", "userID2"},
		{"Abbreviations", "HTTP_REQUEST", "httpRequest"},
		{"Mixed case", "MySQL_Query", "mysqlQuery"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLowerCamel(tt.input); got != tt.want {
				t.Errorf("ToLowerCamel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToUnderScore(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Empty string", "", ""},
		{"Single word", "Hello", "hello"},
		{"Camel case", "helloWorld", "hello_world"},
		{"All caps", "HttpServer", "http_server"},
		{"All caps", "HTTPServer", "http_server"},
		{"With numbers", "UserID2", "user_id_2"},
		{"Abbreviations", "HTTPRequest", "http_request"},
		{"Mixed case", "MySQLQuery", "mysql_query"},
		{"Consecutive caps", "MySSHKey", "my_ssh_key"},
		{"Unicode", "こんにちはWorld", "こんにちは_world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToUnderScore(tt.input); got != tt.want {
				t.Errorf("ToUnderScore(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("Multiple underscores", func(t *testing.T) {
		if got := ToUpperCamel("hello__world"); got != "HelloWorld" {
			t.Errorf("ToUpperCamel('hello__world') = %q, want 'HelloWorld'", got)
		}
	})

	t.Run("Leading underscore", func(t *testing.T) {
		if got := ToUpperCamel("_hello_world"); got != "HelloWorld" {
			t.Errorf("ToUpperCamel('_hello_world') = %q, want 'HelloWorld'", got)
		}
	})

	t.Run("All letters uppercase", func(t *testing.T) {
		if got := ToUnderScore("HELLOWORLD"); got != "helloworld" {
			t.Errorf("ToUnderScore('HELLOWORLD') = %q, want 'helloworld'", got)
		}
	})
}

func BenchmarkToUpperCamel(b *testing.B) {
	testString := "hello_world_this_is_a_benchmark_test"
	for i := 0; i < b.N; i++ {
		ToUpperCamel(testString)
	}
}

func BenchmarkToLowerCamel(b *testing.B) {
	testString := "HELLO_WORLD_THIS_IS_A_BENCHMARK_TEST"
	for i := 0; i < b.N; i++ {
		ToLowerCamel(testString)
	}
}

func BenchmarkToUnderScore(b *testing.B) {
	testString := "HelloWorldThisIsABenchmarkTest"
	for i := 0; i < b.N; i++ {
		ToUnderScore(testString)
	}
}

func FuzzCamelCase(f *testing.F) {
	f.Add("hello_world")
	f.Add("HTTPRequest")
	f.Add("userID2")

	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			t.Skip()
		}

		// Test roundtrip conversions
		upper := ToUpperCamel(s)
		roundtrip := ToUnderScore(upper)
		upper2 := ToUpperCamel(roundtrip)

		if upper != upper2 {
			t.Errorf("Roundtrip failed: original(%q) -> upper(%q) -> underscore(%q) -> upper2(%q)",
				s, upper, roundtrip, upper2)
		}

		// Check all letters are valid
		for _, r := range upper {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				t.Fatalf("Invalid character in result: %q", r)
			}
		}
	})
}
