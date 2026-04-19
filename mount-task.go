package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ddkwork/golibrary/std/mylog"
	"golang.org/x/sys/windows/registry"
)

const targetArch = "x64"
const msvcArch = "amd64"

func main() {
	mgr := NewRegistryEnvManager()

	isoPath := resolveISOPath()
	if _, err := os.Stat(isoPath); os.IsNotExist(err) {
		fmt.Printf("Error: ISO not found: %s\n", isoPath)
		os.Exit(1)
	}
	fmt.Printf("ISO path: %s\n", isoPath)

	fmt.Println("\n=== Step 1: 清理无效环境变量 ===")
	count, err := mgr.CleanInvalidVars()
	if err != nil {
		fmt.Printf("CleanInvalidVars warning: %v\n", err)
	} else {
		fmt.Printf("Cleaned %d invalid environment variables\n", count)
	}

	fmt.Println("\n=== Step 2: 删除旧干扰变量 ===")
	mgr.Delete("INCLUDE")
	mgr.Delete("LIB")

	fmt.Println("\n=== Step 3: 挂载 EWDK ISO ===")
	driveLetter, err := mgr.MountISO(isoPath)
	if err != nil {
		fmt.Printf("MountISO error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("EWDK mounted at: %s:\n", driveLetter)

	setupEnvCmd, err := mgr.GetEWDKSetupEnvCmd()
	if err != nil {
		fmt.Printf("GetEWDKSetupEnvCmd error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("SetupBuildEnv.cmd: %s %s\n", setupEnvCmd, msvcArch)

	fmt.Println("\n=== Step 4: 执行 SetupBuildEnv.cmd amd64 并捕获环境变量差异 ===")
	diff, err := mgr.CaptureDiff(setupEnvCmd + " " + msvcArch)
	if err != nil {
		fmt.Printf("CaptureDiff error: %v\n", err)
		os.Exit(1)
	}
	mylog.Struct(diff)
	fmt.Printf("Added   : %d vars\n", len(diff.Added))
	for name, val := range diff.Added {
		fmt.Printf("  + %-30s = %s\n", name, truncate(val, 80))
	}
	fmt.Printf("Changed : %d vars\n", len(diff.Changed))
	for _, c := range diff.Changed {
		fmt.Printf("  ~ %-30s %s → %s\n", c.Name, truncate(c.OldValue, 50), truncate(c.NewValue, 50))
	}
	fmt.Printf("Removed : %d vars\n", len(diff.Removed))
	for name := range diff.Removed {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\n=== Step 5: 将差异写入系统环境变量（永久生效）===")
	setEnvVars(mgr, diff)

	fmt.Println("\n=== Step 5b: 补全 WDK Include/Lib 路径 ===")
	ensureWDKPaths(mgr, driveLetter)

	fmt.Println("\n=== Step 6: 填充 CMake 变量并写入 ===")
	cmakeVars, err := mgr.FillCMake(diff.Added)
	if err != nil {
		fmt.Printf("FillCMake warning: %v\n", err)
	} else {
		for name, value := range cmakeVars {
			if err := mgr.Set(name, value); err != nil {
				fmt.Printf("  [FAIL] CMake Set %s: %v\n", name, err)
			} else {
				fmt.Printf("  [OK]   CMake SET %-20s = %s\n", name, truncate(value, 60))
			}
		}
	}

	fmt.Println("\n=== Step 7: 强制设置 x64 编译器 ===")
	forceSetX64Compilers(mgr, diff.Added)

	ninjaDir, err := filepath.Abs(filepath.Dir("ninja.exe"))
	if err == nil {
		appendNinjaToPATH(mgr, ninjaDir)
	}

	fmt.Println("\n=== Step 8: 构建 ===")
	runBuild()

	fmt.Println("\n=== Step 9: 最终环境变量健康检查 ===")
	finalCheck(mgr)

	fmt.Println("\nDone!")
}

func setEnvVars(mgr EnvManager, diff *EnvDiff) {
	for name, value := range diff.Added {
		if err := mgr.Set(name, value); err != nil {
			fmt.Printf("  [FAIL] Set %s: %v\n", name, err)
		} else {
			fmt.Printf("  [OK]   SET %s\n", name)
		}
	}
	for _, delta := range diff.Changed {
		if err := mgr.Set(delta.Name, delta.NewValue); err != nil {
			fmt.Printf("  [FAIL] Set %s: %v\n", delta.Name, err)
		} else {
			fmt.Printf("  [OK]   UPD %s\n", delta.Name)
		}
	}
}

func forceSetX64Compilers(mgr EnvManager, addedVars map[string]string) {
	vcToolsDir := addedVars["VCToolsInstallDir"]
	if vcToolsDir == "" {
		fmt.Println("  [WARN] VCToolsInstallDir not found in diff, cannot force x64 compilers")
		return
	}

	clExe := filepath.Join(vcToolsDir, "bin", "Hostx64", "x64", "cl.exe")

	for _, v := range []struct {
		name  string
		value string
	}{
		{"CC", clExe},
		{"CXX", clExe},
		{"CMAKE_C_COMPILER", clExe},
		{"CMAKE_CXX_COMPILER", clExe},
	} {
		if err := mgr.Set(v.name, v.value); err != nil {
			fmt.Printf("  [FAIL] Force %s: %v\n", v.name, err)
		} else {
			fmt.Printf("  [OK]   Force %-20s = %s\n", v.name, v.value)
		}
	}

	fixPATHForX64(mgr, addedVars, clExe)
}

func fixPATHForX64(mgr EnvManager, addedVars map[string]string, x64ClExe string) {
	x64BinDir := filepath.Dir(x64ClExe)
	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	currentPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		return
	}

	if strings.Contains(currentPath, x64BinDir) && !strings.Contains(currentPath, "HostX86\\x86") {
		fmt.Println("  [SKIP] PATH already has x64 tools at front")
		return
	}

	parts := strings.Split(currentPath, ";")
	var cleaned []string
	var x64Inserted bool
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "HostX86\\x86") || strings.Contains(part, "HostX86/x86") {
			continue
		}
		if !x64Inserted && strings.Contains(part, "Hostx64\\x64") {
			cleaned = append(cleaned, part)
			x64Inserted = true
			continue
		}
		cleaned = append(cleaned, part)
	}

	if !x64Inserted {
		cleaned = append([]string{x64BinDir}, cleaned...)
	}

	newPath := strings.Join(cleaned, ";")
	if err := mgr.Set("PATH", newPath); err != nil {
		fmt.Printf("  [FAIL] Fix PATH for x64: %v\n", err)
	} else {
		fmt.Printf("  [OK]   PATH fixed: x64 tools prioritized, x86 removed\n")
	}
}

func resolveISOPath() string {
	githubWorkspace := os.Getenv("GITHUB_WORKSPACE")
	if githubWorkspace != "" {
		return filepath.Join(os.Getenv("TEMP"), "ewdk.iso")
	}
	return `d:\ewdk\EWDK_br_release_28000_251103-1709.iso`
}

func appendNinjaToPATH(mgr EnvManager, ninjaDir string) {
	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	currentPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		return
	}

	if strings.Contains(currentPath, ninjaDir) {
		fmt.Printf("  [SKIP] ninja.exe already in PATH: %s\n", ninjaDir)
		return
	}

	newPath := currentPath + ";" + ninjaDir
	if err := mgr.Set("PATH", newPath); err != nil {
		fmt.Printf("  [FAIL] Append ninja to PATH: %v\n", err)
	} else {
		fmt.Printf("  [OK]   PATH += %s\n", ninjaDir)
	}
}

func runBuild() {
	os.RemoveAll("build")

	freshEnv, err := captureCurrentEnv()
	if err != nil {
		fmt.Printf("WARN: cannot read fresh env from registry: %v\n", err)
		freshEnv = nil
	}
	envSlice := os.Environ()
	if freshEnv != nil {
		envSlice = make([]string, 0, len(freshEnv))
		for k, v := range freshEnv {
			envSlice = append(envSlice, k+"="+v)
		}
	}
	runCmd := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Env = envSlice
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = "."
		return cmd.Run()
	}

	if err := runCmd("cmake", "-B", "build", "-G", "Ninja", "-DCMAKE_BUILD_TYPE=Release", "."); err != nil {
		fmt.Printf("CMake configure error: %v\n", err)
		return
	}

	if err := runCmd("cmake", "--build", "build", "--config", "Release"); err != nil {
		fmt.Printf("Build error: %v\n", err)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

var protectedVars = map[string]bool{
	"path": true, "lib": true, "include": true,
	"homepath": true, "home": true, "userprofile": true,
	"systemroot": true, "windir": true, "comspec": true,
	"pathext": true, "os": true, "processor_architecture": true,
	"number_of_processors": true, "psmodulepath": true, "libpath": true,
}

var compoundPathVars = map[string]bool{
	"path": true, "lib": true, "include": true,
	"psmodulepath": true, "cmake_include_path": true,
	"cmake_library_path": true, "cmake_prefix_path": true,
	"libpath": true, "__vscmd_preinit_path": true,
	"safe_rm_allowed_path": true, "safe_rm_denied_path": true,
}

func finalCheck(mgr EnvManager) {
	allVars, err := mgr.List()
	if err != nil {
		fmt.Printf("  [WARN] List failed: %v\n", err)
		return
	}

	var invalidVars []EnvVar
	for _, v := range allVars {
		if !v.Valid && v.IsPath {
			invalidVars = append(invalidVars, v)
		}
	}

	fmt.Printf("  Total vars: %d\n", len(allVars))
	fmt.Printf("  Invalid (path not found): %d\n", len(invalidVars))

	if len(invalidVars) == 0 {
		fmt.Println("  [OK] All environment variables are valid")
		checkScheduledTasks(mgr)
		return
	}

	cleaned := 0
	fixed := 0
	for _, v := range invalidVars {
		if isCompoundPath(v.Name) {
			cleanedVal := cleanCompoundPath(v.Value)
			if cleanedVal != v.Value {
				if err := mgr.Set(v.Name, cleanedVal); err != nil {
					fmt.Printf("  [FAIL] Fix %s: %v\n", v.Name, err)
				} else {
					fmt.Printf("  [FIX]  %-30s cleaned invalid sub-paths\n", v.Name)
					fixed++
				}
			} else if !isProtected(v.Name) {
				if err := mgr.Delete(v.Name); err != nil {
					fmt.Printf("  [FAIL] Delete %s: %v\n", v.Name, err)
				} else {
					fmt.Printf("  [DEL]  %-30s (entire value invalid)\n", v.Name)
					cleaned++
				}
			} else {
				fmt.Printf("  [SKIP] %-30s (protected, cannot fix)\n", v.Name)
			}
		} else if !isProtected(v.Name) {
			if err := mgr.Delete(v.Name); err != nil {
				fmt.Printf("  [FAIL] Delete %s: %v\n", v.Name, err)
			} else {
				fmt.Printf("  [DEL]  %-30s = %s\n", v.Name, truncate(v.Value, 60))
				cleaned++
			}
		} else {
			fmt.Printf("  [SKIP] %-30s (protected)\n", v.Name)
		}
	}

	fmt.Printf("\n  Fixed: %d compound paths | Deleted: %d invalid vars\n", fixed, cleaned)

	checkScheduledTasks(mgr)
}

func ensureWDKPaths(mgr EnvManager, driveLetter string) {
	wdkRoot := fmt.Sprintf("%s:\\Program Files\\Windows Kits\\10", driveLetter)
	includeDir := filepath.Join(wdkRoot, "Include")
	if _, err := os.Stat(includeDir); os.IsNotExist(err) {
		fmt.Printf("  [FAIL] WDK Include dir not found at %s (ISO may not be mounted at %s:)\n", includeDir, driveLetter)
		return
	}

	version := ""
	entries, _ := os.ReadDir(includeDir)
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "10.") {
			version = e.Name()
			break
		}
	}
	if version == "" {
		fmt.Printf("  [FAIL] No 10.x version dir found under %s\n", includeDir)
		return
	}
	fmt.Printf("  [INFO] WDK root: %s  version: %s\n", wdkRoot, version)

	includeBase := filepath.Join(wdkRoot, "Include", version)
	libBase := filepath.Join(wdkRoot, "Lib", version)

	validIncludes := make([]string, 0, 4)
	for _, d := range []string{"um", "shared", "ucrt", "km"} {
		p := filepath.Join(includeBase, d)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			fmt.Printf("  [WARN] skip %s (not found)\n", p)
		} else {
			validIncludes = append(validIncludes, p)
		}
	}
	validLibs := make([]string, 0, 2)
	for _, d := range []string{"um", "ucrt"} {
		p := filepath.Join(libBase, d, targetArch)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			fmt.Printf("  [WARN] skip %s (not found)\n", p)
		} else {
			validLibs = append(validLibs, p)
		}
	}

	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		fmt.Printf("  [FAIL] Open registry: %v\n", err)
		return
	}
	defer key.Close()

	rebuildPathVar(key, mgr, "INCLUDE", validIncludes, wdkRoot)
	rebuildPathVar(key, mgr, "LIB", validLibs, wdkRoot)
}

func mergeIntoPathVar(key registry.Key, mgr EnvManager, varName string, newParts ...string) {
	currentVal, _, err := key.GetStringValue(varName)
	if err != nil {
		currentVal = ""
	}

	existingMap := make(map[string]bool)
	for _, p := range strings.Split(currentVal, ";") {
		p = strings.TrimSpace(p)
		if p != "" {
			existingMap[strings.ToLower(p)] = true
		}
	}

	var added []string
	for _, part := range newParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		if !existingMap[lower] {
			added = append(added, part)
			existingMap[lower] = true
		}
	}

	if len(added) == 0 {
		fmt.Printf("  [OK]   %s already has all WDK paths\n", varName)
		return
	}

	var newVal string
	if currentVal == "" {
		newVal = strings.Join(added, ";")
	} else {
		newVal = currentVal + ";" + strings.Join(added, ";")
	}

	if err := mgr.Set(varName, newVal); err != nil {
		fmt.Printf("  [FAIL] Set %s: %v\n", varName, err)
	} else {
		for _, a := range added {
			fmt.Printf("  [OK]   %s += %s\n", varName, a)
		}
	}
}

func rebuildPathVar(key registry.Key, mgr EnvManager, varName string, wdkParts []string, wdkRoot string) {
	currentVal, _, err := key.GetStringValue(varName)
	if err != nil {
		currentVal = ""
	}

	kept := make([]string, 0)
	rootLower := strings.ToLower(strings.TrimRight(wdkRoot, `/\`))
	for _, p := range strings.Split(currentVal, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(strings.ToLower(p), rootLower) {
			continue
		}
		kept = append(kept, p)
	}

	var allParts []string
	allParts = append(allParts, kept...)
	allParts = append(allParts, wdkParts...)

	newVal := strings.Join(allParts, ";")
	if err := mgr.Set(varName, newVal); err != nil {
		fmt.Printf("  [FAIL] Set %s: %v\n", varName, err)
	} else {
		fmt.Printf("  [OK]   %s rebuilt (%d kept + %d WDK)\n", varName, len(kept), len(wdkParts))
	}
}

func isProtected(name string) bool    { return protectedVars[strings.ToLower(name)] }
func isCompoundPath(name string) bool { return compoundPathVars[strings.ToLower(name)] }

func checkScheduledTasks(mgr EnvManager) {
	fmt.Println("\n  Checking scheduled tasks...")
	cmd := exec.Command("schtasks", "/Query", "/TN", "EWDK_Mount", "/FO", "LIST")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("  [OK]   No EWDK_Mount task found (clean)")
	} else {
		fmt.Printf("  [INFO] EWDK_Mount task exists:\n%s\n", string(output))
		mgr.DeleteScheduledTask()
		fmt.Println("  [OK]   Removed stale EWDK_Mount task")
	}
}

func cleanCompoundPath(value string) string {
	parts := strings.Split(value, ";")
	var kept []string
	seen := make(map[string]bool)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		if seen[lower] {
			continue
		}
		if isRelativePath(part) || dirExists(part) {
			seen[lower] = true
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, ";")
}

func isRelativePath(p string) bool {
	if len(p) == 0 {
		return false
	}
	if p[0] == '\\' || p[0] == '/' {
		return true
	}
	if p[0] == '.' {
		return true
	}
	if len(p) >= 2 && p[1] == ':' {
		return false
	}
	return true
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir() || fi.Mode().IsRegular()
}
