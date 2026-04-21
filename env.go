package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ddkwork/golibrary/std/mylog"
	"golang.org/x/sys/windows/registry"
)

const EnvPath = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`

const StartupDir = `C:\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp`

const (
	EnvWindowsTargetPlatformVersion = "WindowsTargetPlatformVersion"
	EnvVCToolsInstallDir            = "VCToolsInstallDir"
	EnvWDKContentRoot               = "WDKContentRoot"
	EnvBuildLabSetupRoot            = "BuildLabSetupRoot"
	EnvVSINSTALLDIR                 = "VSINSTALLDIR"
	EnvINCLUDE                      = "INCLUDE"
	EnvLIB                          = "LIB"
	EnvWDKBinRoot                   = "WDKBinRoot"
	EnvDiaRoot                      = "DiaRoot"
	EnvCC                           = "CC"
	EnvCXX                          = "CXX"
)

type EnvVar struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Type       string `json:"type"`        // SZ, EXPAND_SZ, DWORD, MULTI_SZ, BINARY
	Valid      bool   `json:"valid"`       // if value is a path/file, does it exist?
	IsPath     bool   `json:"is_path"`     // does this look like a path?
	IsCompound bool   `json:"is_compound"` // is this a compound path-list variable (PATH, INCLUDE, LIB, etc.)?
	Reason     string `json:"reason"`      // why invalid (empty if valid)
}

type EnvVarList []EnvVar

type EnvManager interface {
	List() (EnvVarList, error)                                 // 列出所有系统环境变量，含类型、路径有效性校验
	Delete(name string) error                                  // 删除指定系统环境变量
	Set(name, value string) error                              // 设置/更新系统环境变量（SZ 类型）
	CreateStartupScript(content, name string) (string, error)  // 将内容写入启动目录生成开机自启脚本，返回完整路径
	Expand(value string) (string, error)                       // 展开字符串中的 %VAR% 环境变量引用
	CreateScheduledTask(isoPath string) error                  // 创建/更新 EWDK_Mount 计划任务（登录时自动挂载 ISO）
	DeleteScheduledTask()                                      // 删除 EWDK_Mount 计划任务
	IsMounted(isoPath string) bool                             // 检测指定 ISO 是否已挂载
	MountISO(isoPath string) (string, error)                   // 挂载 ISO 文件，返回盘符（自动设置 WDK 环境变量 + 开机任务）
	UnmountISO(isoPath string) error                           // 卸载指定 ISO（自动移除开机任务）
	UnmountAll() error                                         // 卸载所有已挂载的 EWDK ISO（自动移除开机任务）
	GetVirtualDiskPhysicalPath(isoPath string) (string, error) // 获取 ISO 的物理磁盘路径（\\.\PhysicalDriveX）
	CleanInvalidVars(shouldDelete func(string) bool) int       // 删除所有无效的环境变量，返回删除的数量
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

		isPath := looksLikePath(val)
		valid, reason := validateValue(val, isPath)

		value := mylog.Check2(m.getEnvValue(name))
		isCompound := strings.Count(value, ";") > 1

		typeName := typeToString(valType)
		result = append(result, EnvVar{
			Name:       name,
			Value:      val,
			Type:       typeName,
			Valid:      valid,
			IsPath:     isPath,
			IsCompound: isCompound,
			Reason:     reason,
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

	parts := strings.Split(value, ";")
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

func runSetupBuildEnv(setupCmd string) (ewdkEnv, error) {
	tmpFile := filepath.Join(os.TempDir(), "ewdk-env-after-setup.txt")
	os.Remove(tmpFile)
	cmd := exec.Command("C:\\Windows\\System32\\cmd.exe", "/c", setupCmd+" amd64 && set > "+tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	mylog.Check(cmd.Run())

	data := mylog.Check2(os.ReadFile(tmpFile))
	mylog.Json(string(data))
	defer os.Remove(tmpFile)

	result := make(map[string]string)
	for s := range strings.Lines(string(data)) {
		key, value, found := strings.Cut(s, "=")
		if found {
			switch key {
			case
				EnvWindowsTargetPlatformVersion,
				EnvVCToolsInstallDir,
				EnvWDKContentRoot,
				EnvBuildLabSetupRoot,
				EnvVSINSTALLDIR,
				EnvINCLUDE,
				EnvLIB,
				EnvWDKBinRoot:
			default:
				continue

			}
			result[key] = strings.TrimSuffix(value, "\r\n")
		}
	}
	DiaRoot := filepath.Join(result[EnvVSINSTALLDIR], "DIA SDK")

	include := strings.Split(result[EnvINCLUDE], ";")
	lib := strings.Split(result[EnvLIB], ";")

	pre := filepath.Join(result[EnvWDKContentRoot], "Include", result[EnvWindowsTargetPlatformVersion])
	include = append(include,
		filepath.Join(DiaRoot, "include"),
		filepath.Join(pre, "km"),
		filepath.Join(pre, "km", "crt"),
		filepath.Join(pre, "um"),
		filepath.Join(pre, "shared"),
		filepath.Join(result[EnvWDKContentRoot], "Include\\wdf\\kmdf\\1.35"),
	)

	lib = append(lib,
		filepath.Join(DiaRoot, "lib"),
		filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "km", "x64"),
		filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "um", "x64"),
	)

	return ewdkEnv{
		WindowsTargetPlatformVersion: result[EnvWindowsTargetPlatformVersion],
		WDKContentRoot:               result[EnvWDKContentRoot],
		BuildLabSetupRoot:            result[EnvBuildLabSetupRoot],
		VSINSTALLDIR:                 result[EnvVSINSTALLDIR],
		INCLUDE:                      include,
		LIB:                          lib,
		WDKBinRoot:                   result[EnvWDKBinRoot],
		DiaRoot:                      DiaRoot,
		VCToolsInstallDir:            result[EnvVCToolsInstallDir],
		CC:                           filepath.Join(result[EnvVCToolsInstallDir], "bin\\Hostx64\\x64\\cl.exe"),
	}, nil
}

type ewdkEnv struct {
	WindowsTargetPlatformVersion string
	WDKContentRoot               string
	BuildLabSetupRoot            string
	VSINSTALLDIR                 string
	INCLUDE                      []string
	LIB                          []string
	WDKBinRoot                   string
	DiaRoot                      string
	VCToolsInstallDir            string
	CC                           string
}

func (m *RegistryEnvManager) getEwdkDriveLetter() string {
	root, err := m.getEnvValue(EnvBuildLabSetupRoot)
	if err != nil || len(root) == 0 {
		return ""
	}
	letter := strings.ToUpper(string(root[0]))
	if letter < "A" || letter > "Z" {
		return ""
	}
	drivePath := letter + ":\\"
	if _, err := os.Stat(drivePath); os.IsNotExist(err) {
		return ""
	}
	setupCmd := drivePath + "BuildEnv\\SetupBuildEnv.cmd"
	if _, err := os.Stat(setupCmd); os.IsNotExist(err) {
		return ""
	}
	return letter
}

const powershell = "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"

func (m *RegistryEnvManager) IsMounted(isoPath string) bool {
	return m.getEwdkDriveLetter() != ""
}

func mountISO(isoPath string) (string, error) {
	script := fmt.Sprintf(`(Mount-DiskImage -ImagePath '%s' -PassThru | Get-Volume).DriveLetter`, strings.ReplaceAll(isoPath, "'", "''"))
	output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
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
		letter := m.getEwdkDriveLetter()
		m.CreateScheduledTask(isoPath)
		return letter, nil
	}
	letter, err := mountISO(isoPath)
	if err != nil {
		return "", fmt.Errorf("mount iso %s: %w", isoPath, err)
	}

	m.CreateScheduledTask(isoPath)
	return letter, nil
}

func (m *RegistryEnvManager) getEnvValue(name string) (string, error) {
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
	output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
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
	m.Delete(EnvWDKContentRoot)
	if err := unmountISO(isoPath); err != nil {
		return fmt.Errorf("unmount iso %s: %w", isoPath, err)
	}
	return nil
}

func (m *RegistryEnvManager) UnmountAll() error {
	return unmountISO("")
}

func (m *RegistryEnvManager) CleanInvalidVars(shouldDelete func(string) bool) int {

	vars := mylog.Check2(m.List())

	count := 0
	for _, envVar := range vars {
		if shouldDelete(envVar.Name) {
			mylog.Check(m.Delete(envVar.Name))
			continue
		}
		if !envVar.Valid {
			if !envVar.IsCompound {
				mylog.Check(m.Delete(envVar.Name))
				count++
				continue
			}
			// 处理复合路径变量（如PATH、INCLUDE、LIB等），拆分后过滤无效条目
			parts := strings.Split(envVar.Value, ";")
			var validParts []string
			seen := make(map[string]bool)

			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				expanded, err := registry.ExpandString(part)
				if err != nil {
					expanded = part
				}
				lower := strings.ToLower(expanded)
				if seen[lower] {
					continue
				}
				if _, statErr := os.Stat(expanded); os.IsNotExist(statErr) {
					continue
				}
				if isRelativePath(part) {
					continue
				}
				seen[lower] = true
				validParts = append(validParts, part)
			}

			if len(validParts) == 0 {
				mylog.Check(m.Delete(envVar.Name))
				count++
				continue
			}

			newValue := strings.Join(validParts, ";")
			mylog.Check(m.Set(envVar.Name, newValue))
			count++
		}
	}

	return count
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
	// 检查是否是绝对路径（例如 C:\）
	if len(p) >= 2 && p[1] == ':' {
		return false
	}
	return true
}

func unmountISO(isoPath string) error {
	if isoPath == "" {
		script := `Get-Volume | Where-Object { $_.DriveType -eq 'CD-ROM' } | ForEach-Object { Dismount-DiskImage -DevicePath "\\.\$($_.DriveLetter):" -ErrorAction SilentlyContinue }`
		output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).CombinedOutput()
		if err != nil {
			return fmt.Errorf("unmount all: %w, output: %s", err, strings.TrimSpace(string(output)))
		}
		return nil
	}
	script := fmt.Sprintf(`Dismount-DiskImage -ImagePath '%s'`, strings.ReplaceAll(isoPath, "'", "''"))
	if output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).CombinedOutput(); err != nil {
		return fmt.Errorf("unmount iso: %w, output: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
