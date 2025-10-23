package main

import (
	"strings"
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

func TestValidateFolderName(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		wantValid  bool
		wantReason string // substring to check in reason
	}{
		// Valid names
		{"valid simple name", "Documents", true, ""},
		{"valid with spaces", "My Documents", true, ""},
		{"valid with underscore", "my_folder", true, ""},
		{"valid with dash", "my-folder", true, ""},
		{"valid with numbers", "folder123", true, ""},
		{"valid with parentheses", "folder(1)", true, ""},
		{"valid with brackets", "folder[1]", true, ""},
		{"valid with unicode", "資料夾", true, ""},
		{"valid starts with dot", ".hidden", true, ""},

		// Invalid: empty
		{"empty name", "", false, "empty"},

		// Invalid: too long
		{"too long", strings.Repeat("a", 256), false, "255"},

		// Invalid: Windows reserved characters
		{"contains <", "folder<name", false, "invalid characters"},
		{"contains >", "folder>name", false, "invalid characters"},
		{"contains :", "folder:name", false, "invalid characters"},
		{"contains \"", "folder\"name", false, "invalid characters"},
		{"contains /", "folder/name", false, "invalid characters"},
		{"contains \\", "folder\\name", false, "invalid characters"},
		{"contains |", "folder|name", false, "invalid characters"},
		{"contains ?", "folder?name", false, "invalid characters"},
		{"contains *", "folder*name", false, "invalid characters"},

		// Invalid: ends with space or period
		{"ends with space", "folder ", false, "end with space or period"},
		{"ends with period", "folder.", false, "end with space or period"},
		{"ends with multiple periods", "folder...", false, "end with space or period"},

		// Invalid: reserved names (Windows)
		{"reserved CON", "CON", false, "reserved name"},
		{"reserved PRN", "PRN", false, "reserved name"},
		{"reserved AUX", "AUX", false, "reserved name"},
		{"reserved NUL", "NUL", false, "reserved name"},
		{"reserved COM1", "COM1", false, "reserved name"},
		{"reserved COM9", "COM9", false, "reserved name"},
		{"reserved LPT1", "LPT1", false, "reserved name"},
		{"reserved LPT9", "LPT9", false, "reserved name"},
		{"reserved con lowercase", "con", false, "reserved name"},
		{"reserved CON.txt", "CON.txt", false, "reserved name"},
		{"reserved aux.log", "aux.log", false, "reserved name"},

		// Invalid: Unix special names
		{"dot", ".", false, "cannot be '.' or '..'"},
		{"double dot", "..", false, "cannot be '.' or '..'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, gotReason := ValidateFolderName(tt.folderName)
			if gotValid != tt.wantValid {
				t.Errorf("ValidateFolderName(%q) valid = %v, want %v", tt.folderName, gotValid, tt.wantValid)
			}
			if !tt.wantValid && tt.wantReason != "" {
				if !strings.Contains(strings.ToLower(gotReason), strings.ToLower(tt.wantReason)) {
					t.Errorf("ValidateFolderName(%q) reason = %q, want to contain %q", tt.folderName, gotReason, tt.wantReason)
				}
			}
		})
	}
}

func TestValidateAllFolderNames(t *testing.T) {
	tests := []struct {
		name       string
		rules      []EnfolderRule
		wantErrors int
	}{
		{
			name: "all valid",
			rules: []EnfolderRule{
				{"Documents", []string{"doc"}},
				{"Pictures", []string{"pic"}},
				{"Videos", []string{"vid"}},
			},
			wantErrors: 0,
		},
		{
			name: "some invalid",
			rules: []EnfolderRule{
				{"Documents", []string{"doc"}},
				{"CON", []string{"con"}},        // reserved
				{"folder:name", []string{"x"}},  // invalid char
				{"folder.", []string{"y"}},      // ends with period
			},
			wantErrors: 3,
		},
		{
			name: "all invalid",
			rules: []EnfolderRule{
				{"", []string{"a"}},           // empty
				{".", []string{"b"}},          // special
				{"folder/name", []string{"c"}}, // invalid char
			},
			wantErrors: 3,
		},
		{
			name:       "empty rules",
			rules:      []EnfolderRule{},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateAllFolderNames(tt.rules)
			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateAllFolderNames() got %d errors, want %d", len(errors), tt.wantErrors)
				for _, e := range errors {
					t.Logf("  Error: %s - %s", e.FolderName, e.Reason)
				}
			}
		})
	}
}
