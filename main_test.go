package main

import (
	"testing"
)

var (
	testRules = []EnfolderRule{
		{"", nil},
		{"folder1", nil},
		{"folder2", []string{}},
		{"folder3", []string{""}},
		{"folder4", []string{"keyword4a"}},
		{"folder5", []string{"keyword5a", "keyword5b"}},
	}
)

func TestGetDestinationFolderName(t *testing.T) {
	type args struct {
		fileName string
		rules    []EnfolderRule
	}
	tests := []struct {
		name                      string
		args                      args
		wantDestinationFolderName string
	}{
		{`"" -> ""`, args{"", testRules}, ""},
		{`"keyword2" -> ""`, args{"keyword2", testRules}, ""},
		{`"keyword3" -> ""`, args{"keyword3", testRules}, ""},
		{`"keyword4a" -> "folder4"`, args{"keyword4a", testRules}, "folder4"},
		{`"keyword5a" -> "folder5"`, args{"keyword5a", testRules}, "folder5"},
		{`"keyword5b" -> "folder5"`, args{"keyword5b", testRules}, "folder5"},
		{`"folder5" -> ""`, args{"folder5", testRules}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDestinationFolderName := GetDestinationFolderName(tt.args.fileName, tt.args.rules); gotDestinationFolderName != tt.wantDestinationFolderName {
				t.Errorf("GetDestinationFolderName() = %v, want %v", gotDestinationFolderName, tt.wantDestinationFolderName)
			}
		})
	}
}
