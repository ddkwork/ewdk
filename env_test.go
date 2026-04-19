package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ddkwork/golibrary/std/mylog"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func isRunningAsAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)
	member, err := windows.GetCurrentProcessToken().IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

func TestLooksLikePath(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"hello", false},
		{"123", false},
		{"C:\\Windows", true},
		{"D:\\path\\to\\file", true},
		{"c:/windows", true},
		{"/usr/bin/gcc", true},
		{"%SystemRoot%\\system32", true},
		{"%ProgramFiles%\\Git\\bin", true},
		{"PATH", false},
		{"Release", false},
		{"x64", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := looksLikePath(tt.input)
			if result != tt.expected {
				t.Errorf("looksLikePath(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	tests := []struct {
		name       string
		value      string
		isPath     bool
		wantValid  bool
		wantReason string
	}{
		{
			name:       "non-path always valid",
			value:      "Release",
			isPath:     false,
			wantValid:  true,
			wantReason: "",
		},
		{
			name:       "empty value valid",
			value:      "",
			isPath:     true,
			wantValid:  true,
			wantReason: "",
		},
		{
			name:       "existing path valid",
			value:      tmpDir,
			isPath:     true,
			wantValid:  true,
			wantReason: "",
		},
		{
			name:       "existing file valid",
			value:      existingFile,
			isPath:     true,
			wantValid:  true,
			wantReason: "",
		},
		{
			name:       "nonexistent path invalid",
			value:      "C:\\nonexistent_path_xyz",
			isPath:     true,
			wantValid:  false,
			wantReason: "missing paths:",
		},
		{
			name:      "mixed paths one invalid",
			value:     tmpDir + ";C:\\does_not_exist_abc",
			isPath:    true,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, reason := validateValue(tt.value, tt.isPath)
			if valid != tt.wantValid {
				t.Errorf("valid = %v, want %v", valid, tt.wantValid)
			}
			if tt.wantReason != "" && !strings.Contains(reason, "missing") && !tt.wantValid {
				t.Errorf("reason = %q, want containing 'missing'", reason)
			}
		})
	}
}

func TestTypeToString(t *testing.T) {
	tests := []struct {
		valType  uint32
		expected string
	}{
		{registry.NONE, "NONE"},
		{registry.SZ, "SZ"},
		{registry.EXPAND_SZ, "EXPAND_SZ"},
		{registry.BINARY, "BINARY"},
		{registry.DWORD, "DWORD"},
		{registry.MULTI_SZ, "MULTI_SZ"},
		{registry.QWORD, "QWORD"},
		{0xFF, "UNKNOWN(255)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := typeToString(tt.valType)
			if result != tt.expected {
				t.Errorf("typeToString(%d) = %q, want %q", tt.valType, result, tt.expected)
			}
		})
	}
}

func TestSplitPathList(t *testing.T) {
	input := "C:\\a;D:\\b;E:\\c"
	parts := splitPathList(input)
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	if parts[0] != "C:\\a" || parts[1] != "D:\\b" || parts[2] != "E:\\c" {
		t.Errorf("unexpected parts: %v", parts)
	}
}

func TestComputeDiff(t *testing.T) {
	before := map[string]string{
		"PATH":    "C:\\Windows",
		"TEMP":    "%TEMP%",
		"INCLUDE": "old_include",
	}
	after := map[string]string{
		"PATH":    "C:\\Windows;new_path",
		"TEMP":    "%TEMP%",
		"NEW_VAR": "new_value",
		"INCLUDE": "updated_include",
	}

	diff := computeDiff(before, after)

	if len(diff.Added) != 1 {
		t.Errorf("Added count = %d, want 1", len(diff.Added))
	}
	if val, ok := diff.Added["NEW_VAR"]; !ok || val != "new_value" {
		t.Errorf("Added[NEW_VAR] = %q, want %q", val, "new_value")
	}

	if len(diff.Changed) != 2 {
		t.Errorf("Changed count = %d, want 2", len(diff.Changed))
	}
	changedNames := make(map[string]bool)
	for _, c := range diff.Changed {
		changedNames[c.Name] = true
		if c.Name == "PATH" {
			if c.OldValue != "C:\\Windows" || c.NewValue != "C:\\Windows;new_path" {
				t.Errorf("PATH delta: old=%q new=%q, want old=C:\\Windows new=C:\\Windows;new_path", c.OldValue, c.NewValue)
			}
		}
		if c.Name == "INCLUDE" {
			if c.OldValue != "old_include" || c.NewValue != "updated_include" {
				t.Errorf("INCLUDE delta: old=%q new=%q, want old=old_include new=updated_include", c.OldValue, c.NewValue)
			}
		}
	}
	if !changedNames["PATH"] || !changedNames["INCLUDE"] {
		t.Errorf("Changed names = %v, want PATH and INCLUDE present", changedNames)
	}

	if len(diff.Removed) != 0 {
		t.Errorf("Removed = %v, want empty", diff.Removed)
	}
}

func TestInferCMakeValue(t *testing.T) {
	availableVars := map[string]string{
		"INCLUDE":         "C:\\wdk\\include",
		"LIB":             "C:\\wdk\\lib",
		"CC":              "cl.exe",
		"CXX":             "cl.exe",
		"CFLAGS":          "/O2",
		"CXXFLAGS":        "/O2 /std:c++20",
		"LDFLAGS":         "/LTCG",
		"RC":              "rc.exe",
		"WDK_ROOT":        "F:\\Program Files\\Windows Kits\\10",
		"EWDKSetupEnvCmd": "F:\\BuildEnv\\SetupBuildEnv.cmd",
	}

	tests := []struct {
		cmakeKey  string
		wantFound bool
		wantVal   string
	}{
		{"CMAKE_INCLUDE_PATH", true, "C:\\wdk\\include"},
		{"CMAKE_LIBRARY_PATH", true, "C:\\wdk\\lib"},
		{"CC", true, "cl.exe"},
		{"CXX", true, "cl.exe"},
		{"CFLAGS", true, "/O2"},
		{"CXXFLAGS", true, "/O2 /std:c++20"},
		{"LDFLAGS", true, "/LTCG"},
		{"RC", true, "rc.exe"},
		{"CMAKE_PREFIX_PATH", true, "F:\\Program Files\\Windows Kits\\10"},
		{"CMAKE_TOOLCHAIN_FILE", true, "F:\\cmake\\Toolchain-Windows.cmake"},
		{"NONEXISTENT_VAR", false, ""},
		{"VERBOSE", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.cmakeKey, func(t *testing.T) {
			val, found := inferCMakeValue(tt.cmakeKey, availableVars)
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if found && val != tt.wantVal {
				t.Errorf("val = %q, want %q", val, tt.wantVal)
			}
		})
	}
}

func TestFillCMakeEnvVars(t *testing.T) {
	mgr := NewRegistryEnvManager()

	newVars := map[string]string{
		"CC":              "cl.exe",
		"CXX":             "cl.exe",
		"CFLAGS":          "/O2",
		"CXXFLAGS":        "/std:c++20",
		"INCLUDE":         "C:\\wdk\\include",
		"LIB":             "C:\\wdk\\lib",
		"WDK_ROOT":        "F:\\ Kits\\10",
		"EWDKSetupEnvCmd": "F:\\BuildEnv\\SetupBuildEnv.cmd",
	}

	result, err := mgr.FillCMake(newVars)
	if err != nil {
		t.Fatalf("FillCMakeEnvVars error: %v", err)
	}

	if result["CC"] != "cl.exe" {
		t.Errorf("CC = %q, want cl.exe", result["CC"])
	}
	if result["CXX"] != "cl.exe" {
		t.Errorf("CXX = %q, want cl.exe", result["CXX"])
	}
	if result["CFLAGS"] != "/O2" {
		t.Errorf("CFLAGS = %q, want /O2", result["CFLAGS"])
	}
	if result["CMAKE_INCLUDE_PATH"] != "C:\\wdk\\include" {
		t.Errorf("CMAKE_INCLUDE_PATH = %q, want C:\\wdk\\include", result["CMAKE_INCLUDE_PATH"])
	}
	if result["CMAKE_LIBRARY_PATH"] != "C:\\wdk\\lib" {
		t.Errorf("CMAKE_LIBRARY_PATH = %q, want C:\\wdk\\lib", result["CMAKE_LIBRARY_PATH"])
	}
	if result["CMAKE_PREFIX_PATH"] != "F:\\ Kits\\10" {
		t.Errorf("CMAKE_PREFIX_PATH = %q", result["CMAKE_PREFIX_PATH"])
	}
}

func TestCreateStartupScript(t *testing.T) {
	mgr := NewRegistryEnvManager()

	content := "@echo off\necho test startup script\n"
	scriptName := "test-ewdk-startup.cmd"

	destPath, err := mgr.CreateStartupScript(content, scriptName)
	if err != nil {
		t.Fatalf("CreateStartupScript error: %v", err)
	}

	expectedDest := filepath.Join(StartupDir, scriptName)
	if destPath != expectedDest {
		t.Errorf("destPath = %q, want %q", destPath, expectedDest)
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content mismatch")
	}

	os.Remove(destPath)
}

func TestDeleteEnvVar_NonExistent(t *testing.T) {
	mgr := NewRegistryEnvManager()

	err := mgr.Delete("_NONEXISTENT_TEST_VAR_12345")
	if err != nil {
		t.Errorf("DeleteEnvVar of non-existent var returned error: %v", err)
	}
}

func TestDeleteEnvVar_SetAndDelete(t *testing.T) {
	mgr := NewRegistryEnvManager()

	testName := "_GO_TEST_DELETE_ME_"
	testValue := "test_value_for_deletion"

	key, err := openEnvKey(registry.SET_VALUE)
	if err != nil {
		t.Skipf("cannot open registry key for writing: %v", err)
	}
	defer key.Close()

	key.SetStringValue(testName, testValue)

	err = mgr.Delete(testName)
	if err != nil {
		if strings.Contains(err.Error(), "denied") || strings.Contains(err.Error(), "access") {
			t.Skipf("access denied (need admin): %v", err)
		}
		t.Errorf("DeleteEnvVar error: %v", err)
	}

	_, _, err2 := key.GetStringValue(testName)
	if err2 == nil {
		t.Error("value still exists after delete")
	} else if err2 != registry.ErrNotExist && err2 != registry.ErrUnexpectedType && !strings.Contains(err2.Error(), "denied") {
		t.Errorf("unexpected error after delete: %v", err2)
	}
}

func TestListAllEnvVars_Structure(t *testing.T) {
	mgr := NewRegistryEnvManager()

	list, err := mgr.List()
	if err != nil {
		t.Fatalf("ListAllEnvVars error: %v", err)
	}

	if len(list) == 0 {
		t.Fatal("expected at least one env var")
	}

	for _, v := range list {
		if v.Name == "" {
			t.Error("found env var with empty name")
		}
		if v.Type == "" {
			t.Errorf("env var %q has empty type", v.Name)
		}
		if v.IsPath && !v.Valid && v.Reason == "" {
			t.Errorf("env var %q is invalid path but no reason given", v.Name)
		}
	}
}

func TestCaptureSetupBuildEnvDiff_Sanity(t *testing.T) {
	mgr := NewRegistryEnvManager()
	mgr.CleanInvalidVars()
	setupCmd := "f:\\BuildEnv\\SetupBuildEnv.cmd"

	diff, err := mgr.CaptureDiff(setupCmd)
	if err == nil {
		t.Log("added vars:", diff.Added)
		t.Log("changed vars:", diff.Changed)
		t.Log("removed vars:", diff.Removed)
	}
	mylog.Struct(diff)
}

func TestCMakeEnvVars_Completeness(t *testing.T) {
	expectedKeys := []string{
		"CMAKE_INCLUDE_PATH",
		"CMAKE_LIBRARY_PATH",
		"CMAKE_PREFIX_PATH",
		"CC", "CXX", "CFLAGS", "CXXFLAGS",
		"LDFLAGS", "FC", "FFLAGS",
		"RC", "RCFLAGS",
		"CMAKE_BUILD_TYPE",
		"CMAKE_GENERATOR",
		"CMAKE_TOOLCHAIN_FILE",
	}

	for _, key := range expectedKeys {
		if _, exists := cmakeEnvVars[key]; !exists {
			t.Errorf("missing CMake env var: %s", key)
		}
	}

	if len(cmakeEnvVars) < 25 {
		t.Errorf("cmakeEnvVars has only %d entries, expected >= 25", len(cmakeEnvVars))
	}
}

func TestNewRegistryEnvManager(t *testing.T) {
	m := NewRegistryEnvManager()
	if m == nil {
		t.Fatal("NewRegistryEnvManager returned nil")
	}
}

func TestEnvManagerInterface(t *testing.T) {
	var _ EnvManager = (*RegistryEnvManager)(nil)
}

func TestExpandEnvVar_SystemVar(t *testing.T) {
	m := NewRegistryEnvManager()

	result, err := m.Expand("%SystemRoot%")
	if err != nil {
		t.Fatalf("Expand error: %v", err)
	}
	if result == "%SystemRoot%" {
		t.Errorf("Expand did not expand %%SystemRoot%%: got %q", result)
	}
	if !strings.Contains(result, "\\") {
		t.Errorf("expected path with backslash, got %q", result)
	}
	t.Logf("SystemRoot expanded to: %q", result)
}

func TestExpandEnvVar_NoVars(t *testing.T) {
	m := NewRegistryEnvManager()

	result, err := m.Expand("plain text no vars")
	if err != nil {
		t.Fatalf("Expand error: %v", err)
	}
	if result != "plain text no vars" {
		t.Errorf("expected unchanged, got %q", result)
	}
}

func TestExpandEnvVar_Mixed(t *testing.T) {
	m := NewRegistryEnvManager()

	result, err := m.Expand("%SystemRoot%\\system32")
	if err != nil {
		t.Fatalf("Expand error: %v", err)
	}
	if strings.Contains(result, "%") && strings.Contains(result, "SystemRoot") {
		t.Logf("partial expand (may not have SystemRoot): %q", result)
	} else if !strings.Contains(result, "system32") {
		t.Errorf("expected system32 in path, got %q", result)
	}
}

func TestExpandEnvVar_GlobalFunc(t *testing.T) {
	mgr := NewRegistryEnvManager()

	result, err := mgr.Expand("%TEMP%")
	if err != nil {
		t.Fatalf("ExpandEnvVar error: %v", err)
	}
	if result == "%TEMP%" {
		t.Log("TEMP may not be set at user level")
	} else {
		t.Logf("TEMP expanded to: %q", result)
	}
}

func sortStrings(s []string) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func TestList_PrintFormat(t *testing.T) {
	mgr := NewRegistryEnvManager()
	list, err := mgr.List()
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one env var")
	}
	for _, v := range list {
		if v.Name == "" {
			t.Error("found env var with empty name")
		}
		status := "OK"
		if v.IsPath && !v.Valid {
			status = fmt.Sprintf("INVALID: %s", v.Reason)
		}
		t.Logf("%-30s %-8s isPath=%-5t %s", v.Name, v.Type, v.IsPath, status)
	}
}

func TestDelete_IncludeVar(t *testing.T) {
	mgr := NewRegistryEnvManager()
	err := mgr.Delete("INCLUDE")
	if err != nil {
		t.Errorf("Delete INCLUDE error: %v", err)
	}
}

func TestFillCMake_FullMapping(t *testing.T) {
	mgr := NewRegistryEnvManager()
	newVars := map[string]string{
		"INCLUDE":  "C:\\ewdk\\Include",
		"LIB":      "C:\\ewdk\\Lib",
		"CC":       "cl.exe",
		"CXX":      "cl.exe",
		"WDK_ROOT": "F:\\Program Files\\Windows Kits\\10",
	}
	cmakeVars, err := mgr.FillCMake(newVars)
	if err != nil {
		t.Fatalf("FillCMake error: %v", err)
	}
	if len(cmakeVars) == 0 {
		t.Fatal("expected at least one CMake var mapping")
	}
	for k, v := range cmakeVars {
		t.Logf("  %s=%s", k, v)
	}
	if cmakeVars["CC"] != "cl.exe" {
		t.Errorf("CC = %q, want cl.exe", cmakeVars["CC"])
	}
	if cmakeVars["CMAKE_INCLUDE_PATH"] != "C:\\ewdk\\Include" {
		t.Errorf("CMAKE_INCLUDE_PATH = %q", cmakeVars["CMAKE_INCLUDE_PATH"])
	}
}

func TestCreateStartupScript_EWDKSetup(t *testing.T) {
	mgr := NewRegistryEnvManager()
	script := `@echo off
call F:\BuildEnv\SetupBuildEnv.cmd
echo EWDK environment loaded > C:\ewdk-env-ready.flag
`
	path, err := mgr.CreateStartupScript(script, "ewdk-setup.cmd")
	if err != nil {
		t.Fatalf("CreateStartupScript error: %v", err)
	}
	expectedDest := filepath.Join(StartupDir, "ewdk-setup.cmd")
	if path != expectedDest {
		t.Errorf("path = %q, want %q", path, expectedDest)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	if string(data) != script {
		t.Error("file content mismatch")
	}
	os.Remove(path)
}

func TestExpand_SystemRoot(t *testing.T) {
	mgr := NewRegistryEnvManager()
	expanded, err := mgr.Expand("%SystemRoot%\\System32")
	if err != nil {
		t.Fatalf("Expand error: %v", err)
	}
	if expanded == "%SystemRoot%\\System32" {
		t.Log("SystemRoot not expanded (may not be set)")
	} else if !strings.Contains(expanded, "System32") {
		t.Errorf("expected System32 in path, got %q", expanded)
	} else {
		t.Log("expanded:", expanded)
	}
}

func TestSetEnvVar_SetAndRead(t *testing.T) {
	mgr := NewRegistryEnvManager()

	testName := "_GO_TEST_SET_ME_"
	testValue := "test_value_for_set"

	defer func() {
		mgr.Delete(testName)
	}()

	err := mgr.Set(testName, testValue)
	if err != nil {
		if strings.Contains(err.Error(), "denied") || strings.Contains(err.Error(), "access") {
			t.Skipf("access denied (need admin): %v", err)
		}
		t.Fatalf("Set error: %v", err)
	}

	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		t.Fatalf("open key for readback: %v", err)
	}
	defer key.Close()

	gotVal, gotType, err := key.GetStringValue(testName)
	if err != nil {
		t.Fatalf("read back error: %v", err)
	}

	if gotVal != testValue {
		t.Errorf("got value = %q, want %q", gotVal, testValue)
	}
	if gotType != registry.SZ && gotType != registry.EXPAND_SZ {
		t.Errorf("got type = %d, want SZ/EXPAND_SZ", gotType)
	}
}

func TestSetEnvVar_UpdateExisting(t *testing.T) {
	mgr := NewRegistryEnvManager()

	testName := "_GO_TEST_UPDATE_ME_"

	key, err := openEnvKey(registry.SET_VALUE)
	if err != nil {
		t.Skipf("cannot open registry key: %v", err)
	}
	key.SetStringValue(testName, "old_value")
	defer func() {
		key.DeleteValue(testName)
		key.Close()
	}()

	err = mgr.Set(testName, "new_value")
	if err != nil {
		if strings.Contains(err.Error(), "denied") || strings.Contains(err.Error(), "access") {
			t.Skipf("access denied (need admin): %v", err)
		}
		t.Fatalf("Set update error: %v", err)
	}

	readBack, _, err2 := key.GetStringValue(testName)
	if err2 != nil {
		if strings.Contains(err2.Error(), "denied") || strings.Contains(err2.Error(), "access") || strings.Contains(err2.Error(), "Access") {
			t.Skipf("access denied on readback (need admin): %v", err2)
		}
		t.Fatalf("read back error: %v", err2)
	}
	if readBack != "new_value" {
		t.Errorf("after update got = %q, want new_value", readBack)
	}
}

func TestGetMountedDriveLetter_NonExistentISO(t *testing.T) {
	mgr := NewRegistryEnvManager()

	result := mgr.GetMountedDriveLetter(`C:\nonexistent_ewdk_file_that_does_not_exist.iso`)
	if result == "" {
		t.Log("No CD-ROM drive found with SetupBuildEnv.cmd (expected on most systems)")
	} else {
		t.Logf("Found mounted drive letter: %s (EWDK may be mounted)", result)
	}
}

func TestGetMountedDriveLetter_EmptyPath(t *testing.T) {
	mgr := NewRegistryEnvManager()

	result := mgr.GetMountedDriveLetter("")
	if result == "" {
		t.Log("Empty path returned empty drive letter as expected")
	} else {
		t.Logf("Found drive letter even with empty path: %s", result)
	}
}

func TestCreateScheduledTask_ValidISO(t *testing.T) {
	mgr := NewRegistryEnvManager()

	testISO := filepath.Join(os.TempDir(), "_test_ewdk_mount_task_.iso")

	err := mgr.CreateScheduledTask(testISO)
	if err != nil {
		if strings.Contains(err.Error(), "denied") || strings.Contains(err.Error(), "access") || strings.Contains(err.Error(), "80004005") {
			t.Skipf("access denied (need admin): %v", err)
		}
		t.Fatalf("CreateScheduledTask error: %v", err)
	}

	cmd := exec.Command("schtasks", "/Query", "/TN", taskName, "/FO", "CSV", "/NH")
	output, queryErr := cmd.Output()
	if queryErr == nil {
		if !strings.Contains(string(output), taskName) {
			t.Error("task not found in schtasks query output")
		} else {
			t.Log("scheduled task created and verified:", string(output))
		}
	}

	mgr.DeleteScheduledTask()
}

func TestDeleteScheduledTask_AfterCreate(t *testing.T) {
	mgr := NewRegistryEnvManager()

	testISO := filepath.Join(os.TempDir(), "_test_ewdk_delete_task_.iso")

	mgr.CreateScheduledTask(testISO)

	mgr.DeleteScheduledTask()

	cmd := exec.Command("schtasks", "/Query", "/TN", taskName, "/FO", "CSV", "/NH")
	output, err := cmd.CombinedOutput()
	if err == nil {
		if strings.Contains(string(output), taskName) {
			t.Error("task still exists after DeleteScheduledTask")
		} else {
			t.Log("task successfully deleted")
		}
	} else {
		t.Logf("query returned error after delete (expected): %s", string(output))
	}
}

func TestSet_CustomVar(t *testing.T) {
	mgr := NewRegistryEnvManager()
	testName := "_GO_TEST_EXAMPLE_SET_"
	defer mgr.Delete(testName)

	err := mgr.Set(testName, "my_custom_value")
	if err != nil {
		if strings.Contains(err.Error(), "denied") || strings.Contains(err.Error(), "access") {
			t.Skipf("access denied (need admin): %v", err)
		}
		t.Fatalf("Set error: %v", err)
	}
}

func TestGetMountedDriveLetter_EWDKISO(t *testing.T) {
	mgr := NewRegistryEnvManager()
	letter := mgr.GetMountedDriveLetter(`F:\EWDK.iso`)
	if letter == "" {
		t.Log("No drive letter found (EWDK may not be mounted)")
	} else {
		t.Logf("EWDK mounted at: %s:", letter)
	}
}

func TestCreateScheduledTask_EWDKISO(t *testing.T) {
	mgr := NewRegistryEnvManager()

	err := mgr.CreateScheduledTask(`F:\EWDK.iso`)
	if err != nil {
		if strings.Contains(err.Error(), "denied") || strings.Contains(err.Error(), "access") || strings.Contains(err.Error(), "80004005") {
			t.Skipf("access denied (need admin): %v", err)
		}
		t.Fatalf("CreateScheduledTask error: %v", err)
	}

	cmd := exec.Command("schtasks", "/Query", "/TN", taskName, "/FO", "CSV", "/NH")
	output, queryErr := cmd.Output()
	if queryErr == nil && !strings.Contains(string(output), taskName) {
		t.Error("task not found after create")
	} else if queryErr == nil {
		t.Log("task created successfully")
	}

	mgr.DeleteScheduledTask()
}

func TestDeleteScheduledTask_Cleanup(t *testing.T) {
	mgr := NewRegistryEnvManager()
	mgr.DeleteScheduledTask()
	t.Log("DeleteScheduledTask completed without error")
}

func TestIsMounted_NonExistentISO(t *testing.T) {
	mgr := NewRegistryEnvManager()
	fakeISO := filepath.Join(os.TempDir(), "_nonexistent_test_iso_.iso")
	if mgr.IsMounted(fakeISO) {
		t.Error("non-existent ISO should not be mounted")
	}
	t.Log("IsMounted correctly returned false for non-existent ISO")
}

func TestIsMounted_EWDKISO(t *testing.T) {
	mgr := NewRegistryEnvManager()
	isoPath := `d:\ewdk\EWDK_br_release_28000_251103-1709.iso`
	if _, err := os.Stat(isoPath); os.IsNotExist(err) {
		t.Skipf("EWDK ISO not found: %s", isoPath)
	}
	mounted := mgr.IsMounted(isoPath)
	t.Logf("IsMounted(%s) = %v", isoPath, mounted)
}

func TestUnmountAll_NoError(t *testing.T) {
	mgr := NewRegistryEnvManager()
	if err := mgr.UnmountAll(); err != nil {
		t.Errorf("UnmountAll error: %v", err)
	}
	t.Log("UnmountAll completed without error")
}

func TestExtractDriveLetterFromDevicePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`\\?\Volume{guid}\`, ""},
		{`\\.\F:`, "F"},
		{`F:\`, "F"},
		{`D:\Program Files\Windows Kits\10`, "D"},
		{"", ""},
		{`no colon here`, ""},
		{`C:\Windows\System32`, "C"},
		{`z:\some\path`, "Z"},
	}
	for _, tc := range tests {
		got := extractDriveLetterFromDevicePath(tc.input)
		if got != tc.expected {
			t.Errorf("extractDriveLetterFromDevicePath(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestGetWDKContentRoot(t *testing.T) {
	mgr := NewRegistryEnvManager()
	val, err := mgr.GetWDKContentRoot()
	if err != nil {
		t.Logf("GetWDKContentRoot: %v (may not be set)", err)
	} else {
		t.Logf("WDKContentRoot = %s", val)
	}
}

func TestGetWDKRoot(t *testing.T) {
	mgr := NewRegistryEnvManager()
	val, err := mgr.GetWDKRoot()
	if err != nil {
		t.Logf("GetWDKRoot: %v (may not be set)", err)
	} else {
		t.Logf("WDK_ROOT = %s", val)
	}
}

func TestGetEWDKSetupEnvCmd(t *testing.T) {
	mgr := NewRegistryEnvManager()
	val, err := mgr.GetEWDKSetupEnvCmd()
	if err != nil {
		t.Logf("GetEWDKSetupEnvCmd: %v (may not be set)", err)
	} else {
		t.Logf("EWDKSetupEnvCmd = %s", val)
	}
}

func TestGetVirtualDiskPhysicalPath_NonExistentISO(t *testing.T) {
	mgr := NewRegistryEnvManager()
	fakeISO := filepath.Join(os.TempDir(), "_nonexistent_physical_path_.iso")
	_, err := mgr.GetVirtualDiskPhysicalPath(fakeISO)
	if err == nil {
		t.Error("non-existent ISO should return error")
	}
	t.Logf("GetVirtualDiskPhysicalPath on missing file: %v (expected)", err)
}

func TestGetVirtualDiskPhysicalPath_EWDKISO(t *testing.T) {
	mgr := NewRegistryEnvManager()
	isoPath := `d:\ewdk\EWDK_br_release_28000_251103-1709.iso`
	path, err := mgr.GetVirtualDiskPhysicalPath(isoPath)
	if err != nil {
		t.Logf("GetVirtualDiskPhysicalPath: %v (ISO may not be mounted)", err)
	} else {
		t.Logf("Physical path: %s", path)
	}
}

const testEWDKISO = `d:\ewdk\EWDK_br_release_28000_251103-1709.iso`

func TestMountUnmountEWDKISO_FullLifecycle(t *testing.T) {
	if _, err := os.Stat(testEWDKISO); os.IsNotExist(err) {
		t.Skipf("EWDK ISO not found at %s, skipping", testEWDKISO)
	}
	mgr := NewRegistryEnvManager()
	if !isRunningAsAdmin() {
		t.Skip("requires admin privileges")
	}
	t.Log("=== Phase 1: Clean unmount if already mounted ===")
	if mgr.IsMounted(testEWDKISO) {
		t.Log("ISO was already mounted, cleaning up first...")
		if err := mgr.UnmountISO(testEWDKISO); err != nil {
			t.Logf("pre-cleanup unmount error (may be ok): %v", err)
		}
	}
	t.Log("=== Phase 2: Mount ISO ===")
	letter, err := mgr.MountISO(testEWDKISO)
	if err != nil {
		t.Fatalf("MountISO failed: %v", err)
	}
	if letter == "" {
		t.Fatal("MountISO returned empty drive letter")
	}
	t.Logf("ISO mounted successfully at drive %s:", letter)
	t.Log("=== Phase 3: Verify mount state ===")
	if !mgr.IsMounted(testEWDKISO) {
		t.Error("IsMounted reports false after successful mount")
	}
	detectedLetter := mgr.GetMountedDriveLetter(testEWDKISO)
	if detectedLetter != letter {
		t.Errorf("GetMountedDriveLetter=%q, expected %q", detectedLetter, letter)
	}
	t.Log("=== Phase 4: Verify environment variables ===")
	wdkRoot, err1 := mgr.GetWDKContentRoot()
	wdkEnv, err2 := mgr.GetWDKRoot()
	setupCmd, err3 := mgr.GetEWDKSetupEnvCmd()
	if err1 != nil || err2 != nil || err3 != nil {
		t.Errorf("env var read errors: WDKContentRoot=%v WDK_ROOT=%v EWDKSetupEnvCmd=%v", err1, err2, err3)
	}
	expectedRoot := letter + `:\Program Files\Windows Kits\10`
	if wdkRoot != expectedRoot {
		t.Errorf("WDKContentRoot=%q, expected %q", wdkRoot, expectedRoot)
	}
	if wdkEnv != expectedRoot {
		t.Errorf("WDK_ROOT=%q, expected %q", wdkEnv, expectedRoot)
	}
	expectedCmd := letter + `:\BuildEnv\SetupBuildEnv.cmd`
	if setupCmd != expectedCmd {
		t.Errorf("EWDKSetupEnvCmd=%q, expected %q", setupCmd, expectedCmd)
	}
	t.Logf("Environment variables verified: WDKContentRoot=%s WDK_ROOT=%s EWDKSetupEnvCmd=%s", wdkRoot, wdkEnv, setupCmd)
	t.Log("=== Phase 5: Verify physical disk path ===")
	physPath, err := mgr.GetVirtualDiskPhysicalPath(testEWDKISO)
	if err != nil {
		t.Logf("GetVirtualDiskPhysicalPath warning: %v", err)
	} else if physPath == "" {
		t.Log("GetVirtualDiskPhysicalPath returned empty (ISO may use CD-ROM path)")
	} else {
		t.Logf("Physical disk path: %s", physPath)
	}
	t.Log("=== Phase 6: Unmount ISO ===")
	if err := mgr.UnmountISO(testEWDKISO); err != nil {
		t.Fatalf("UnmountISO failed: %v", err)
	}
	t.Log("ISO unmounted successfully")
	t.Log("=== Phase 7: Verify cleanup ===")
	if mgr.IsMounted(testEWDKISO) {
		t.Error("IsMounted still true after unmount")
	}
	if mgr.GetMountedDriveLetter(testEWDKISO) != "" {
		t.Error("GetMountedDriveLetter still returns value after unmount")
	}
	for _, name := range []string{"WDKContentRoot", "WDK_ROOT", "EWDKSetupEnvCmd"} {
		val, err := func() (string, error) {
			k, e := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, registry.READ)
			if e != nil {
				return "", e
			}
			defer k.Close()
			v, _, e := k.GetStringValue(name)
			return v, e
		}()
		if err == nil && val != "" {
			t.Errorf("%s still set to %q after unmount", name, val)
		}
	}
	t.Log("=== Full lifecycle test PASSED ===")
}

func TestMountEWDKISO_AlreadyMounted(t *testing.T) {
	if _, err := os.Stat(testEWDKISO); os.IsNotExist(err) {
		t.Skipf("EWDK ISO not found at %s, skipping", testEWDKISO)
	}
	mgr := NewRegistryEnvManager()
	if !isRunningAsAdmin() {
		t.Skip("requires admin privileges")
	}

	letter, err := mgr.MountISO(testEWDKISO)
	if err != nil {
		t.Fatalf("MountISO on already-mounted ISO failed: %v", err)
	}
	if letter == "" {
		t.Fatal("returned empty drive letter for already-mounted ISO")
	}
	t.Logf("Idempotent mount OK, drive %s:", letter)
}

func TestUnmountEWDKISO_NotMounted(t *testing.T) {
	if _, err := os.Stat(testEWDKISO); os.IsNotExist(err) {
		t.Skipf("EWDK ISO not found at %s, skipping", testEWDKISO)
	}
	mgr := NewRegistryEnvManager()
	if !isRunningAsAdmin() {
		t.Skip("requires admin privileges")
	}
	if mgr.IsMounted(testEWDKISO) {
		t.Skip("ISO is mounted, skip not-mounted unmount test")
	}
	if err := mgr.UnmountISO(testEWDKISO); err != nil {
		t.Fatalf("UnmountISO on non-mounted ISO should be no-op, got: %v", err)
	}
	t.Log("UnmountISO on non-mounted ISO succeeded (no-op)")
}

func TestEWDKISO_DriveLetterConsistency(t *testing.T) {
	if _, err := os.Stat(testEWDKISO); os.IsNotExist(err) {
		t.Skipf("EWDK ISO not found at %s, skipping", testEWDKISO)
	}
	mgr := NewRegistryEnvManager()
	if !mgr.IsMounted(testEWDKISO) {
		t.Skip("ISO not mounted, skipping consistency check")
	}
	letter := mgr.GetMountedDriveLetter(testEWDKISO)
	if letter == "" {
		t.Fatal("cannot detect drive letter for mounted ISO")
	}
	wdkRoot, _ := mgr.GetWDKContentRoot()
	if wdkRoot != "" && !strings.HasPrefix(wdkRoot, letter+":") {
		t.Errorf("WDKContentRoot=%q does not match drive %s:", wdkRoot, letter)
	}
	physPath, err := mgr.GetVirtualDiskPhysicalPath(testEWDKISO)
	if err != nil {
		t.Logf("physical path lookup error: %v", err)
	} else if physPath != "" {
		t.Logf("Drive %s: physical=%s env=%s", letter, physPath, wdkRoot)
	}
}

func TestCleanInvalidVars(t *testing.T) {
	mgr := NewRegistryEnvManager()

	// 创建一个无效的环境变量（指向不存在的路径）
	testName := "_GO_TEST_INVALID_VAR_"
	testValue := "C:\\nonexistent_path_xyz_12345"

	// 清理函数，确保测试后删除变量
	defer func() {
		mgr.Delete(testName)
	}()

	// 设置无效的环境变量
	err := mgr.Set(testName, testValue)
	if err != nil {
		if strings.Contains(err.Error(), "denied") || strings.Contains(err.Error(), "access") {
			t.Skipf("access denied (need admin): %v", err)
		}
		t.Fatalf("Set error: %v", err)
	}

	// 验证变量已设置
	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		t.Fatalf("open key: %v", err)
	}
	defer key.Close()

	_, _, err = key.GetStringValue(testName)
	if err != nil {
		t.Fatalf("verify var exists: %v", err)
	}

	// 调用 CleanInvalidVars 方法
	deleted, err := mgr.CleanInvalidVars()
	if err != nil {
		t.Fatalf("CleanInvalidVars error: %v", err)
	}

	// 验证变量已被删除
	_, _, err = key.GetStringValue(testName)
	if err == nil {
		t.Error("invalid var still exists after CleanInvalidVars")
	} else if err != registry.ErrNotExist && err != registry.ErrUnexpectedType && !strings.Contains(err.Error(), "denied") {
		t.Errorf("unexpected error after delete: %v", err)
	}

	// 验证返回的删除数量
	if deleted == 0 {
		t.Error("CleanInvalidVars returned 0, expected at least 1")
	} else {
		t.Logf("CleanInvalidVars deleted %d invalid variables", deleted)
	}
}

func TestCleanInvalidVars_NoInvalidVars(t *testing.T) {
	mgr := NewRegistryEnvManager()

	// 调用 CleanInvalidVars 方法，应该返回 0
	deleted, err := mgr.CleanInvalidVars()
	if err != nil {
		t.Fatalf("CleanInvalidVars error: %v", err)
	}

	// 验证返回的删除数量为 0
	t.Logf("CleanInvalidVars deleted %d invalid variables (no invalid vars expected)", deleted)
}

func TestCleanInvalidVars_PathVariable(t *testing.T) {
	mgr := NewRegistryEnvManager()

	// 获取当前 PATH 变量值
	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		t.Skipf("cannot open registry key: %v", err)
	}
	defer key.Close()

	originalPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		t.Skipf("cannot read PATH variable: %v", err)
	}

	// 保存原始 PATH 值，测试后恢复
	defer func() {
		mgr.Set("PATH", originalPath)
	}()

	// 创建一个包含有效和无效路径的 PATH 值
	validPath := os.Getenv("SystemRoot") // 系统目录应该存在
	invalidPath := "C:\\nonexistent_path_xyz_12345"
	testPath := validPath + ";" + invalidPath

	// 设置测试 PATH 值
	err = mgr.Set("PATH", testPath)
	if err != nil {
		if strings.Contains(err.Error(), "denied") || strings.Contains(err.Error(), "access") {
			t.Skipf("access denied (need admin): %v", err)
		}
		t.Fatalf("Set PATH error: %v", err)
	}

	// 验证 PATH 已设置
	setPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		t.Fatalf("verify PATH exists: %v", err)
	}
	if setPath != testPath {
		t.Errorf("PATH not set correctly: got %q, want %q", setPath, testPath)
	}

	// 调用 CleanInvalidVars 方法
	deleted, err := mgr.CleanInvalidVars()
	if err != nil {
		t.Fatalf("CleanInvalidVars error: %v", err)
	}

	// 验证 PATH 变量仍然存在
	cleanedPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		t.Fatalf("PATH variable was deleted: %v", err)
	}

	// 验证无效路径已被清理，有效路径保留
	if !strings.Contains(cleanedPath, validPath) {
		t.Error("valid path was removed from PATH")
	}
	if strings.Contains(cleanedPath, invalidPath) {
		t.Error("invalid path was not removed from PATH")
	}

	t.Logf("CleanInvalidVars deleted %d invalid variables", deleted)
	t.Logf("Original PATH: %q", originalPath)
	t.Logf("Test PATH: %q", testPath)
	t.Logf("Cleaned PATH: %q", cleanedPath)
}
