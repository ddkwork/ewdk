package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const EnvPath = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`

const StartupDir = `C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp`

type EnvVar struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Type   string `json:"type"`    // SZ, EXPAND_SZ, DWORD, MULTI_SZ, BINARY
	Valid  bool   `json:"valid"`   // if value is a path/file, does it exist?
	IsPath bool   `json:"is_path"` // does this look like a path?
	Reason string `json:"reason"`  // why invalid (empty if valid)
}

type EnvVarList []EnvVar

type EnvVarDelta struct {
	Name     string `json:"name"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

type EnvDiff struct {
	Added   map[string]string `json:"added"`   // name -> new value (新增变量完整值，可直接喂给cmake)
	Changed []EnvVarDelta     `json:"changed"` // 变更详情：name + old_value + new_value
	Removed map[string]string `json:"removed"` // name -> old value (被删变量的最后值)
}

var cmakeEnvVars = map[string]string{
	"CMAKE_INCLUDE_PATH":   "include paths for find_file()",
	"CMAKE_LIBRARY_PATH":   "lib paths for find_library()",
	"CMAKE_PREFIX_PATH":    "prefix paths for find_*",
	"CMAKE_PROGRAM_PATH":   "program paths",
	"CMAKE_FRAMEWORK_PATH": "framework paths (macOS)",
	"CMAKE_APPBUNDLE_PATH": "app bundle paths (macOS)",
	"CC":                   "C compiler",
	"CFLAGS":               "C compiler flags",
	"CXX":                  "C++ compiler",
	"CXXFLAGS":             "C++ compiler flags",
	"LDFLAGS":              "linker flags",
	"ADSP_ROOT":            "Analog Devices SHARC root",
	"CSFLAGS":              "C# compiler flags",
	"CUDACXX":              "CUDA C++ compiler",
	"CUDAFLAGS":            "CUDA flags",
	"CUDAHOSTCXX":          "CUDA host C++ compiler",
	"FC":                   "Fortran compiler",
	"FFLAGS":               "Fortran flags",
	"HIPCXX":               "HIP C++ compiler",
	"HIPFLAGS":             "HIP flags",
	"ISPC":                 "Intel SPMD Program Compiler",
	"RC":                   "Windows resource compiler",
	"RCFLAGS":              "Windows resource compiler flags",
	"SWIFTC":               "Swift compiler",
	"OBJC":                 "Objective-C compiler",
	"OBJCFLAGS":            "Objective-C compiler flags",
	"OBJCXX":               "Objective-C++ compiler",
	"OBJCXXFLAGS":          "Objective-C++ compiler flags",
	"CMAKE_BUILD_TYPE":     "build type (Release/Debug etc.)",
	"CMAKE_GENERATOR":      "generator name",
	"CMAKE_INSTALL_PREFIX": "install prefix",
	"CMAKE_TOOLCHAIN_FILE": "toolchain file path",
	"DSTDIR":               "install destination root",
	"VERBOSE":              "verbose output flag",
}

type EnvManager interface {
	List() (EnvVarList, error)                                      // 列出所有系统环境变量，含类型、路径有效性校验
	Delete(name string) error                                       // 删除指定系统环境变量
	Set(name, value string) error                                   // 设置/更新系统环境变量（SZ 类型）
	CaptureDiff(setupCmd string) (*EnvDiff, error)                  // 执行 SetupBuildEnv.cmd 并捕获环境变量变更差异（含完整值）
	FillCMake(newVars map[string]string) (map[string]string, error) // 根据新增变量推断并填充 CMake 相关变量映射
	CreateStartupScript(content, name string) (string, error)       // 将内容写入启动目录生成开机自启脚本，返回完整路径
	Expand(value string) (string, error)                            // 展开字符串中的 %VAR% 环境变量引用
	GetMountedDriveLetter(isoPath string) string                    // 检测 EWDK ISO 已挂载的盘符（VHD + CD-ROM 回退）
	CreateScheduledTask(isoPath string) error                       // 创建/更新 EWDK_Mount 计划任务（登录时自动挂载 ISO）
	DeleteScheduledTask()                                           // 删除 EWDK_Mount 计划任务
	IsMounted(isoPath string) bool                                  // 检测指定 ISO 是否已挂载
	MountISO(isoPath string) (string, error)                        // 挂载 ISO 文件，返回盘符（自动设置 WDK 环境变量 + 开机任务）
	UnmountISO(isoPath string) error                                // 卸载指定 ISO（自动移除开机任务）
	UnmountAll() error                                              // 卸载所有已挂载的 EWDK ISO（自动移除开机任务）
	GetWDKContentRoot() (string, error)                             // 读取系统环境变量 WDKContentRoot
	GetWDKRoot() (string, error)                                    // 读取系统环境变量 WDK_ROOT
	GetEWDKSetupEnvCmd() (string, error)                            // 读取系统环境变量 EWDKSetupEnvCmd
	GetVirtualDiskPhysicalPath(isoPath string) (string, error)      // 获取 ISO 的物理磁盘路径（\\.\PhysicalDriveX）
	CleanInvalidVars() (int, error)                                 // 删除所有无效的环境变量，返回删除的数量
}

type RegistryEnvManager struct{}

func NewRegistryEnvManager() EnvManager { return &RegistryEnvManager{} }

var taskName = "EWDK_Mount"

func openEnvKey(access uint32) (registry.Key, error) {
	return registry.OpenKey(registry.LOCAL_MACHINE, EnvPath, access)
}

func (m *RegistryEnvManager) List() (EnvVarList, error) {
	key, err := openEnvKey(registry.QUERY_VALUE | registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()

	info, err := key.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat key: %w", err)
	}
	_ = info

	names, err := key.ReadValueNames(-1)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("read value names: %w", err)
	}

	var result EnvVarList
	for _, name := range names {
		_, valType, err := key.GetValue(name, nil)
		if err != nil {
			continue
		}
		var val string
		switch valType {
		case registry.SZ, registry.EXPAND_SZ:
			s, _, e := key.GetStringValue(name)
			if e != nil {
				continue
			}
			val = s
		default:
			n, _, e := key.GetValue(name, make([]byte, 0))
			if e != nil {
				val = fmt.Sprintf("<type=%d>", valType)
			} else {
				val = fmt.Sprintf("<%d bytes type=%d>", n, valType)
			}
		}

		typeName := typeToString(valType)
		isPath := looksLikePath(val)
		valid, reason := validateValue(val, isPath)

		result = append(result, EnvVar{
			Name:   name,
			Value:  val,
			Type:   typeName,
			Valid:  valid,
			IsPath: isPath,
			Reason: reason,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}

func (m *RegistryEnvManager) Delete(name string) error {
	key, err := openEnvKey(registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()

	err = key.DeleteValue(name)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

func (m *RegistryEnvManager) Set(name, value string) error {
	key, err := openEnvKey(registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()

	return key.SetStringValue(name, value)
}

func (m *RegistryEnvManager) CaptureDiff(setupCmd string) (*EnvDiff, error) {
	beforeMap, err := captureCurrentEnv()
	if err != nil {
		return nil, fmt.Errorf("capture before: %w", err)
	}

	afterMap, err := runSetupBuildEnv(setupCmd)
	if err != nil {
		return nil, fmt.Errorf("run setup build env: %w", err)
	}

	return computeDiff(beforeMap, afterMap), nil
}

func (m *RegistryEnvManager) FillCMake(newVars map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	for cmakeKey := range cmakeEnvVars {
		lowerKey := strings.ToLower(cmakeKey)
		for envName, envVal := range newVars {
			if strings.ToLower(envName) == lowerKey {
				result[cmakeKey] = envVal
				break
			}
		}

		if _, exists := result[cmakeKey]; !exists {
			envVal, found := inferCMakeValue(cmakeKey, newVars)
			if found {
				result[cmakeKey] = envVal
			}
		}
	}

	return result, nil
}

func (m *RegistryEnvManager) CreateStartupScript(content, name string) (string, error) {
	if err := os.MkdirAll(StartupDir, 0755); err != nil {
		return "", fmt.Errorf("create startup dir: %w", err)
	}

	destPath := filepath.Join(StartupDir, name)
	if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write startup script: %w", err)
	}

	return destPath, nil
}

func (m *RegistryEnvManager) Expand(value string) (string, error) {
	expanded, err := registry.ExpandString(value)
	if err != nil {
		return value, fmt.Errorf("expand env var: %w", err)
	}
	return expanded, nil
}

func (m *RegistryEnvManager) GetMountedDriveLetter(isoPath string) string {
	return findCdRomDriveLetter()
}

func (m *RegistryEnvManager) CreateScheduledTask(isoPath string) error {
	m.DeleteScheduledTask()

	script := fmt.Sprintf(`Mount-DiskImage -ImagePath '%s' -PassThru | Out-Null`, strings.ReplaceAll(isoPath, "'", "''"))
	psCmd := fmt.Sprintf(`powershell.exe -ExecutionPolicy Bypass -Command "%s"`, script)

	exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()

	err := exec.Command("schtasks", "/Create",
		"/TN", taskName,
		"/TR", psCmd,
		"/SC", "ONLOGON",
		"/DELAY", "0000:00:10",
		"/F",
	).Run()
	if err != nil {
		return fmt.Errorf("create scheduled task: %w", err)
	}

	fmt.Printf("Task registered: %s\n", taskName)
	fmt.Println("Triggers: At logon (delay 10s)")
	return nil
}

func (m *RegistryEnvManager) DeleteScheduledTask() {
	exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()
}

func typeToString(valType uint32) string {
	switch valType {
	case registry.NONE:
		return "NONE"
	case registry.SZ:
		return "SZ"
	case registry.EXPAND_SZ:
		return "EXPAND_SZ"
	case registry.BINARY:
		return "BINARY"
	case registry.DWORD:
		return "DWORD"
	case registry.DWORD_BIG_ENDIAN:
		return "DWORD_BIG_ENDIAN"
	case registry.LINK:
		return "LINK"
	case registry.MULTI_SZ:
		return "MULTI_SZ"
	case registry.RESOURCE_LIST:
		return "RESOURCE_LIST"
	case registry.FULL_RESOURCE_DESCRIPTOR:
		return "FULL_RESOURCE_DESCRIPTOR"
	case registry.RESOURCE_REQUIREMENTS_LIST:
		return "RESOURCE_REQUIREMENTS_LIST"
	case registry.QWORD:
		return "QWORD"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", valType)
	}
}

func looksLikePath(value string) bool {
	if len(value) == 0 {
		return false
	}
	if len(value) >= 2 && value[1] == ':' {
		return true
	}
	if strings.Contains(value, "\\") || strings.Contains(value, "/") {
		return true
	}
	if strings.HasPrefix(value, "%") && strings.Contains(value, "%\\") {
		return true
	}
	return false
}

func validateValue(value string, isPath bool) (bool, string) {
	if !isPath || len(value) == 0 {
		return true, ""
	}

	parts := splitPathList(value)
	allValid := true
	var invalidParts []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		expanded, err := registry.ExpandString(part)
		if err != nil {
			expanded = part
		}
		if _, err := os.Stat(expanded); os.IsNotExist(err) {
			allValid = false
			invalidParts = append(invalidParts, part)
		}
	}

	if !allValid {
		return false, fmt.Sprintf("missing paths: %s", strings.Join(invalidParts, "; "))
	}
	return true, ""
}

func splitPathList(value string) []string {
	return strings.Split(value, ";")
}

func captureCurrentEnv() (map[string]string, error) {
	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	names, err := key.ReadValueNames(-1)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	result := make(map[string]string, len(names))
	for _, name := range names {
		val, valType, err := key.GetStringValue(name)
		if err != nil {
			continue
		}
		if valType == registry.SZ || valType == registry.EXPAND_SZ {
			result[name] = val
		}
	}
	return result, nil
}

func runSetupBuildEnv(setupCmd string) (map[string]string, error) {
	tmpFile := filepath.Join(os.TempDir(), "ewdk-env-after-setup.txt")
	os.Remove(tmpFile)

	// 先运行 setupCmd，然后输出所有环境变量
	// 使用更安全的方式构建命令，避免引号嵌套问题
	cmd := exec.Command("cmd", "/c", setupCmd+" && set > "+tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec setup cmd: %w", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("read env file: %w", err)
	}
	defer os.Remove(tmpFile)

	result := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		name := line[:idx]
		value := line[idx+1:]
		result[name] = value
	}
	return result, nil
}

func computeDiff(before, after map[string]string) *EnvDiff {
	diff := &EnvDiff{
		Added:   make(map[string]string),
		Changed: []EnvVarDelta{},
		Removed: make(map[string]string),
	}

	for name, oldVal := range before {
		newVal, exists := after[name]
		if !exists {
			diff.Removed[name] = oldVal
		} else if oldVal != newVal {
			diff.Changed = append(diff.Changed, EnvVarDelta{
				Name:     name,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}

	for name, newVal := range after {
		if _, exists := before[name]; !exists {
			diff.Added[name] = newVal
		}
	}

	sort.Slice(diff.Changed, func(i, j int) bool {
		return strings.ToLower(diff.Changed[i].Name) < strings.ToLower(diff.Changed[j].Name)
	})

	return diff
}

func inferCMakeValue(cmakeKey string, availableVars map[string]string) (string, bool) {
	switch cmakeKey {
	case "CMAKE_INCLUDE_PATH":
		if v, ok := availableVars["INCLUDE"]; ok {
			return v, true
		}
	case "CMAKE_LIBRARY_PATH":
		if v, ok := availableVars["LIB"]; ok {
			return v, true
		}
	case "CC":
		if v, ok := availableVars["CC"]; ok {
			return v, true
		}
	case "CXX":
		if v, ok := availableVars["CXX"]; ok {
			return v, true
		}
	case "CFLAGS":
		if v, ok := availableVars["CFLAGS"]; ok {
			return v, true
		}
	case "CXXFLAGS":
		if v, ok := availableVars["CXXFLAGS"]; ok {
			return v, true
		}
	case "LDFLAGS":
		if v, ok := availableVars["LDFLAGS"]; ok {
			return v, true
		}
	case "FC":
		if v, ok := availableVars["FC"]; ok {
			return v, true
		}
	case "FFLAGS":
		if v, ok := availableVars["FFLAGS"]; ok {
			return v, true
		}
	case "RC":
		if v, ok := availableVars["RC"]; ok {
			return v, true
		}
	case "RCFLAGS":
		if v, ok := availableVars["RCFLAGS"]; ok {
			return v, true
		}
	case "SWIFTC":
		if v, ok := availableVars["SWIFTC"]; ok {
			return v, true
		}
	case "CMAKE_PREFIX_PATH":
		if v, ok := availableVars["WDK_ROOT"]; ok {
			return v, true
		}
		if v, ok := availableVars["WDKContentRoot"]; ok {
			return v, true
		}
	case "CMAKE_TOOLCHAIN_FILE":
		if v, ok := availableVars["EWDKSetupEnvCmd"]; ok {
			dir := filepath.Dir(v)
			toolchain := filepath.Join(dir, "..", "cmake", "Toolchain-Windows.cmake")
			return toolchain, true
		}
	}
	return "", false
}

func findCdRomDriveLetter() string {
	kernel32 := windows.MustLoadDLL("kernel32.dll")

	getLogicalDrives, _ := kernel32.FindProc("GetLogicalDrives")
	getDriveType, _ := kernel32.FindProc("GetDriveTypeW")

	ret, _, _ := getLogicalDrives.Call()
	if ret == 0 {
		return ""
	}

	for i := 0; i < 26; i++ {
		if (ret & (1 << uint(i))) != 0 {
			letter := string(rune('A' + i))
			drivePath := fmt.Sprintf("%s:\\", letter)
			drivePathPtr, _ := windows.UTF16PtrFromString(drivePath)

			driveType, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(drivePathPtr)))
			if driveType == windows.DRIVE_CDROM {
				testPath := fmt.Sprintf("%s:\\BuildEnv\\SetupBuildEnv.cmd", letter)
				if _, err := os.Stat(testPath); err == nil {
					return letter
				}
			}
		}
	}

	return ""
}

func (m *RegistryEnvManager) IsMounted(isoPath string) bool {
	return m.GetMountedDriveLetter(isoPath) != ""
}

func mountISO(isoPath string) (string, error) {
	script := fmt.Sprintf(`(Mount-DiskImage -ImagePath '%s' -PassThru | Get-Volume).DriveLetter`, strings.ReplaceAll(isoPath, "'", "''"))
	output, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
	if err != nil {
		return "", fmt.Errorf("mount iso: %w", err)
	}
	letter := strings.TrimSpace(string(output))
	if letter == "" || len(letter) != 1 {
		return "", fmt.Errorf("invalid drive letter: %q", letter)
	}
	return letter, nil
}

func (m *RegistryEnvManager) MountISO(isoPath string) (string, error) {
	if m.IsMounted(isoPath) {
		letter := m.GetMountedDriveLetter(isoPath)
		m.setEWDKEnvVars(letter)
		m.verifyEWDKEnvVars(letter)
		m.CreateScheduledTask(isoPath)
		return letter, nil
	}
	letter, err := mountISO(isoPath)
	if err != nil {
		return "", fmt.Errorf("mount iso %s: %w", isoPath, err)
	}
	m.setEWDKEnvVars(letter)
	if err := m.verifyEWDKEnvVars(letter); err != nil {
		return "", fmt.Errorf("env var verification failed: %w", err)
	}
	m.CreateScheduledTask(isoPath)
	return letter, nil
}

func (m *RegistryEnvManager) verifyEWDKEnvVars(driveLetter string) error {
	expectedRoot := fmt.Sprintf("%s:\\Program Files\\Windows Kits\\10", driveLetter)
	expectedCmd := fmt.Sprintf("%s:\\BuildEnv\\SetupBuildEnv.cmd", driveLetter)

	checks := map[string]string{
		"WDKContentRoot":  expectedRoot,
		"WDK_ROOT":        expectedRoot,
		"EWDKSetupEnvCmd": expectedCmd,
	}
	for name, expected := range checks {
		got, err := m.getEnvVar(name)
		if err != nil {
			return fmt.Errorf("%s not set: %w", name, err)
		}
		if got != expected {
			return fmt.Errorf("%s mismatch: got %q, want %q", name, got, expected)
		}
	}
	return nil
}

func (m *RegistryEnvManager) setEWDKEnvVars(driveLetter string) {
	wdkRoot := fmt.Sprintf("%s:\\Program Files\\Windows Kits\\10", driveLetter)
	setupEnvCmd := fmt.Sprintf("%s:\\BuildEnv\\SetupBuildEnv.cmd", driveLetter)
	m.Set("WDKContentRoot", wdkRoot)
	m.Set("WDK_ROOT", wdkRoot)
	m.Set("EWDKSetupEnvCmd", setupEnvCmd)
}

func (m *RegistryEnvManager) GetWDKContentRoot() (string, error) {
	return m.getEnvVar("WDKContentRoot")
}

func (m *RegistryEnvManager) GetWDKRoot() (string, error) {
	return m.getEnvVar("WDK_ROOT")
}

func (m *RegistryEnvManager) GetEWDKSetupEnvCmd() (string, error) {
	return m.getEnvVar("EWDKSetupEnvCmd")
}

func (m *RegistryEnvManager) getEnvVar(name string) (string, error) {
	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("open registry: %w", err)
	}
	defer key.Close()

	val, valType, err := key.GetStringValue(name)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", name, err)
	}
	if valType != registry.SZ && valType != registry.EXPAND_SZ {
		return "", fmt.Errorf("%s is type %d, not SZ/EXPAND_SZ", name, valType)
	}
	return val, nil
}

func (m *RegistryEnvManager) GetVirtualDiskPhysicalPath(isoPath string) (string, error) {
	script := fmt.Sprintf(`(Get-DiskImage -ImagePath '%s' | Get-Disk).Path`, strings.ReplaceAll(isoPath, "'", "''"))
	output, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
	if err != nil {
		return "", fmt.Errorf("get physical path: %w", err)
	}
	path := strings.TrimSpace(string(output))
	if path == "" {
		return "", fmt.Errorf("disk image not found or not attached: %s", isoPath)
	}
	return path, nil
}

func (m *RegistryEnvManager) UnmountISO(isoPath string) error {
	if !m.IsMounted(isoPath) {
		return nil
	}
	m.DeleteScheduledTask()
	m.Delete("WDKContentRoot")
	m.Delete("WDK_ROOT")
	m.Delete("EWDKSetupEnvCmd")
	if err := m.verifyEWDKEnvVarsCleared(); err != nil {
		return fmt.Errorf("env var cleanup verification failed: %w", err)
	}
	if err := unmountISO(isoPath); err != nil {
		return fmt.Errorf("unmount iso %s: %w", isoPath, err)
	}
	return nil
}

func (m *RegistryEnvManager) verifyEWDKEnvVarsCleared() error {
	names := []string{"WDKContentRoot", "WDK_ROOT", "EWDKSetupEnvCmd"}
	for _, name := range names {
		val, err := m.getEnvVar(name)
		if err == nil && val != "" {
			return fmt.Errorf("%s still exists: %q", name, val)
		}
	}
	return nil
}

func (m *RegistryEnvManager) UnmountAll() error {
	return unmountISO("")
}

func (m *RegistryEnvManager) CleanInvalidVars() (int, error) {
	vars, err := m.List()
	if err != nil {
		return 0, fmt.Errorf("list env vars: %w", err)
	}

	count := 0
	for _, envVar := range vars {
		if !envVar.Valid {
			// 对于 PATH 变量，只清理无效路径部分，不删除整个变量
			if strings.EqualFold(envVar.Name, "PATH") {
				if err := m.cleanInvalidPathParts(envVar.Name, envVar.Value); err != nil {
					fmt.Printf("Failed to clean invalid path parts in %s: %v\n", envVar.Name, err)
					continue
				}
				count++
			} else {
				// 对于其他变量，直接删除
				err := m.Delete(envVar.Name)
				if err != nil {
					// 记录错误但继续处理其他无效变量
					fmt.Printf("Failed to delete invalid var %s: %v\n", envVar.Name, err)
					continue
				}
				count++
			}
		}
	}

	return count, nil
}

// cleanInvalidPathParts 清理环境变量中的无效路径部分
func (m *RegistryEnvManager) cleanInvalidPathParts(name, value string) error {
	parts := splitPathList(value)
	var validParts []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		expanded, err := registry.ExpandString(part)
		if err != nil {
			expanded = part
		}
		if _, err := os.Stat(expanded); !os.IsNotExist(err) {
			validParts = append(validParts, part)
		}
	}

	// 如果所有路径都无效，保持原样不修改
	if len(validParts) == 0 {
		return nil
	}

	// 更新环境变量，只保留有效路径
	newValue := strings.Join(validParts, ";")
	return m.Set(name, newValue)
}

func unmountISO(isoPath string) error {
	if isoPath == "" {
		script := `Get-Volume | Where-Object { $_.DriveType -eq 'CD-ROM' } | ForEach-Object { Dismount-DiskImage -DevicePath "\\.\$($_.DriveLetter):" -ErrorAction SilentlyContinue }`
		output, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).CombinedOutput()
		if err != nil {
			return fmt.Errorf("unmount all: %w, output: %s", err, strings.TrimSpace(string(output)))
		}
		return nil
	}
	script := fmt.Sprintf(`Dismount-DiskImage -ImagePath '%s'`, strings.ReplaceAll(isoPath, "'", "''"))
	if output, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).CombinedOutput(); err != nil {
		return fmt.Errorf("unmount iso: %w, output: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func extractDriveLetterFromDevicePath(devicePath string) string {
	devicePath = strings.TrimSpace(devicePath)
	for i := 1; i < len(devicePath); i++ {
		if devicePath[i] == ':' {
			c := devicePath[i-1]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				return strings.ToUpper(string(c))
			}
		}
	}
	return ""
}
