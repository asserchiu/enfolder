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
		wantError  bool
		wantReason string // substring to check in reason
	}{
		// Valid names
		{"valid simple name", "Documents", false, ""},
		{"valid with spaces", "My Documents", false, ""},
		{"valid with underscore", "my_folder", false, ""},
		{"valid with dash", "my-folder", false, ""},
		{"valid with numbers", "folder123", false, ""},
		{"valid with parentheses", "folder(1)", false, ""},
		{"valid with brackets", "folder[1]", false, ""},
		{"valid with unicode", "資料夾", false, ""},
		{"valid starts with dot", ".hidden", false, ""},

		// Invalid: empty
		{"empty name", "", true, "empty"},

		// Invalid: too long
		{"too long", strings.Repeat("a", 256), true, "255"},

		// Invalid: Windows reserved characters
		{"contains <", "folder<name", true, "invalid characters"},
		{"contains >", "folder>name", true, "invalid characters"},
		{"contains :", "folder:name", true, "invalid characters"},
		{"contains \"", "folder\"name", true, "invalid characters"},
		{"contains /", "folder/name", true, "invalid characters"},
		{"contains \\", "folder\\name", true, "invalid characters"},
		{"contains |", "folder|name", true, "invalid characters"},
		{"contains ?", "folder?name", true, "invalid characters"},
		{"contains *", "folder*name", true, "invalid characters"},

		// Invalid: ends with space or period
		{"ends with space", "folder ", true, "end with space or period"},
		{"ends with period", "folder.", true, "end with space or period"},
		{"ends with multiple periods", "folder...", true, "end with space or period"},

		// Invalid: reserved names (Windows)
		{"reserved CON", "CON", true, "reserved name"},
		{"reserved PRN", "PRN", true, "reserved name"},
		{"reserved AUX", "AUX", true, "reserved name"},
		{"reserved NUL", "NUL", true, "reserved name"},
		{"reserved COM1", "COM1", true, "reserved name"},
		{"reserved COM9", "COM9", true, "reserved name"},
		{"reserved LPT1", "LPT1", true, "reserved name"},
		{"reserved LPT9", "LPT9", true, "reserved name"},
		{"reserved con lowercase", "con", true, "reserved name"},
		{"reserved CON.txt", "CON.txt", true, "reserved name"},
		{"reserved aux.log", "aux.log", true, "reserved name"},

		// Invalid: Unix special names
		{"dot", ".", true, "cannot be '.' or '..'"},
		{"double dot", "..", true, "cannot be '.' or '..'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFolderName(tt.folderName)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateFolderName(%q) expected error, got nil", tt.folderName)
					return
				}

				// Type assert to ValidationError
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("ValidateFolderName(%q) returned non-ValidationError: %T", tt.folderName, err)
					return
				}

				// Check folder name is set correctly
				if ve.FolderName != tt.folderName {
					t.Errorf("ValidationError.FolderName = %q, want %q", ve.FolderName, tt.folderName)
				}

				// Check reason contains expected substring
				if tt.wantReason != "" {
					if !strings.Contains(strings.ToLower(ve.Reason), strings.ToLower(tt.wantReason)) {
						t.Errorf("ValidationError.Reason = %q, want to contain %q", ve.Reason, tt.wantReason)
					}
				}

				// Check Error() method
				errMsg := err.Error()
				if !strings.Contains(errMsg, tt.folderName) {
					t.Errorf("ValidationError.Error() = %q, want to contain folder name %q", errMsg, tt.folderName)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateFolderName(%q) expected no error, got %v", tt.folderName, err)
				}
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name         string
		ve           ValidationError
		wantContains []string
	}{
		{
			name: "basic error",
			ve: ValidationError{
				FolderName: "test",
				Reason:     "invalid",
			},
			wantContains: []string{"test", "invalid"},
		},
		{
			name: "reserved name error",
			ve: ValidationError{
				FolderName: "CON",
				Reason:     "folder name 'CON' is a reserved name on Windows",
			},
			wantContains: []string{"CON", "reserved name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.ve.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(errMsg, want) {
					t.Errorf("Error() = %q, want to contain %q", errMsg, want)
				}
			}
		})
	}
}
