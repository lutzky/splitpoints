package main

import (
	"fmt"
	"testing"
)

func TestTimeCode(t *testing.T) {
	testCases := []struct {
		in      string
		want    timecode
		wantErr error
	}{
		{"01:02:03:24", timecode{1, 2, 3, 24}, nil},
		{"1:02:03:24", timecode{1, 2, 3, 24}, nil},
		{"1:2:03:24", timecode{1, 2, 3, 24}, nil},
		{"1:09:03:24", timecode{1, 9, 3, 24}, nil}, // Not octal
		{"1:09:93:22", timecode{}, fmt.Errorf(`invalid timecode "1:09:93:22": number too high`)},
		{"1:09:03:26", timecode{}, fmt.Errorf(`invalid timecode "1:09:03:26": number too high`)},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseTimecode(tc.in)
			if err != tc.wantErr {
				if err == nil || tc.wantErr == nil {
					t.Fatalf("Want error %v, got %v", tc.wantErr, err)
				}
				if err.Error() != tc.wantErr.Error() {
					t.Fatalf("Want error %v, got %v", tc.wantErr, err)
				}
			}
			if tc.wantErr != nil {
				return
			}
			if got != tc.want {
				t.Errorf("Want %+v, got %+v", tc.want, got)
			}
		})
	}
}

func TestDecimal(t *testing.T) {
	testCases := []struct {
		in   timecode
		want string
	}{
		{timecode{1, 2, 3, 0}, "1:02:03.000"},
		{timecode{1, 2, 3, 1}, "1:02:03.040"},
		{timecode{1, 2, 3, 5}, "1:02:03.200"},
		{timecode{1, 2, 3, 4}, "1:02:03.160"},
		{timecode{1, 2, 3, 24}, "1:02:03.960"},
	}

	for _, tc := range testCases {
		got := tc.in.decimal()
		if got != tc.want {
			t.Errorf("%+v.decimal() = %q; want %q", tc.in, got, tc.want)
		}

	}
}

func TestCommand(t *testing.T) {
	testCases := []struct {
		in   string
		want string
	}{
		{
			"00:00:01:18-00:10:11:09",
			"ffmpeg -ss 0:00:01.720 -i input.mov -to 0:10:09.640 -c copy output.mov",
		},
		{
			"00:10:11:10-00:28:46:15",
			"ffmpeg -ss 0:10:11.400 -i input.mov -to 0:18:35.200 -c copy output.mov",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := splitCommand(tc.in, "input.mov", "output.mov")
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("splitCommand diff.\nGot:  %s\nWant: %s", got, tc.want)
			}
		})
	}
}
