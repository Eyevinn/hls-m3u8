package m3u8

import (
	"reflect"
	"testing"
)

func TestDecodeAttributes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Attribute
	}{
		{
			name:  "Normal attributes string",
			input: `AVERAGE-BANDWIDTH=20985770,VIDEO-RANGE=SDR,CODECS="hvc1.2.4.L150.B0",RESOLUTION=3840x2160`,
			want: []Attribute{
				Attribute{
					Key: "AVERAGE-BANDWIDTH",
					Val: "20985770",
				},
				Attribute{
					Key: "VIDEO-RANGE",
					Val: "SDR",
				},
				Attribute{
					Key: "CODECS",
					Val: `"hvc1.2.4.L150.B0"`,
				},
				Attribute{
					Key: "RESOLUTION",
					Val: "3840x2160",
				},
			},
		},
		{
			name:  "Spaces in attributes string",
			input: ` AVERAGE-BANDWIDTH=20985770, VIDEO-RANGE=SDR, CODECS="hvc1.2.4.L150.B0", RESOLUTION=3840x2160`,
			want: []Attribute{
				Attribute{
					Key: "AVERAGE-BANDWIDTH",
					Val: "20985770",
				},
				Attribute{
					Key: "VIDEO-RANGE",
					Val: "SDR",
				},
				Attribute{
					Key: "CODECS",
					Val: `"hvc1.2.4.L150.B0"`,
				},
				Attribute{
					Key: "RESOLUTION",
					Val: "3840x2160",
				},
			},
		},
		{
			name:  "Missing value in attributes string",
			input: `AVERAGE-BANDWIDTH=,VIDEO-RANGE=SDR,CODECS="hvc1.2.4.L150.B0",RESOLUTION=3840x2160`,
			want: []Attribute{
				Attribute{
					Key: "AVERAGE-BANDWIDTH",
					Val: "",
				},
				Attribute{
					Key: "VIDEO-RANGE",
					Val: "SDR",
				},
				Attribute{
					Key: "CODECS",
					Val: `"hvc1.2.4.L150.B0"`,
				},
				Attribute{
					Key: "RESOLUTION",
					Val: "3840x2160",
				},
			},
		},
		{
			name:  "Comma in quoted value",
			input: `TEST="abc,123"`,
			want: []Attribute{
				Attribute{
					Key: "TEST",
					Val: `"abc,123"`,
				},
			},
		},
		{
			name:  "Empty input",
			input: ``,
			want:  []Attribute{},
		},
		{
			name:  "Malformed key",
			input: `CODECÅÄÖ=123,VIDEO-RANGE=SDR`,
			want: []Attribute{
				Attribute{
					Key: "VIDEO-RANGE",
					Val: "SDR",
				},
			},
		},
		{
			name:  "No key",
			input: `123"hello",VIDEO-RANGE=SDR`,
			want: []Attribute{
				Attribute{
					Key: "VIDEO-RANGE",
					Val: "SDR",
				},
			},
		},
		{
			name:  "No value on last entry (out of bounds test)",
			input: `CODEC=123,VIDEO-RANGE=`,
			want: []Attribute{
				Attribute{
					Key: "CODEC",
					Val: "123",
				},
				Attribute{
					Key: "VIDEO-RANGE",
					Val: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeAttributes(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeAttributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindEndsInNumber(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantString string
		wantBool   bool
	}{
		{
			name:       "String with one number in the middle",
			input:      `This is a string with 100% potential`,
			wantString: "100",
			wantBool:   false,
		},
		{
			name:       "String with two numbers in the middle",
			input:      `This is a string with 100% potential, but 50% success rate`,
			wantString: "50",
			wantBool:   false,
		},
		{
			name:       "String with no numbers in the middle",
			input:      `This is a string with no potential`,
			wantString: "",
			wantBool:   false,
		},
		{
			name:       "String that ends in a number",
			input:      `This is a string has not 1 number, but 2`,
			wantString: "2",
			wantBool:   true,
		},
		{
			name:       "Empty string",
			input:      ``,
			wantString: "",
			wantBool:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numberString, endsInNumber := findEndsInNumber(tt.input)
			if numberString != tt.wantString || endsInNumber != tt.wantBool {
				t.Errorf("findEndsInNumber() got = %v,%v want %v,%v", numberString, endsInNumber, tt.wantString, tt.wantBool)
			}
		})
	}
}

func BenchmarkDecodeAttributes(b *testing.B) {
	line := `AVERAGE-BANDWIDTH=20985770,BANDWIDTH=28058971,VIDEO-RANGE=SDR,CODECS="hvc1.2.4.L150.B0",RESOLUTION=3840x2160,FRAME-RATE=23.976,CLOSED-CAPTIONS=NONE,HDCP-LEVEL=TYPE-1`

	for i := 0; i < b.N; i++ {
		_ = decodeAttributes(line)
	}
}

func BenchmarkFindEndsInString(b *testing.B) {
	line := `This is a string has not 1 number, but 2`

	for i := 0; i < b.N; i++ {
		_, _ = findEndsInNumber(line)
	}
}
