package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ddkwork/golibrary/cmake"
	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
)

//go:embed km_crt_stubs.c
var kmCrtStubs string

const (
	EnvWindowsTargetPlatformVersion = "WindowsTargetPlatformVersion"
	EnvVCToolsInstallDir            = "VCToolsInstallDir"
	EnvWDKContentRoot               = "WDKContentRoot"
	EnvBuildLabSetupRoot            = "BuildLabSetupRoot"
	EnvVSINSTALLDIR                 = "VSINSTALLDIR"
	EnvINCLUDE                      = "INCLUDE"
	EnvLIB                          = "LIB"
	EnvWDKBinRoot                   = "WDKBinRoot"
)

const powershell = "powershell.exe"

type ewdkCommonEnv struct {
	WDKContentRoot               string
	WindowsTargetPlatformVersion string
	VCToolsInstallDir            string
	WDKBinRoot                   string
	DiaRoot                      string
	VSINSTALLDIR                 string
	BuildLabSetupRoot            string
	CC                           string
	CCX86                        string
	RC                           string
	MT                           string
	SignTool                     string
	NTDDKFile                    string
}

type ewdkKMEnv struct {
	IncludeDirs []string
	LibDirs     []string
}

type ewdkUMEnv struct {
	IncludeDirs        []string
	LibDirs            []string
	LibDirsX86         []string
	MfcIncludeDir      string
	MfcLibDirX64       string
	MfcLibDirX86       string
	QtIncludeDirs      []string
	QtPluginDirs       []string
	QtVersion          string
	QtCompileDefs      []string
	QtCompileOptions   []string
	QtLinkLibs         []string
	MsvcRuntimeLibrary string
}

type ewdkEnv struct {
	Common ewdkCommonEnv
	KM     ewdkKMEnv
	UM     ewdkUMEnv
}

func main() {

	info := cmake.Module()
	mylog.Success(info)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "unmount":
			isoPath := resolveISOPath()
			mylog.Check(unmountISO(isoPath))
			mylog.Success("EWDK unmounted")
			return
		case "clean":
			cleanGenerated()
			return
		}
	}

	isoPath := resolveISOPath()
	mylog.Success("ISO: ", isoPath)

	var driveLetter string
	if isMounted() {
		driveLetter = getEwdkDriveLetter()
		mylog.Success("EWDK already mounted at: ", driveLetter)
	} else {
		driveLetter = mylog.Check2(mountISO(isoPath))
		mylog.Success("EWDK mounted at: ", driveLetter)
	}

	setupEnvCmd := driveLetter + ":\\BuildEnv\\SetupBuildEnv.cmd"
	mylog.Success("SetupCmd: ", setupEnvCmd)
	env := mylog.Check2(runSetupBuildEnv(setupEnvCmd))
	mylog.Struct(env)

	mylog.Check(generateEwdkCmake(env, cmake.EwdkCmakeFile))
	mylog.Success("Generated: ", cmake.EwdkCmakeFile)

	unityFile := filepath.Join(cmake.BinDir, "unity.cmake")
	mylog.Check(generateUnityCmake(unityFile))
	mylog.Success("Generated: ", unityFile)

	stream.CopyFile("ninja.exe", filepath.Join(cmake.BinDir, "ninja.exe"))

	ensureTestCertificate()

	envData := mylog.Check2(json.MarshalIndent(env, "", "  "))
	mylog.Check(os.WriteFile(cmake.EwdkEnvFile, envData, 0644))
	mylog.Success("Generated: ", cmake.EwdkEnvFile)

	mylog.Success("Environment ready. Run build.bat to start building.")
}

func getExecDir() string {
	if workspace := os.Getenv("GITHUB_WORKSPACE"); workspace != "" {
		return workspace
	}
	execDir, _ := os.Executable()
	return filepath.Dir(execDir)
}

func resolveISOPath() string {
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		iso := filepath.Join(os.Getenv("TEMP"), "ewdk.iso")
		if !stream.FileExists(iso) {
			panic("CI env: ewdk.iso not found in TEMP")
		}
		return iso
	}
	execDir := getExecDir()
	files, _ := filepath.Glob(filepath.Join(execDir, "EWDK_*.iso"))
	if len(files) == 0 {
		// Fall back to current working directory (useful when running via "go run .")
		cwd, _ := os.Getwd()
		files, _ = filepath.Glob(filepath.Join(cwd, "EWDK_*.iso"))
	}
	if len(files) == 0 {
		panic("No ISO file found in: " + execDir)
	}
	return files[0]
}

const testSignCertName = "WDKTestCert"

func ensureTestCertificate() {
	// Check if cert already exists in TrustedPublisher (the only store we keep)
	script := fmt.Sprintf(
		`$cert = Get-ChildItem Cert:\LocalMachine\TrustedPublisher | Where-Object { $_.Subject -match 'CN=%s' }; `+
			`if ($cert) { Write-Host $cert.Thumbprint } else { Write-Host 'NOT_FOUND' }`,
		testSignCertName)
	output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
	if err != nil {
		fmt.Printf("  [WARN] check certificate: %v\n", err)
		return
	}
	thumbprint := strings.TrimSpace(string(output))
	if thumbprint != "NOT_FOUND" && thumbprint != "" {
		fmt.Printf("  [OK]   Test certificate already in TrustedPublisher (thumbprint: %s)\n", thumbprint)
		removeCertFromAllOtherStores()
		return
	}

	// Create cert in LocalMachine\My (New-SelfSignedCertificate only supports My store)
	createScript := fmt.Sprintf(
		`$cert = New-SelfSignedCertificate -Type CodeSigningCert `+
			`-Subject "CN=%s" `+
			`-CertStoreLocation "Cert:\LocalMachine\My"; `+
			`Write-Host $cert.Thumbprint`,
		testSignCertName)
	out, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", createScript).CombinedOutput()
	if err != nil {
		fmt.Printf("  [WARN] create certificate: %v, output: %s\n", err, strings.TrimSpace(string(out)))
		return
	}
	thumbprint = strings.TrimSpace(string(out))
	fmt.Printf("  [OK]   Created test certificate (thumbprint: %s)\n", thumbprint)

	// Export .cer to ewdk bin dir
	certDir := cmake.BinDir
	cerPath := filepath.Join(certDir, testSignCertName+".cer")
	exportScript := fmt.Sprintf(
		`$cert = Get-ChildItem Cert:\LocalMachine\My\%s; `+
			`Export-Certificate -Cert $cert -FilePath '%s' -Type CERT | Out-Null; `+
			`Write-Host 'EXPORTED'`,
		thumbprint, strings.ReplaceAll(cerPath, "'", "''"))
	if out2, err2 := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", exportScript).CombinedOutput(); err2 != nil {
		fmt.Printf("  [WARN] export certificate: %v, output: %s\n", err2, strings.TrimSpace(string(out2)))
	} else {
		fmt.Printf("  [OK]   Exported certificate to %s\n", cerPath)
	}

	// Install to TrustedPublisher
	installScript := fmt.Sprintf(
		`$cert = Get-ChildItem Cert:\LocalMachine\My\%s; `+
			`$store = [System.Security.Cryptography.X509Certificates.X509Store]::new("TrustedPublisher", "LocalMachine"); `+
			`$store.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite); `+
			`$store.Add($cert); `+
			`$store.Close(); `+
			`Write-Host 'INSTALLED'`,
		thumbprint)
	if out3, err3 := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", installScript).CombinedOutput(); err3 != nil {
		fmt.Printf("  [WARN] install to TrustedPublisher: %v, output: %s\n", err3, strings.TrimSpace(string(out3)))
	} else {
		fmt.Printf("  [OK]   Installed to LocalMachine\\TrustedPublisher\n", testSignCertName)
	}

	// Remove from My — signtool reads from TrustedPublisher via /sm /s TrustedPublisher
	removeCertFromLocalMachineMy(thumbprint)
	removeCertFromAllOtherStores()
}

func removeCertFromLocalMachineMy(thumbprint string) {
	removeFromStore := fmt.Sprintf(
		`try { `+
			`$s = [System.Security.Cryptography.X509Certificates.X509Store]::new("My","LocalMachine"); `+
			`$s.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite); `+
			`$c = $s.Certificates | Where-Object { $_.Thumbprint -eq '%s' }; `+
			`if ($c) { $s.Remove($c); Write-Host 'Removed from LocalMachine\My' } else { Write-Host 'Not in LocalMachine\My' }; `+
			`$s.Close() } catch { Write-Host 'skip' }`,
		thumbprint)
	exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", removeFromStore).Run()
}

func removeCertFromAllOtherStores() {
	// Clean CurrentUser\My, LocalMachine\Root, LocalMachine\My (any stale copies)
	stores := []string{
		`$s = [System.Security.Cryptography.X509Certificates.X509Store]::new("My","CurrentUser"); $s.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite); $s.Certificates | Where-Object { $_.Subject -match 'CN=WDKTestCert' } | ForEach-Object { $s.Remove($_); Write-Host "Removed from CurrentUser\My: $($_.Thumbprint)" }; $s.Close()`,
		`$s = [System.Security.Cryptography.X509Certificates.X509Store]::new("Root","LocalMachine"); $s.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite); $s.Certificates | Where-Object { $_.Subject -match 'CN=WDKTestCert' } | ForEach-Object { $s.Remove($_); Write-Host "Removed from LocalMachine\Root: $($_.Thumbprint)" }; $s.Close()`,
	}
	for _, s := range stores {
		exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", s).Run()
	}
}

func generateUnityCmake(outputPath string) error {
	content := `# ── Source file collection ───────────────────────────────────
# 复制此文件到项目中使用。
#
# 不用 CMake 原生 UNITY_BUILD（聚合编译）的原因：
#   1. 不同源文件中的同名 static 变量/函数会冲突
#   2. CMake 自动分桶(batch)策略不可控，排查问题困难
#   3. x86 交叉编译(custom command mode)不支持 CMake 原生 unity
# 改用手动收集 + 可选 unity.cpp 方案，控制权交给开发者。
#
# collect_sources(dir1 dir2 ... outvar)：
#   扫描指定目录下的所有 .cpp 和 .c 文件，存入 outvar。
#   发现 .asm 文件时输出警告（ewdk 模板已自动处理 asm）。
function(collect_sources)
    set(_dirs ${ARGN})
    list(POP_BACK _dirs _outvar)
    set(_files)
    foreach(_d ${_dirs})
        file(GLOB _cpp "${CMAKE_CURRENT_SOURCE_DIR}/${_d}/*.cpp")
        list(APPEND _files ${_cpp})
        file(GLOB _c "${CMAKE_CURRENT_SOURCE_DIR}/${_d}/*.c")
        list(APPEND _files ${_c})
        file(GLOB _asm "${CMAKE_CURRENT_SOURCE_DIR}/${_d}/*.asm")
        if(_asm)
            message(WARNING "collect_sources: .asm files in ${_d} are NOT included — "
                            "ewdk templates auto-handle asm per architecture")
        endif()
    endforeach()
    set(${_outvar} ${_files} PARENT_SCOPE)
endfunction()

# generate_unity(unity_name src1 src2 ...)：
#   生成一个 unity.cpp 文件，用绝对路径 #include 所有源文件。
#   编译此文件替代逐个编译，可获得合并编译的加速效果。
#   用绝对路径而非相对路径，以兼容 x86 自定义命令模式。
#   如遇 static 变量冲突，切回逐个编译即可。
function(generate_unity _unity_name)
    set(_sources ${ARGN})
    set(_unity_file "${CMAKE_CURRENT_BINARY_DIR}/${_unity_name}.cpp")
    file(WRITE ${_unity_file} "// Unity build - auto-generated\n")
    foreach(_src ${_sources})
        get_filename_component(_abs "${_src}" ABSOLUTE)
        file(APPEND ${_unity_file} "#include \"${_abs}\"\n")
    endforeach()
    set(UNITY_FILE "${_unity_file}" PARENT_SCOPE)
endfunction()
`
	return os.WriteFile(outputPath, []byte(content), 0644)
}

func cleanGenerated() {
	files := []string{cmake.EwdkCmakeFile, filepath.Join(cmake.BinDir, "unity.cmake")}
	for _, f := range files {
		if stream.FileExists(f) {
			mylog.Check(os.Remove(f))
			fmt.Printf("  [OK]   Removed: %s\n", f)
		}
	}
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

func isMounted() bool {
	return getEwdkDriveLetter() != ""
}

func getEwdkDriveLetter() string {
	for c := 'F'; c >= 'A'; c-- {
		setupCmd := string(c) + ":\\BuildEnv\\SetupBuildEnv.cmd"
		if _, err := os.Stat(setupCmd); err == nil {
			return string(c)
		}
	}
	return ""
}

type qtStaticInfo struct {
	IncludeDirs      []string
	PluginDirs       []string
	Version          string
	CompileDefs      []string
	CompileOptions   []string
	LinkLibs         []string
	MsvcRuntimeLibrary string
}

func scanQtStaticDir(qtBaseDir string) qtStaticInfo {
	var info qtStaticInfo
	info.MsvcRuntimeLibrary = "MultiThreaded$<$<CONFIG:Debug>:Debug>"
	includeDir := filepath.Join(qtBaseDir, "include")
	if entries, err := os.ReadDir(includeDir); err == nil {
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "Qt") {
				info.IncludeDirs = append(info.IncludeDirs, filepath.Join(includeDir, e.Name()))
			}
		}
	}
	pluginDir := filepath.Join(qtBaseDir, "plugins")
	if entries, err := os.ReadDir(pluginDir); err == nil {
		for _, cat := range entries {
			if cat.IsDir() {
				catPath := filepath.Join(pluginDir, cat.Name())
				if subEntries, err := os.ReadDir(catPath); err == nil {
					for _, p := range subEntries {
						name := p.Name()
						if strings.HasSuffix(strings.ToLower(name), ".lib") || p.IsDir() {
							info.PluginDirs = append(info.PluginDirs, catPath)
							break
						}
					}
				}
			}
		}
	}
	qtCoreDir := filepath.Join(includeDir, "QtCore")
	if entries, err := os.ReadDir(qtCoreDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				parts := strings.Split(e.Name(), ".")
				if len(parts) >= 3 {
					info.Version = fmt.Sprintf("%s.%s.%s", parts[0], parts[1], parts[2])
					break
				}
			}
		}
	}
	if info.Version == "" {
		panic(fmt.Sprintf("Qt version detection failed: %s not found or empty. Check CI yml download/extract step", qtCoreDir))
	}
	libDir := filepath.Join(qtBaseDir, "lib")
	info.CompileDefs = []string{
		"QT_BUILDING_QT",
		"MIQT_BUILDING_DLL",
		"MIQT_WINDOWSQTSTATIC",
        "QT_STATIC",
		"QT_NO_CAST_FROM_ASCII",
		"QT_NO_CAST_TO_ASCII",
		"QT_NO_EXCEPTIONS",
		"QT_NO_DEBUG",
		fmt.Sprintf("QT_VERSION_MAJOR=%s", strings.Split(info.Version, ".")[0]),
		fmt.Sprintf("QT_VERSION_MINOR=%s", strings.Split(info.Version, ".")[1]),
		fmt.Sprintf("QT_VERSION_PATCH=%s", strings.Split(info.Version, ".")[2]),
		"Q_CORE_EXPORT=",
		"Q_GUI_EXPORT=",
		"Q_WIDGETS_EXPORT=",
		"_WIN32_WINNT=0x0602",
	}
	info.CompileOptions = []string{
		"/Zc:__cplusplus",
		"/bigobj",
		"/MP",
		"/EHsc",
		"/permissive-",
		"/utf-8",
	}
	entries, err := os.ReadDir(libDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lib") {
				name := entry.Name()
				if strings.HasPrefix(name, "Qt6") {
					info.LinkLibs = append(info.LinkLibs, name)
				}
			}
		}
	}
	pluginLibs := []string{"qwindows.lib", "qmodernwindowsstyle.lib"}
	for _, plib := range pluginLibs {
		info.LinkLibs = append(info.LinkLibs, plib)
	}
	winLibs := []string{
		"icuuc.lib", "icuin.lib",
		"advapi32.lib", "shell32.lib", "ole32.lib", "oleaut32.lib", "uuid.lib",
		"user32.lib", "gdi32.lib", "comdlg32.lib", "winspool.lib", "imm32.lib", "version.lib", "ws2_32.lib",
		"dwmapi.lib", "d3d9.lib", "dwrite.lib", "dxgi.lib", "netapi32.lib", "opengl32.lib",
		"uiautomationcore.lib", "uxtheme.lib",
		"shlwapi.lib", "authz.lib", "userenv.lib", "ntdll.lib", "winmm.lib",
		"runtimeobject.lib", "setupapi.lib", "d3d11.lib", "d3d12.lib", "dxguid.lib",
		"shcore.lib", "wtsapi32.lib",
		"kernel32.lib", "Mpr.lib", "Secur32.lib", "Iphlpapi.lib", "Winhttp.lib", "Dnsapi.lib",
		"Propsys.lib", "Avrt.lib",
	}
	info.LinkLibs = append(info.LinkLibs, winLibs...)
	return info
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

	pre := filepath.Join(result[EnvWDKContentRoot], "Include", result[EnvWindowsTargetPlatformVersion])
	ucrtInc := filepath.Join(result[EnvWDKContentRoot], "Include", result[EnvWindowsTargetPlatformVersion], "ucrt")
	ucrtLib := filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "ucrt", "x64")
	umLib := filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "um", "x64")

	// x86 user-mode lib dirs
	umLibX86 := filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "um", "x86")
	ucrtLibX86 := filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "ucrt", "x86")

	vcInc := []string{
		filepath.Join(result[EnvVCToolsInstallDir], "include"),
		filepath.Join(result[EnvVCToolsInstallDir], "ATLMFC", "include"),
		filepath.Join(result[EnvVSINSTALLDIR], "VC", "Auxiliary", "VS", "include"),
		filepath.Join(DiaRoot, "include"),
	}
	vcLib := []string{
		filepath.Join(result[EnvVCToolsInstallDir], "lib", "x64"),
		filepath.Join(result[EnvVCToolsInstallDir], "ATLMFC", "lib", "x64"),
		filepath.Join(DiaRoot, "lib"),
	}
	// x86 MSVC lib dirs
	vcLibX86 := []string{
		filepath.Join(result[EnvVCToolsInstallDir], "lib", "x86"),
		filepath.Join(result[EnvVCToolsInstallDir], "ATLMFC", "lib", "x86"),
	}

	cc := filepath.Join(result[EnvVCToolsInstallDir], "bin\\Hostx64\\x64\\cl.exe")
	ccX86 := filepath.Join(result[EnvVCToolsInstallDir], "bin\\Hostx64\\x86\\cl.exe")
	rc := filepath.Join(result[EnvWDKContentRoot], "bin", result[EnvWindowsTargetPlatformVersion], "x64", "rc.exe")
	mt := filepath.Join(result[EnvWDKContentRoot], "bin", result[EnvWindowsTargetPlatformVersion], "x64", "mt.exe")
	signtool := filepath.Join(result[EnvWDKContentRoot], "bin", result[EnvWindowsTargetPlatformVersion], "x64", "signtool.exe")
	ntddkFile := filepath.Join(result[EnvWDKContentRoot], "Include", result[EnvWindowsTargetPlatformVersion], "km", "ntddk.h")

	common := ewdkCommonEnv{
		WDKContentRoot:               result[EnvWDKContentRoot],
		WindowsTargetPlatformVersion: result[EnvWindowsTargetPlatformVersion],
		VCToolsInstallDir:            result[EnvVCToolsInstallDir],
		WDKBinRoot:                   result[EnvWDKBinRoot],
		DiaRoot:                      DiaRoot,
		VSINSTALLDIR:                 result[EnvVSINSTALLDIR],
		BuildLabSetupRoot:            result[EnvBuildLabSetupRoot],
		CC:                           cc,
		CCX86:                        ccX86,
		RC:                           rc,
	}
	if _, err := os.Stat(mt); err == nil {
		common.MT = mt
	}
	if _, err := os.Stat(signtool); err == nil {
		common.SignTool = signtool
	}
	if _, err := os.Stat(ntddkFile); err == nil {
		common.NTDDKFile = ntddkFile
	}

	km := ewdkKMEnv{
		IncludeDirs: []string{
			filepath.Join(pre, "shared"),
			filepath.Join(pre, "km"),
			filepath.Join(pre, "km", "crt"),
		},
		LibDirs: []string{
			filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "km", "x64"),
		},
	}

	execDir := getExecDir()
	qtBaseDir := filepath.Join(execDir, "qt_static")
	qtInclude := filepath.Join(qtBaseDir, "include")
	qtLib := filepath.Join(qtBaseDir, "lib")
	qtInfo := scanQtStaticDir(qtBaseDir)

	um := ewdkUMEnv{
		IncludeDirs: append([]string{
			filepath.Join(pre, "shared"),
			filepath.Join(pre, "um"),
			ucrtInc,
			qtInclude,
		}, vcInc...),
		LibDirs: append([]string{umLib, ucrtLib, qtLib,
			filepath.Join(qtBaseDir, "plugins", "platforms"),
			filepath.Join(qtBaseDir, "plugins", "styles"),
		}, vcLib...),
		LibDirsX86: append([]string{umLibX86, ucrtLibX86, qtLib}, vcLibX86...),
		MfcIncludeDir: filepath.Join(result[EnvVCToolsInstallDir], "ATLMFC", "include"),
		MfcLibDirX64:  filepath.Join(result[EnvVCToolsInstallDir], "ATLMFC", "lib", "x64"),
		MfcLibDirX86:  filepath.Join(result[EnvVCToolsInstallDir], "ATLMFC", "lib", "x86"),
		QtIncludeDirs:      qtInfo.IncludeDirs,
		QtPluginDirs:       qtInfo.PluginDirs,
		QtVersion:          qtInfo.Version,
		QtCompileDefs:      qtInfo.CompileDefs,
		QtCompileOptions:   qtInfo.CompileOptions,
		QtLinkLibs:         qtInfo.LinkLibs,
		MsvcRuntimeLibrary: qtInfo.MsvcRuntimeLibrary,
	}

	return ewdkEnv{
		Common: common,
		KM:     km,
		UM:     um,
	}, nil
}

func generateEwdkCmake(env ewdkEnv, outputPath string) error {
	cm := func(p string) string {
		p = strings.ReplaceAll(p, `\`, "/")
		p = strings.TrimRight(p, "/")
		return p
	}

	writeList := func(b *strings.Builder, name string, dirs []string) {
		b.WriteString(fmt.Sprintf("set(%s \"", name))
		for i, p := range dirs {
			if i > 0 {
				b.WriteString(";")
			}
			b.WriteString(cm(p))
		}
		b.WriteString("\")\n")
	}

	c := env.Common

	var b strings.Builder
	b.WriteString("# Auto-generated by ewdk toolchain manager - DO NOT EDIT\n")
	b.WriteString("# include(ewdk.cmake) provides WDK detection, compiler setup, and all wdk_* / um_* functions\n")

	b.WriteString("\n# ---- WDK Core ----\n")
	b.WriteString("set(WDK_FOUND TRUE)\n")
	b.WriteString(fmt.Sprintf("set(WDK_ROOT \"%s\")\n", cm(c.WDKContentRoot)))
	b.WriteString(fmt.Sprintf("set(WDK_VERSION \"%s\")\n", c.WindowsTargetPlatformVersion))
	b.WriteString(fmt.Sprintf("set(WDK_INC_VERSION \"%s\")\n", c.WindowsTargetPlatformVersion))
	b.WriteString(fmt.Sprintf("set(WDK_LIB_VERSION \"%s\")\n", c.WindowsTargetPlatformVersion))
	b.WriteString("set(WDK_PLATFORM \"x64\")\n")
	b.WriteString(fmt.Sprintf("set(VC_INCLUDE_DIR \"%s/include\")\n", cm(c.VCToolsInstallDir)))
	b.WriteString(fmt.Sprintf("set(UCRT_INCLUDE_DIR \"%s/Include/%s/ucrt\")\n", cm(c.WDKContentRoot), c.WindowsTargetPlatformVersion))

	b.WriteString("\n# ---- KM (Kernel-Mode) ----\n")
	writeList(&b, "WDK_KM_INCLUDE_DIRS", env.KM.IncludeDirs)
	writeList(&b, "WDK_KM_LIB_DIRS", env.KM.LibDirs)
	if len(env.KM.IncludeDirs) >= 2 {
		fmt.Fprintf(&b, "set(WDK_SHARED_INCLUDE_DIR \"%s\")\n", cm(env.KM.IncludeDirs[0]))
		fmt.Fprintf(&b, "set(WDK_KM_INCLUDE_DIR \"%s\")\n", cm(env.KM.IncludeDirs[1]))
	}

	b.WriteString("\n# ---- UM (User-Mode) ----\n")
	writeList(&b, "WDK_UM_INCLUDE_DIRS", env.UM.IncludeDirs)
	writeList(&b, "WDK_UM_LIB_DIRS", env.UM.LibDirs)
	b.WriteString(`
# 通用 Windows SDK libs（所有 UM 目标自动链接）
set(WDK_UM_SDK_LIBS
    kernel32.lib user32.lib gdi32.lib shell32.lib
    ole32.lib oleaut32.lib advapi32.lib ws2_32.lib
    shlwapi.lib comctl32.lib wbemuuid.lib psapi.lib
    winmm.lib iphlpapi.lib mpr.lib
)
# WDK 扩展 libs（DDK 设备/电源相关）
set(WDK_UM_EXTRA_LIBS setupapi.lib powrprof.lib slwga.lib)
# KM 内核模式默认 libs（km_sys/km_lib 自动链接）
set(WDK_KM_SDK_LIBS ntoskrnl.lib hal.lib wmilib.lib)
set(WDK_KM_EXTRA_LIBS bufferoverflowk.lib)

# ── ewdk.cmake 函数清单 ──────────────────────────────
# km_sys(target)           — 内核驱动 .sys
# km_lib(target)           — 内核静态库 .lib
# um_exe(target)           — 用户态 EXE（自动链接 ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS}）
# um_lib(target)           — 用户态静态库
# um_dll(target)           — 用户态 DLL
# um_dp64(target)          — x64 x64dbg 插件 .dp64
# um_dp86(target)          — x86 x32dbg 插件 .dp32（交叉编译）
# um_exe_x86(target)       — x86 EXE（交叉编译）
# um_dll_x86(target)       — x86 DLL（交叉编译）
# um_lib_x86(target)       — x86 静态库（交叉编译）
# um_exe_mfc(target)       — MFC EXE（x64）
# um_exe_mfc_x86(target)   — MFC EXE（x86 交叉编译）
# collect_sources()        — 源码收集（来自 unity.cmake）
# generate_unity()         — 合并编译（来自 unity.cmake）
# ──────────────────────────────────────────────────
`)

	b.WriteString("\n# ---- UM x86 ----\n")
	b.WriteString(fmt.Sprintf("set(WDK_UM_INCLUDE_DIRS_X86 \"${WDK_UM_INCLUDE_DIRS}\")\n"))
	writeList(&b, "WDK_UM_LIB_DIRS_X86", env.UM.LibDirsX86)
	b.WriteString(fmt.Sprintf("set(X86_CL \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.CCX86)))
	b.WriteString(fmt.Sprintf("set(X86_ML \"%s\")\n", cm(
		filepath.Join(c.VCToolsInstallDir, "bin", "Hostx64", "x86", "ml.exe"),
	)))
	b.WriteString(fmt.Sprintf("set(X86_LINK \"%s\")\n", cm(
		filepath.Join(c.VCToolsInstallDir, "bin", "Hostx64", "x86", "link.exe"),
	)))
	b.WriteString(fmt.Sprintf("set(X86_RC \"%s\")\n", cm(
		filepath.Join(c.WDKContentRoot, "bin", c.WindowsTargetPlatformVersion, "x86", "rc.exe"),
	)))

	b.WriteString("\n# ---- MFC ----\n")
	b.WriteString(fmt.Sprintf("set(MFC_INCLUDE_DIR \"%s\")\n", cm(env.UM.MfcIncludeDir)))
	b.WriteString(fmt.Sprintf("set(MFC_LIB_DIR_X64 \"%s\")\n", cm(env.UM.MfcLibDirX64)))
	b.WriteString(fmt.Sprintf("set(MFC_LIB_DIR_X86 \"%s\")\n", cm(env.UM.MfcLibDirX86)))

	if len(env.UM.QtIncludeDirs) > 0 {
		b.WriteString("\n# ---- Qt Static ----\n")
		b.WriteString(fmt.Sprintf("set(QT_VERSION \"%s\")\n", env.UM.QtVersion))
		writeList(&b, "QT_INCLUDE_DIRS", env.UM.QtIncludeDirs)
		writeList(&b, "QT_PLUGIN_DIRS", env.UM.QtPluginDirs)
		b.WriteString("set(QT_BASE_DIR \"\")\n")
		b.WriteString(`foreach(_inc ${WDK_UM_INCLUDE_DIRS})
    if(_inc MATCHES "qt_static")
        get_filename_component(QT_BASE_DIR "${_inc}" DIRECTORY)
        break()
    endif()
endforeach()
`)
		b.WriteString("set(QT_LIB_DIR \"${QT_BASE_DIR}/lib\")\n")
		b.WriteString("set(QT_PLUGIN_DIR \"${QT_BASE_DIR}/plugins\")\n")
		b.WriteString(fmt.Sprintf("set(QT_MSVC_RUNTIME_LIBRARY \"%s\")\n", env.UM.MsvcRuntimeLibrary))
		b.WriteString("\n# Qt Compile Definitions\n")
		b.WriteString("set(QT_COMPILE_DEFINITIONS\n")
		for _, def := range env.UM.QtCompileDefs {
			b.WriteString(fmt.Sprintf("    %s\n", def))
		}
		b.WriteString(")\n")
		b.WriteString("\n# Qt Compile Options\n")
		b.WriteString("set(QT_COMPILE_OPTIONS\n")
		for _, opt := range env.UM.QtCompileOptions {
			b.WriteString(fmt.Sprintf("    %s\n", opt))
		}
		b.WriteString(")\n")
		b.WriteString("\n# Qt Link Libraries\n")
		b.WriteString("set(QT_LINK_LIBRARIES\n")
		for _, lib := range env.UM.QtLinkLibs {
			b.WriteString(fmt.Sprintf("    %s\n", cm(lib)))
		}
		b.WriteString(")\n")
	}

	b.WriteString("\n# ---- Compiler ----\n")
	b.WriteString(fmt.Sprintf("set(CMAKE_C_COMPILER \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.CC)))
	b.WriteString(fmt.Sprintf("set(CMAKE_CXX_COMPILER \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.CC)))
	b.WriteString(fmt.Sprintf("set(CMAKE_RC_COMPILER \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.RC)))
	b.WriteString("set(CMAKE_C_COMPILER_WORKS 1 CACHE INTERNAL \"\")\n")
	b.WriteString("set(CMAKE_CXX_COMPILER_WORKS 1 CACHE INTERNAL \"\")\n")
	b.WriteString("set(CMAKE_C_STANDARD_LIBRARIES \"\")\n")
	b.WriteString("set(CMAKE_CXX_STANDARD_LIBRARIES \"\")\n")
	b.WriteString("set(CMAKE_MSVC_RUNTIME_LIBRARY \"\" CACHE STRING \"\" FORCE)\n")
	b.WriteString("set(CMAKE_INCLUDE_PATH \"${WDK_KM_INCLUDE_DIRS};${WDK_UM_INCLUDE_DIRS}\" CACHE STRING \"\" FORCE)\n")
	b.WriteString("set(CMAKE_LIBRARY_PATH \"${WDK_KM_LIB_DIRS};${WDK_UM_LIB_DIRS}\" CACHE STRING \"\" FORCE)\n")
	b.WriteString(fmt.Sprintf("list(APPEND CMAKE_PROGRAM_PATH \"%s\")\n", cm(filepath.Dir(c.CC))))
	if _, err := os.Stat(c.RC); err == nil {
		b.WriteString(fmt.Sprintf("list(APPEND CMAKE_PROGRAM_PATH \"%s\")\n", cm(filepath.Dir(c.RC))))
	}
	if c.MT != "" {
		b.WriteString(fmt.Sprintf("set(CMAKE_MT \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.MT)))
	}

	b.WriteString("\nset(ENV{WDKContentRoot} \"${WDK_ROOT}\")\n")
	b.WriteString("set(ENV{INCLUDE} \"${WDK_KM_INCLUDE_DIRS};${WDK_UM_INCLUDE_DIRS}\")\n")
	b.WriteString("set(ENV{LIB} \"${WDK_KM_LIB_DIRS};${WDK_UM_LIB_DIRS}\")\n")

	b.WriteString("\n# ---- WDK Settings ----\n")
	b.WriteString("set(WDK_WINVER \"0x0601\" CACHE STRING \"Default WINVER for WDK targets\")\n")
	b.WriteString("set(WDK_NTDDI_VERSION \"\" CACHE STRING \"Specified NTDDI_VERSION for WDK targets if needed\")\n")
	b.WriteString("if(DEFINED ENV{GITHUB_ACTIONS})\n")
	b.WriteString("    set(KM_TEST_SIGN OFF CACHE BOOL \"Enable test signing for drivers (disabled in CI)\")\n")
	b.WriteString("else()\n")
	b.WriteString("    set(KM_TEST_SIGN ON CACHE BOOL \"Enable test signing for drivers\")\n")
	b.WriteString("endif()\n")
	b.WriteString("set(KM_TEST_SIGN_NAME \"WDKTestCert\" CACHE STRING \"Certificate name for test signing\")\n")

	b.WriteString(`
set(KM_ADDITIONAL_FLAGS_FILE "${CMAKE_CURRENT_BINARY_DIR}${CMAKE_FILES_DIRECTORY}/wdkflags.h")
file(WRITE ${KM_ADDITIONAL_FLAGS_FILE} "#pragma runtime_checks(\"suc\", off)\n#pragma warning(disable: 4117)")

set(KM_COMPILE_FLAGS
    "/Zp8"
    "/GF"
    "/GR-"
    "/Gz"
    "/FIwarning.h"
    "/FI${KM_ADDITIONAL_FLAGS_FILE}"
    "/Oi"
    )

set(KM_COMPILE_DEFINITIONS "WINNT=1;_AMD64_;AMD64;_msvc;_amd64")
set(KM_COMPILE_DEFINITIONS_DEBUG "MSC_NOOPT;DEPRECATE_DDK_FUNCTIONS=1;DBG=1")

string(CONCAT KM_LINK_FLAGS
    "/MANIFEST:NO "
    "/DRIVER "
    "/OPT:REF "
    "/INCREMENTAL:NO "
    "/OPT:ICF "
    "/SUBSYSTEM:NATIVE "
    "/ENTRY:DriverEntry "
    "/MERGE:_TEXT=.text;_PAGE=PAGE "
    "/NODEFAULTLIB "
    "/SECTION:INIT,d "
    "/VERSION:10.0 "
    "/FORCE:MULTIPLE "
    )
`)

	b.WriteString("\n# ---- KM Libraries ----\n")
	b.WriteString(fmt.Sprintf("file(GLOB KM_LIB_FILES \"%s/*.lib\")\n", cm(env.KM.LibDirs[0])))
	b.WriteString(`foreach(LIB_FILE ${KM_LIB_FILES})
    get_filename_component(LIB_NAME ${LIB_FILE} NAME_WE)
    string(TOUPPER ${LIB_NAME} LIB_NAME)
    add_library(WDK::${LIB_NAME} INTERFACE IMPORTED)
    set_property(TARGET WDK::${LIB_NAME} PROPERTY INTERFACE_LINK_LIBRARIES ${LIB_FILE})
endforeach()
`)

	if c.SignTool != "" {
		b.WriteString(fmt.Sprintf("\nset(KM_SIGNTOOL_PATH \"%s\")\n", cm(c.SignTool)))
	}

	b.WriteString(`

# ---- x64dbg Plugin Support ----
# Read x64dbg/x32dbg install paths from registry (cached once at configure time)
if(NOT DEFINED X64DBG_X64_DIR)
    execute_process(
        COMMAND powershell -NoLogo -NoProfile -Command
            "(Get-ItemPropertyValue 'Registry::HKEY_CLASSES_ROOT\\.dd64\\DefaultIcon' '(Default)')"
        OUTPUT_VARIABLE _X64
        OUTPUT_STRIP_TRAILING_WHITESPACE
        ERROR_QUIET
    )
    if(_X64)
        get_filename_component(X64DBG_X64_DIR "${_X64}" DIRECTORY)
        message(STATUS "x64dbg found: ${X64DBG_X64_DIR}")
    else()
        set(X64DBG_X64_DIR "")
        message(STATUS "x64dbg not found (registry key missing)")
    endif()
endif()
if(NOT DEFINED X64DBG_X32_DIR)
    execute_process(
        COMMAND powershell -NoLogo -NoProfile -Command
            "(Get-ItemPropertyValue 'Registry::HKEY_CLASSES_ROOT\\.dd32\\DefaultIcon' '(Default)')"
        OUTPUT_VARIABLE _X32
        OUTPUT_STRIP_TRAILING_WHITESPACE
        ERROR_QUIET
    )
    if(_X32)
        get_filename_component(X64DBG_X32_DIR "${_X32}" DIRECTORY)
        message(STATUS "x32dbg found: ${X64DBG_X32_DIR}")
    else()
        set(X64DBG_X32_DIR "")
        message(STATUS "x32dbg not found (registry key missing)")
    endif()
endif()

# ---- KM/UM Functions ----

function(km_sys _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "LIBS;DEFINES;INCLUDES" ${ARGN})

    # Strip CMake's default /EHsc (kernel mode doesn't want exceptions) — write to cache
    string(REPLACE "/EHsc" "" _km_cxx_flags "${CMAKE_CXX_FLAGS}")
    set(CMAKE_CXX_FLAGS "${_km_cxx_flags}" CACHE STRING "Flags used by the compiler during all build types." FORCE)

    # Clear CMake's default MSVC link libraries (kernel32.lib etc.)
    set(CMAKE_C_STANDARD_LIBRARIES "" CACHE STRING "" FORCE)
    set(CMAKE_CXX_STANDARD_LIBRARIES "" CACHE STRING "" FORCE)

    add_executable(${_target} ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES SUFFIX ".sys")
    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "${KM_COMPILE_DEFINITIONS};$<$<CONFIG:Debug>:${KM_COMPILE_DEFINITIONS_DEBUG}>;_WIN32_WINNT=${WDK_WINVER}"
        )
    set_target_properties(${_target} PROPERTIES LINK_FLAGS "${KM_LINK_FLAGS}")
    set_target_properties(${_target} PROPERTIES LINK_INTERFACE_LIBRARIES "")
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded")

    # Kernel-mode C++ flags (RTTI off, exceptions off, disable CRT debug libs)
    target_compile_options(${_target} PRIVATE
        $<$<COMPILE_LANGUAGE:C,CXX>:/GR->
        $<$<COMPILE_LANGUAGE:C,CXX>:/EHs-c->
        )
    foreach(_flag ${KM_COMPILE_FLAGS})
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:${_flag}>)
    endforeach()

    # /kernel via target_compile_options so PCH compilation picks it up
    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/kernel>)

    # Apply /kernel to each source file individually (allows per-file override)
    get_target_property(_ks_sources ${_target} SOURCES)
    foreach(_ks_src ${_ks_sources})
        if(_ks_src MATCHES "\\.(c|cpp|cxx)$")
            set_source_files_properties(${_ks_src} PROPERTIES COMPILE_FLAGS "/kernel")
        endif()
    endforeach()

    target_link_options(${_target} PRIVATE
        /NODEFAULTLIB:libucrtd.lib
        /NODEFAULTLIB:ucrtd.lib
        )

    # Generate CRT stub for kernel-mode ucrt symbols (embedded from km_crt_stubs.c)
    set(_km_crt_stub "${CMAKE_CURRENT_BINARY_DIR}/km_crt_stubs.c")
    file(WRITE ${_km_crt_stub} [=[` + kmCrtStubs + `]=])
    target_sources(${_target} PRIVATE ${_km_crt_stub})
    set_source_files_properties(${_km_crt_stub} PROPERTIES COMPILE_FLAGS "/kernel")

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()
    if(WDK_DEFINES)
        target_compile_definitions(${_target} PRIVATE ${WDK_DEFINES})
    endif()

    # MSVC/ucrt include must come BEFORE WDK km include for correct header resolution
    target_include_directories(${_target} BEFORE PRIVATE
        "${VC_INCLUDE_DIR}"
        "${UCRT_INCLUDE_DIR}"
        )
    target_include_directories(${_target} PRIVATE ${WDK_KM_INCLUDE_DIRS})
    if(WDK_INCLUDES)
        target_include_directories(${_target} PRIVATE ${WDK_INCLUDES})
    endif()

    target_link_libraries(${_target} WDK::NTOSKRNL WDK::HAL WDK::WMILIB WDK::LIBCNTPR)

    if(WDK::BUFFEROVERFLOWK)
        target_link_libraries(${_target} WDK::BUFFEROVERFLOWK)
    else()
        target_link_libraries(${_target} WDK::BUFFEROVERFLOWFASTFAILK)
    endif()

    if(WDK_LIBS)
        target_link_libraries(${_target} ${WDK_LIBS})
    endif()

    set_property(TARGET ${_target} APPEND_STRING PROPERTY LINK_FLAGS "/ENTRY:GsDriverEntry")

    if(KM_TEST_SIGN AND DEFINED KM_SIGNTOOL_PATH)
        add_custom_command(TARGET ${_target} POST_BUILD
            COMMAND ${KM_SIGNTOOL_PATH} sign /fd SHA256 /sm /s TrustedPublisher /n ${KM_TEST_SIGN_NAME} $<TARGET_FILE:${_target}>
            WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
            COMMENT "Signing driver with test certificate: ${KM_TEST_SIGN_NAME}"
            VERBATIM
        )
    endif()
endfunction()

# km_sys_cpp was removed — use km_sys instead (supports both C and C++).
# The /kernel flag is now applied per-file, allowing per-file override.

function(km_lib _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "LIBS;DEFINES;INCLUDES" ${ARGN})

    # Strip CMake's default /EHsc (kernel mode doesn't want exceptions) — write to cache
    string(REPLACE "/EHsc" "" _km_cxx_flags "${CMAKE_CXX_FLAGS}")
    set(CMAKE_CXX_FLAGS "${_km_cxx_flags}" CACHE STRING "Flags used by the compiler during all build types." FORCE)

    add_library(${_target} ${WDK_UNPARSED_ARGUMENTS})

    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8>)
    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "${KM_COMPILE_DEFINITIONS};$<$<CONFIG:Debug>:${KM_COMPILE_DEFINITIONS_DEBUG};>_WIN32_WINNT=${WDK_WINVER}"
        )
    # Kernel-mode C++ flags (RTTI off, exceptions off)
    target_compile_options(${_target} PRIVATE
        $<$<COMPILE_LANGUAGE:C,CXX>:/GR->
        $<$<COMPILE_LANGUAGE:C,CXX>:/EHs-c->
        )
    foreach(_flag ${KM_COMPILE_FLAGS})
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:${_flag}>)
    endforeach()
    
    # /kernel via target_compile_options so PCH compilation picks it up
    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/kernel>)
    
    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()
    if(WDK_DEFINES)
        target_compile_definitions(${_target} PRIVATE ${WDK_DEFINES})
    endif()

    # MSVC/ucrt include must come BEFORE WDK km include
    target_include_directories(${_target} BEFORE PRIVATE
        "${VC_INCLUDE_DIR}"
        "${UCRT_INCLUDE_DIR}"
        )
    target_include_directories(${_target} PRIVATE ${WDK_KM_INCLUDE_DIRS})
    if(WDK_INCLUDES)
        target_include_directories(${_target} PRIVATE ${WDK_INCLUDES})
    endif()
    if(WDK_LIBS)
        target_link_libraries(${_target} ${WDK_LIBS})
    endif()
endfunction()

function(um_exe _target)
    cmake_parse_arguments(WDK "NOAUTO" "SUBSYSTEM;WINVER;NTDDI_VERSION" "SOURCES;INCLUDES;DEFINES;LIBS;COMPILE_OPTIONS" ${ARGN})

    if(NOT WDK_SUBSYSTEM)
        set(WDK_SUBSYSTEM "CONSOLE")
    endif()

    if(WDK_SOURCES)
        add_executable(${_target} ${WDK_SOURCES})
    else()
        add_executable(${_target} ${WDK_UNPARSED_ARGUMENTS})
    endif()

    string(TOUPPER "${WDK_SUBSYSTEM}" _subsystem_upper)

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE")
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")
    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()
    if(WDK_DEFINES)
        target_compile_definitions(${_target} PRIVATE ${WDK_DEFINES})
    endif()
    if(WDK_COMPILE_OPTIONS)
        target_compile_options(${_target} PRIVATE ${WDK_COMPILE_OPTIONS})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})
    if(WDK_INCLUDES)
        target_include_directories(${_target} PRIVATE ${WDK_INCLUDES})
    endif()

    if(WDK_NOAUTO)
        # Don't auto-link kernel32/user32 - caller provides all libs
        if(WDK_LIBS)
            target_link_libraries(${_target} ${WDK_LIBS})
        endif()
    else()
        target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS})
        if(WDK_LIBS)
            target_link_libraries(${_target} ${WDK_LIBS})
        endif()
    endif()

    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()

    if(_subsystem_upper STREQUAL "CONSOLE" OR _subsystem_upper STREQUAL "WINCON")
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:CONSOLE")
    elseif(_subsystem_upper STREQUAL "WINDOWS" OR _subsystem_upper STREQUAL "WIN")
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:WINDOWS")
    endif()
endfunction()

function(um_lib _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "SOURCES;INCLUDES;DEFINES;LIBS;COMPILE_OPTIONS" ${ARGN})

    if(WDK_SOURCES)
        add_library(${_target} ${WDK_SOURCES})
    else()
        add_library(${_target} ${WDK_UNPARSED_ARGUMENTS})
    endif()

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE")
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()
    if(WDK_DEFINES)
        target_compile_definitions(${_target} PRIVATE ${WDK_DEFINES})
    endif()
    if(WDK_COMPILE_OPTIONS)
        target_compile_options(${_target} PRIVATE ${WDK_COMPILE_OPTIONS})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})
    if(WDK_INCLUDES)
        target_include_directories(${_target} PRIVATE ${WDK_INCLUDES})
    endif()
    if(WDK_LIBS)
        target_link_libraries(${_target} ${WDK_LIBS})
    endif()
    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()
endfunction()

function(um_dll _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "SOURCES;INCLUDES;DEFINES;LIBS;COMPILE_OPTIONS" ${ARGN})

    if(WDK_SOURCES)
        add_library(${_target} SHARED ${WDK_SOURCES})
    else()
        add_library(${_target} SHARED ${WDK_UNPARSED_ARGUMENTS})
    endif()

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;_USRDLL;_WINDLL"
        )

    if(DEFINED QT_INCLUDE_DIRS AND NOT "${QT_INCLUDE_DIRS}" STREQUAL "")
        cmake_policy(SET CMP0117 NEW)
        set_target_properties(${_target} PROPERTIES
            CXX_STANDARD 17
            CXX_STANDARD_REQUIRED ON
            CXX_EXTENSIONS OFF
            MSVC_RUNTIME_LIBRARY "${QT_MSVC_RUNTIME_LIBRARY}"
        )
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
        target_include_directories(${_target} PRIVATE ${QT_INCLUDE_DIRS} ${CMAKE_CURRENT_SOURCE_DIR})
        target_compile_definitions(${_target} PRIVATE ${QT_COMPILE_DEFINITIONS})
        target_compile_options(${_target} PRIVATE ${QT_COMPILE_OPTIONS})
        target_link_libraries(${_target} ${QT_LINK_LIBRARIES})
    else()
        set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
    endif()

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()
    if(WDK_DEFINES)
        target_compile_definitions(${_target} PRIVATE ${WDK_DEFINES})
    endif()
    if(WDK_COMPILE_OPTIONS)
        target_compile_options(${_target} PRIVATE ${WDK_COMPILE_OPTIONS})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})
    if(WDK_INCLUDES)
        target_include_directories(${_target} PRIVATE ${WDK_INCLUDES})
    endif()
    target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS})
    if(WDK_LIBS)
        target_link_libraries(${_target} ${WDK_LIBS})
    endif()
    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()
endfunction()

# ---- x64dbg Plugin Functions ----

function(um_dp64 _target)
    # x64 x64dbg plugin → .dp64, auto-copy to x64dbg plugins dir
    # Auto: /utf-8 /EHsc, common win libs, all pluginsdk libs
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION;PCH" "SOURCES;DEFINES;COMPILE_OPTIONS;INCLUDE_DIRS;LINK_LIBS;EXTRA_SOURCES" ${ARGN})

    if(WDK_SOURCES)
        add_library(${_target} SHARED ${WDK_SOURCES})
    else()
        add_library(${_target} SHARED ${WDK_UNPARSED_ARGUMENTS})
    endif()
    if(WDK_EXTRA_SOURCES)
        target_sources(${_target} PRIVATE ${WDK_EXTRA_SOURCES})
    endif()
    if(WDK_PCH)
        target_precompile_headers(${_target} PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/${WDK_PCH}")
    endif()
    set_target_properties(${_target} PROPERTIES SUFFIX ".dp64")
    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;_USRDLL;_WINDLL;_WINDOWS"
    )
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()
    if(WDK_DEFINES)
        target_compile_definitions(${_target} PRIVATE ${WDK_DEFINES})
    endif()

    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/EHsc /std:c++latest>)
    if(WDK_COMPILE_OPTIONS)
        target_compile_options(${_target} PRIVATE ${WDK_COMPILE_OPTIONS})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})
    target_include_directories(${_target} PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
    target_include_directories(${_target} PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/pluginsdk")
    if(WDK_INCLUDE_DIRS)
        target_include_directories(${_target} PRIVATE ${WDK_INCLUDE_DIRS})
    endif()

    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()
    target_link_directories(${_target} PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}")
    target_link_libraries(${_target}
        kernel32.lib user32.lib gdi32.lib advapi32.lib shell32.lib
        ole32.lib oleaut32.lib uuid.lib comdlg32.lib winspool.lib
    )
    # Auto-link all x64 pluginsdk libs
    target_link_libraries(${_target}
        pluginsdk/x64bridge
        pluginsdk/x64dbg
    )
    file(GLOB_RECURSE _DP64_SDK_LIBS "${CMAKE_CURRENT_SOURCE_DIR}/pluginsdk/*_x64.lib")
    if(_DP64_SDK_LIBS)
        foreach(_f ${_DP64_SDK_LIBS})
            string(REPLACE "${CMAKE_CURRENT_SOURCE_DIR}/" "" _rel "${_f}")
            string(REGEX REPLACE "\\.lib$" "" _rel "${_rel}")
            target_link_libraries(${_target} "${_rel}")
        endforeach()
    endif()
    if(WDK_LINK_LIBS)
        target_link_libraries(${_target} ${WDK_LINK_LIBS})
    endif()

    if(X64DBG_X64_DIR)
        add_custom_command(TARGET ${_target} POST_BUILD
            COMMAND ${CMAKE_COMMAND} -E make_directory "${X64DBG_X64_DIR}/plugins"
            COMMAND ${CMAKE_COMMAND} -E copy "$<TARGET_FILE:${_target}>" "${X64DBG_X64_DIR}/plugins/$<TARGET_FILE_NAME:${_target}>"
            COMMAND powershell -NoLogo -NoProfile -Command
                "if(Test-Path '${X64DBG_X64_DIR}/plugins/$<TARGET_FILE_NAME:${_target}>'){exit 0}else{exit 1}"
            COMMAND ${CMAKE_COMMAND} -E echo "[OK] $<TARGET_FILE_NAME:${_target}> copied to ${X64DBG_X64_DIR}/plugins/"
        )
    endif()
endfunction()

function(um_dp86 _target)
    # x86 x32dbg plugin → .dp32 (cross-compile), auto-copy to x32dbg plugins dir
    # Auto: /utf-8 /EHsc, common win libs, all pluginsdk x86 libs
    # Internal target name is ${_target}_x86 to allow same name as um_dp64
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION;PCH" "SOURCES;DEFINES;COMPILE_OPTIONS;LINK_LIBS;INCLUDE_DIRS" ${ARGN})

    set(_target_internal "${_target}_x86")
    set(_output "${CMAKE_CURRENT_BINARY_DIR}/${_target}.dp32")
    set(_implib "${CMAKE_CURRENT_BINARY_DIR}/${_target}_x86.lib")
    set(_obj_dir "${CMAKE_CURRENT_BINARY_DIR}/${_target}_x86_objs")
    file(MAKE_DIRECTORY "${_obj_dir}")

    if(CMAKE_BUILD_TYPE STREQUAL "Debug")
        set(_CRT "/MTd")
        set(_OPT "/Od" "/Ob0" "/DDEBUG")
    else()
        set(_CRT "/MT")
        set(_OPT "/O2" "/Ob2" "/DNDEBUG")
    endif()

    set(_inc_flags)
    foreach(_d ${WDK_UM_INCLUDE_DIRS_X86})
        list(APPEND _inc_flags "/I\"${_d}\"")
    endforeach()
    list(APPEND _inc_flags "/I\"${CMAKE_CURRENT_SOURCE_DIR}\"")
    list(APPEND _inc_flags "/I\"${CMAKE_CURRENT_SOURCE_DIR}/pluginsdk\"")
    foreach(_d ${WDK_INCLUDE_DIRS})
        list(APPEND _inc_flags "/I\"${_d}\"")
    endforeach()

    set(_def_flags "/D_WIN32_WINNT=${WDK_WINVER}" "/DWIN32" "/D_WINDOWS" "/D_USRDLL" "/D_WINDLL" "/DUNICODE" "/D_UNICODE")
    foreach(_def ${WDK_DEFINES})
        list(APPEND _def_flags "/D${_def}")
    endforeach()

    set(_libpaths)
    foreach(_d ${WDK_UM_LIB_DIRS_X86})
        list(APPEND _libpaths "/LIBPATH:\"${_d}\"")
    endforeach()
    list(APPEND _libpaths "/LIBPATH:\"${CMAKE_CURRENT_SOURCE_DIR}\"")
    list(APPEND _libpaths "/LIBPATH:\"${CMAKE_CURRENT_SOURCE_DIR}/pluginsdk\"")
    # Auto-detect pluginsdk subdirectories with libs
    file(GLOB_RECURSE _DP86_SDK_FILES "${CMAKE_CURRENT_SOURCE_DIR}/pluginsdk/*_x86.lib")
    set(_added_subdirs "")
    foreach(_f ${_DP86_SDK_FILES})
        get_filename_component(_d "${_f}" DIRECTORY)
        if(NOT "${_d}" IN_LIST _added_subdirs)
            list(APPEND _added_subdirs "${_d}")
            list(APPEND _libpaths "/LIBPATH:\"${_d}\"")
        endif()
    endforeach()

    set(_link_libs kernel32.lib user32.lib gdi32.lib advapi32.lib shell32.lib
        ole32.lib oleaut32.lib uuid.lib comdlg32.lib winspool.lib)
    # Auto-link x32bridge/x32dbg (relative paths from source dir via LIBPATH)
    list(APPEND _link_libs pluginsdk/x32bridge.lib pluginsdk/x32dbg.lib)
    foreach(_f ${_DP86_SDK_FILES})
        string(REPLACE "${CMAKE_CURRENT_SOURCE_DIR}/" "" _rel "${_f}")
        list(APPEND _link_libs "${_rel}")
    endforeach()
    set(_link_libs_target_deps)
    if(WDK_LINK_LIBS)
        foreach(_lib ${WDK_LINK_LIBS})
            if(TARGET "${_lib}")
                list(APPEND _link_libs_target_deps "${_lib}")
                list(APPEND _link_libs "$<TARGET_FILE:${_lib}>")
            else()
                list(APPEND _link_libs "${_lib}")
            endif()
        endforeach()
    endif()

    if(WDK_COMPILE_OPTIONS)
        list(APPEND _OPT ${WDK_COMPILE_OPTIONS})
    endif()
    list(APPEND _OPT /utf-8)

    set(_all_objs)
    set(_src_idx 0)

    # Compile each source individually (.cpp/.c with cl.exe, .asm with ml.exe)
    foreach(_src ${WDK_SOURCES})
        get_filename_component(_abs "${_src}" ABSOLUTE)
        get_filename_component(_ext "${_src}" EXT)
        if(_ext STREQUAL ".rc")
            continue()
        endif()
        set(_obj "${_obj_dir}/${_target}_src${_src_idx}.obj")
        list(APPEND _all_objs "${_obj}")
        if(_ext STREQUAL ".asm")
            add_custom_command(
                OUTPUT "${_obj}"
                COMMAND "${X86_ML}" /nologo /c /Fo"${_obj}" "${_abs}"
                DEPENDS "${_abs}"
                COMMENT "Assembling x86: ${_target}_src${_src_idx}"
            )
        else()
            add_custom_command(
                OUTPUT "${_obj}"
                COMMAND "${X86_CL}" /utf-8 /nologo /c "${_abs}" /Fo"${_obj}" ${_inc_flags} ${_def_flags} ${_CRT} ${_OPT}
                DEPENDS "${_abs}"
                COMMENT "Compiling x86: ${_target}_src${_src_idx}"
            )
        endif()
        math(EXPR _src_idx "${_src_idx} + 1")
    endforeach()

    # ---- Compile .rc files individually ----
    foreach(_src ${WDK_SOURCES})
        get_filename_component(_abs "${_src}" ABSOLUTE)
        get_filename_component(_name "${_src}" NAME_WE)
        get_filename_component(_ext "${_src}" EXT)
        if(_ext STREQUAL ".rc")
            set(_res "${_obj_dir}/${_name}.res")
            list(APPEND _all_objs "${_res}")
            add_custom_command(
                OUTPUT "${_res}"
                COMMAND "${X86_RC}" /nologo /Fo"${_res}" ${_inc_flags} "${_abs}"
                DEPENDS "${_abs}"
                COMMENT "Compiling x86 resource: ${_name}"
            )
        endif()
    endforeach()

    add_custom_command(
        OUTPUT "${_output}" "${_implib}"
        COMMAND "${X86_LINK}" /nologo /DLL /OUT:"${_output}" /IMPLIB:"${_implib}" ${_all_objs} /MACHINE:X86 ${_libpaths} ${_link_libs}
        DEPENDS ${_all_objs} ${_link_libs_target_deps}
        COMMENT "Linking x86: ${_target}.dp32"
    )

    add_custom_target(${_target_internal} ALL DEPENDS "${_output}")

    if(X64DBG_X32_DIR)
        add_custom_command(TARGET ${_target_internal} POST_BUILD
            COMMAND ${CMAKE_COMMAND} -E make_directory "${X64DBG_X32_DIR}/plugins"
            COMMAND ${CMAKE_COMMAND} -E copy "${_output}" "${X64DBG_X32_DIR}/plugins/${_target}.dp32"
            COMMAND powershell -NoLogo -NoProfile -Command
                "if(Test-Path '${X64DBG_X32_DIR}/plugins/${_target}.dp32'){exit 0}else{exit 1}"
            COMMAND ${CMAKE_COMMAND} -E echo "[OK] ${_target}.dp32 copied to ${X64DBG_X32_DIR}/plugins/"
        )
    endif()
endfunction()

# ---- Qt-Specific Functions (强制 Release /MT，适配 Qt6 静态库) ----
# um_qt_exe(target)   — Qt EXE x64，不管外部的 CMAKE_BUILD_TYPE 都强制 /MT
# um_qt_dll(target)   — Qt DLL x64，不管外部的 CMAKE_BUILD_TYPE 都强制 /MT
# um_qt_exe_x86(target) — Qt EXE x86，强制 /MT
# um_qt_dll_x86(target) — Qt DLL x86，强制 /MT

function(um_qt_exe _target)
    cmake_parse_arguments(WDK "NOAUTO" "SUBSYSTEM;WINVER;NTDDI_VERSION" "SOURCES;INCLUDES;DEFINES;LIBS;COMPILE_OPTIONS" ${ARGN})

    if(NOT WDK_SUBSYSTEM)
        set(WDK_SUBSYSTEM "CONSOLE")
    endif()

    if(WDK_SOURCES)
        add_executable(${_target} ${WDK_SOURCES})
    else()
        add_executable(${_target} ${WDK_UNPARSED_ARGUMENTS})
    endif()

    string(TOUPPER "${WDK_SUBSYSTEM}" _subsystem_upper)

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE")
    # Qt 静态库只有 Release 版，强制 /MT 不管外部 CMAKE_BUILD_TYPE
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded")

    cmake_policy(SET CMP0117 NEW)
    set_target_properties(${_target} PROPERTIES
        CXX_STANDARD 17
        CXX_STANDARD_REQUIRED ON
        CXX_EXTENSIONS OFF
    )
    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8>)
    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
    target_include_directories(${_target} PRIVATE ${QT_INCLUDE_DIRS} ${CMAKE_CURRENT_SOURCE_DIR})
    target_compile_definitions(${_target} PRIVATE ${QT_COMPILE_DEFINITIONS})
    target_compile_options(${_target} PRIVATE ${QT_COMPILE_OPTIONS})
    target_link_libraries(${_target} ${QT_LINK_LIBRARIES})

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()
    if(WDK_DEFINES)
        target_compile_definitions(${_target} PRIVATE ${WDK_DEFINES})
    endif()
    if(WDK_COMPILE_OPTIONS)
        target_compile_options(${_target} PRIVATE ${WDK_COMPILE_OPTIONS})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})
    if(WDK_INCLUDES)
        target_include_directories(${_target} PRIVATE ${WDK_INCLUDES})
    endif()

    if(WDK_NOAUTO)
        if(WDK_LIBS)
            target_link_libraries(${_target} ${WDK_LIBS})
        endif()
    else()
        target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS})
        if(WDK_LIBS)
            target_link_libraries(${_target} ${WDK_LIBS})
        endif()
    endif()

    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()

    if(_subsystem_upper STREQUAL "CONSOLE" OR _subsystem_upper STREQUAL "WINCON")
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:CONSOLE")
    elseif(_subsystem_upper STREQUAL "WINDOWS" OR _subsystem_upper STREQUAL "WIN")
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:WINDOWS")
    endif()
endfunction()

function(um_qt_dll _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "SOURCES;INCLUDES;DEFINES;LIBS;COMPILE_OPTIONS" ${ARGN})

    if(WDK_SOURCES)
        add_library(${_target} SHARED ${WDK_SOURCES})
    else()
        add_library(${_target} SHARED ${WDK_UNPARSED_ARGUMENTS})
    endif()

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;_USRDLL;_WINDLL"
    )
    # Qt 静态库只有 Release 版，强制 /MT
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded")

    cmake_policy(SET CMP0117 NEW)
    set_target_properties(${_target} PROPERTIES
        CXX_STANDARD 17
        CXX_STANDARD_REQUIRED ON
        CXX_EXTENSIONS OFF
    )
    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8>)
    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
    target_include_directories(${_target} PRIVATE ${QT_INCLUDE_DIRS} ${CMAKE_CURRENT_SOURCE_DIR})
    target_compile_definitions(${_target} PRIVATE ${QT_COMPILE_DEFINITIONS})
    target_compile_options(${_target} PRIVATE ${QT_COMPILE_OPTIONS})
    target_link_libraries(${_target} ${QT_LINK_LIBRARIES})

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()
    if(WDK_DEFINES)
        target_compile_definitions(${_target} PRIVATE ${WDK_DEFINES})
    endif()
    if(WDK_COMPILE_OPTIONS)
        target_compile_options(${_target} PRIVATE ${WDK_COMPILE_OPTIONS})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})
    if(WDK_INCLUDES)
        target_include_directories(${_target} PRIVATE ${WDK_INCLUDES})
    endif()
    target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS})
    if(WDK_LIBS)
        target_link_libraries(${_target} ${WDK_LIBS})
    endif()
    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()
endfunction()

function(um_qt_exe_x86 _target)
    cmake_parse_arguments(WDK "" "SUBSYSTEM;WINVER;NTDDI_VERSION" "SOURCES;INCLUDE_DIRS;LINK_DIRS;LINK_LIBS;DEFINITIONS" ${ARGN})

    if(NOT WDK_SUBSYSTEM)
        set(WDK_SUBSYSTEM "CONSOLE")
    endif()
    string(TOUPPER "${WDK_SUBSYSTEM}" _subsystem_upper)

    if(CMAKE_C_COMPILER STREQUAL X86_CL)
        # Native mode — Qt 目标强制 /MT
        add_executable(${_target} ${WDK_SOURCES})
        set_target_properties(${_target} PROPERTIES
            COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;DWIN32;_WINDOWS${WDK_DEFINITIONS}"
            MSVC_RUNTIME_LIBRARY "MultiThreaded"
        )
        if(_subsystem_upper STREQUAL "CONSOLE" OR _subsystem_upper STREQUAL "WINCON")
            set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:CONSOLE /MACHINE:X86")
        else()
            set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:WINDOWS /MACHINE:X86")
        endif()
        target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
        target_include_directories(${_target} PRIVATE ${QT_INCLUDE_DIRS} ${CMAKE_CURRENT_SOURCE_DIR})
        target_compile_definitions(${_target} PRIVATE ${QT_COMPILE_DEFINITIONS})
        target_compile_options(${_target} PRIVATE ${QT_COMPILE_OPTIONS})
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
        foreach(_lib_dir ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
        endforeach()
        target_link_libraries(${_target} ${QT_LINK_LIBRARIES})
        target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS} ${WDK_LINK_LIBS})
    else()
        # Custom command mode — Qt 目标强制 /MT
        set(_output "${CMAKE_CURRENT_BINARY_DIR}/${_target}.exe")
        set(_obj_dir "${CMAKE_CURRENT_BINARY_DIR}/${_target}_objs")
        file(MAKE_DIRECTORY "${_obj_dir}")

        set(_CRT "/MT")
        set(_OPT "/O2" "/Ob2" "/DNDEBUG")

        set(_inc_flags)
        foreach(_d ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
            list(APPEND _inc_flags "/I\"${_d}\"")
        endforeach()
        foreach(_d ${QT_INCLUDE_DIRS})
            list(APPEND _inc_flags "/I\"${_d}\"")
        endforeach()
        list(APPEND _inc_flags "/I\"${CMAKE_CURRENT_SOURCE_DIR}\"")

        set(_def_flags "/D_WIN32_WINNT=${WDK_WINVER}" "/DWIN32" "/D_WINDOWS" "/DUNICODE" "/D_UNICODE")
        foreach(_def ${WDK_DEFINITIONS})
            list(APPEND _def_flags "/D${_def}")
        endforeach()
        foreach(_def ${QT_COMPILE_DEFINITIONS})
            list(APPEND _def_flags "/D${_def}")
        endforeach()

        set(_libpaths)
        foreach(_d ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            list(APPEND _libpaths "/LIBPATH:\"${_d}\"")
        endforeach()

        if(_subsystem_upper STREQUAL "CONSOLE" OR _subsystem_upper STREQUAL "WINCON")
            set(_subsys "/SUBSYSTEM:CONSOLE")
        else()
            set(_subsys "/SUBSYSTEM:WINDOWS")
        endif()

        set(_link_libs ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS} ${WDK_LINK_LIBS} ${QT_LINK_LIBRARIES})

        set(_all_objs "")
        set(_src_idx 0)

        foreach(_src ${WDK_SOURCES})
            get_filename_component(_abs "${_src}" ABSOLUTE)
            get_filename_component(_ext "${_src}" EXT)
            set(_obj "${_obj_dir}/${_target}_src${_src_idx}.obj")
            list(APPEND _all_objs "${_obj}")
            if(_ext STREQUAL ".asm")
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_ML}" /nologo /c /Fo"${_obj}" "${_abs}"
                    DEPENDS "${_abs}"
                    COMMENT "Assembling x86 Qt: ${_target}_src${_src_idx}"
                )
            else()
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_CL}" /utf-8 /nologo /c "${_abs}" /Fo"${_obj}" ${_inc_flags} ${_def_flags} ${_CRT} ${_OPT}
                    DEPENDS "${_abs}"
                    COMMENT "Compiling x86 Qt: ${_target}_src${_src_idx}"
                )
            endif()
            math(EXPR _src_idx "${_src_idx} + 1")
        endforeach()

        add_custom_command(
            OUTPUT "${_output}"
            COMMAND "${X86_LINK}" /nologo /OUT:"${_output}" ${_all_objs} ${_subsys} /MACHINE:X86 ${_libpaths} ${_link_libs}
            DEPENDS ${_all_objs}
            COMMENT "Linking x86 Qt: ${_target}.exe"
        )

        add_custom_target(${_target} ALL DEPENDS "${_output}")
    endif()
endfunction()

function(um_qt_dll_x86 _target)
    cmake_parse_arguments(WDK "" "SUFFIX" "SOURCES;INCLUDE_DIRS;LINK_DIRS;LINK_LIBS;DEFINITIONS" ${ARGN})

    if(NOT WDK_SUFFIX)
        set(WDK_SUFFIX ".dll")
    endif()

    if(CMAKE_C_COMPILER STREQUAL X86_CL)
        # Native mode — Qt 目标强制 /MT
        add_library(${_target} SHARED ${WDK_SOURCES})
        set_target_properties(${_target} PROPERTIES
            COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;DWIN32;_WINDOWS;_USRDLL;_WINDLL${WDK_DEFINITIONS}"
            MSVC_RUNTIME_LIBRARY "MultiThreaded"
            SUFFIX "${WDK_SUFFIX}"
        )
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/MACHINE:X86")
        target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
        target_include_directories(${_target} PRIVATE ${QT_INCLUDE_DIRS} ${CMAKE_CURRENT_SOURCE_DIR})
        target_compile_definitions(${_target} PRIVATE ${QT_COMPILE_DEFINITIONS})
        target_compile_options(${_target} PRIVATE ${QT_COMPILE_OPTIONS})
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
        foreach(_lib_dir ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
        endforeach()
        target_link_libraries(${_target} ${QT_LINK_LIBRARIES})
        target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS} ${WDK_LINK_LIBS})
        target_link_options(${_target} PRIVATE "/DEF:${_def_file}")
    else()
        # Custom command mode — Qt 目标强制 /MT
        set(_output "${CMAKE_CURRENT_BINARY_DIR}/${_target}${WDK_SUFFIX}")
        set(_obj_dir "${CMAKE_CURRENT_BINARY_DIR}/${_target}_objs")
        file(MAKE_DIRECTORY "${_obj_dir}")

        set(_CRT "/MT")
        set(_OPT "/O2" "/Ob2" "/DNDEBUG")

        set(_inc_flags)
        foreach(_d ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
            list(APPEND _inc_flags "/I\"${_d}\"")
        endforeach()
        foreach(_d ${QT_INCLUDE_DIRS})
            list(APPEND _inc_flags "/I\"${_d}\"")
        endforeach()
        list(APPEND _inc_flags "/I\"${CMAKE_CURRENT_SOURCE_DIR}\"")

        set(_def_flags "/D_WIN32_WINNT=${WDK_WINVER}" "/DWIN32" "/D_WINDOWS" "/D_USRDLL" "/D_WINDLL" "/DUNICODE" "/D_UNICODE")
        foreach(_def ${WDK_DEFINITIONS})
            list(APPEND _def_flags "/D${_def}")
        endforeach()
        foreach(_def ${QT_COMPILE_DEFINITIONS})
            list(APPEND _def_flags "/D${_def}")
        endforeach()

        set(_libpaths)
        foreach(_d ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            list(APPEND _libpaths "/LIBPATH:\"${_d}\"")
        endforeach()

        set(_link_libs ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS} ${WDK_LINK_LIBS} ${QT_LINK_LIBRARIES})

        set(_all_objs "")
        set(_src_idx 0)

        foreach(_src ${WDK_SOURCES})
            get_filename_component(_abs "${_src}" ABSOLUTE)
            get_filename_component(_ext "${_src}" EXT)
            set(_obj "${_obj_dir}/${_target}_src${_src_idx}.obj")
            list(APPEND _all_objs "${_obj}")
            if(_ext STREQUAL ".asm")
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_ML}" /nologo /c /Fo"${_obj}" "${_abs}"
                    DEPENDS "${_abs}"
                    COMMENT "Assembling x86 Qt DLL: ${_target}_src${_src_idx}"
                )
            else()
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_CL}" /utf-8 /nologo /c "${_abs}" /Fo"${_obj}" ${_inc_flags} ${_def_flags} ${_CRT} ${_OPT}
                    DEPENDS "${_abs}"
                    COMMENT "Compiling x86 Qt DLL: ${_target}_src${_src_idx}"
                )
            endif()
            math(EXPR _src_idx "${_src_idx} + 1")
        endforeach()

        add_custom_command(
            OUTPUT "${_output}"
            COMMAND "${X86_LINK}" /nologo /DLL /OUT:"${_output}" ${_all_objs} ${_subsys} /MACHINE:X86 ${_libpaths} ${_link_libs}
            DEPENDS ${_all_objs}
            COMMENT "Linking x86 Qt DLL: ${_target}${WDK_SUFFIX}"
        )

        add_custom_target(${_target} ALL DEPENDS "${_output}")
    endif()
endfunction()

# ---- x86 Functions ----

function(um_exe_x86 _target)
    cmake_parse_arguments(WDK "" "SUBSYSTEM;WINVER;NTDDI_VERSION" "SOURCES;INCLUDE_DIRS;LINK_DIRS;LINK_LIBS;DEFINITIONS" ${ARGN})

    if(NOT WDK_SUBSYSTEM)
        set(WDK_SUBSYSTEM "CONSOLE")
    endif()
    string(TOUPPER "${WDK_SUBSYSTEM}" _subsystem_upper)

    if(CMAKE_C_COMPILER STREQUAL X86_CL)
        # Native mode
        add_executable(${_target} ${WDK_SOURCES})
        set_target_properties(${_target} PROPERTIES
            COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;DWIN32;_WINDOWS${WDK_DEFINITIONS}"
            MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>"
        )
        if(_subsystem_upper STREQUAL "CONSOLE" OR _subsystem_upper STREQUAL "WINCON")
            set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:CONSOLE /MACHINE:X86")
        else()
            set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:WINDOWS /MACHINE:X86")
        endif()
        target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
        foreach(_lib_dir ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
        endforeach()
        target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS} ${WDK_LINK_LIBS})
    else()
        # Custom command mode
        set(_output "${CMAKE_CURRENT_BINARY_DIR}/${_target}.exe")
        set(_obj_dir "${CMAKE_CURRENT_BINARY_DIR}/${_target}_objs")
        file(MAKE_DIRECTORY "${_obj_dir}")

        if(CMAKE_BUILD_TYPE STREQUAL "Debug")
            set(_CRT "/MTd")
            set(_OPT "/Od" "/Ob0" "/DDEBUG")
        else()
            set(_CRT "/MT")
            set(_OPT "/O2" "/Ob2" "/DNDEBUG")
        endif()

        set(_inc_flags)
        foreach(_d ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
            list(APPEND _inc_flags "/I\"${_d}\"")
        endforeach()

        set(_def_flags "/D_WIN32_WINNT=${WDK_WINVER}" "/DWIN32" "/D_WINDOWS" "/DUNICODE" "/D_UNICODE")
        foreach(_def ${WDK_DEFINITIONS})
            list(APPEND _def_flags "/D${_def}")
        endforeach()

        set(_libpaths)
        foreach(_d ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            list(APPEND _libpaths "/LIBPATH:\"${_d}\"")
        endforeach()

        if(_subsystem_upper STREQUAL "CONSOLE" OR _subsystem_upper STREQUAL "WINCON")
            set(_subsys "/SUBSYSTEM:CONSOLE")
        else()
            set(_subsys "/SUBSYSTEM:WINDOWS")
        endif()

        set(_link_libs ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS} ${WDK_LINK_LIBS})

        set(_all_objs "")
        set(_src_idx 0)

        # Compile each source individually (.cpp/.c with cl.exe, .asm with ml.exe)
        foreach(_src ${WDK_SOURCES})
            get_filename_component(_abs "${_src}" ABSOLUTE)
            get_filename_component(_ext "${_src}" EXT)
            set(_obj "${_obj_dir}/${_target}_src${_src_idx}.obj")
            list(APPEND _all_objs "${_obj}")
            if(_ext STREQUAL ".asm")
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_ML}" /nologo /c /Fo"${_obj}" "${_abs}"
                    DEPENDS "${_abs}"
                    COMMENT "Assembling x86: ${_target}_src${_src_idx}"
                )
            else()
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_CL}" /utf-8 /nologo /c "${_abs}" /Fo"${_obj}" ${_inc_flags} ${_def_flags} ${_CRT} ${_OPT}
                    DEPENDS "${_abs}"
                    COMMENT "Assembling x86: ${_target}_src${_src_idx}"
                )
            endif()
            math(EXPR _src_idx "${_src_idx} + 1")
        endforeach()

        add_custom_command(
            OUTPUT "${_output}"
            COMMAND "${X86_LINK}" /nologo /OUT:"${_output}" ${_all_objs} ${_subsys} /MACHINE:X86 ${_libpaths} ${_link_libs}
            DEPENDS ${_all_objs}
            COMMENT "Linking x86: ${_target}.exe"
        )

        add_custom_target(${_target} ALL DEPENDS "${_output}")
    endif()
endfunction()

function(um_dll_x86 _target)
    cmake_parse_arguments(WDK "" "SUFFIX" "SOURCES;INCLUDE_DIRS;LINK_DIRS;LINK_LIBS;DEFINITIONS" ${ARGN})

    if(NOT WDK_SUFFIX)
        set(WDK_SUFFIX ".dll")
    endif()

    if(CMAKE_C_COMPILER STREQUAL X86_CL)
        # Native mode
        add_library(${_target} SHARED ${WDK_SOURCES})
        set_target_properties(${_target} PROPERTIES
            SUFFIX "${WDK_SUFFIX}"
            COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;DWIN32;_WINDOWS;_USRDLL;_WINDLL${WDK_DEFINITIONS}"
            MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>"
            LINK_FLAGS "/MACHINE:X86"
        )
        target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/EHsc /std:c++latest>)
        foreach(_lib_dir ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
        endforeach()
        target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS} ${WDK_LINK_LIBS})
    else()
        # Custom command mode
        set(_output "${CMAKE_CURRENT_BINARY_DIR}/${_target}${WDK_SUFFIX}")
        set(_implib "${CMAKE_CURRENT_BINARY_DIR}/${_target}.lib")
        set(_obj_dir "${CMAKE_CURRENT_BINARY_DIR}/${_target}_objs")
        file(MAKE_DIRECTORY "${_obj_dir}")

        if(CMAKE_BUILD_TYPE STREQUAL "Debug")
            set(_CRT "/MTd")
            set(_OPT "/Od" "/Ob0" "/DDEBUG")
        else()
            set(_CRT "/MT")
            set(_OPT "/O2" "/Ob2" "/DNDEBUG")
        endif()

        set(_inc_flags)
        foreach(_d ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
            list(APPEND _inc_flags "/I\"${_d}\"")
        endforeach()

        set(_def_flags "/D_WIN32_WINNT=${WDK_WINVER}" "/DWIN32" "/D_WINDOWS" "/D_USRDLL" "/D_WINDLL" "/DUNICODE" "/D_UNICODE")
        foreach(_def ${WDK_DEFINITIONS})
            list(APPEND _def_flags "/D${_def}")
        endforeach()

        set(_libpaths)
        foreach(_d ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            list(APPEND _libpaths "/LIBPATH:\"${_d}\"")
        endforeach()

        set(_link_libs kernel32.lib user32.lib ${WDK_LINK_LIBS})

        set(_all_objs "")
        set(_src_idx 0)

        # Compile each source individually (.cpp/.c with cl.exe, .asm with ml.exe)
        foreach(_src ${WDK_SOURCES})
            get_filename_component(_abs "${_src}" ABSOLUTE)
            get_filename_component(_ext "${_src}" EXT)
            if(_ext STREQUAL ".rc")
                continue()
            endif()
            set(_obj "${_obj_dir}/${_target}_src${_src_idx}.obj")
            list(APPEND _all_objs "${_obj}")
            if(_ext STREQUAL ".asm")
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_ML}" /nologo /c /Fo"${_obj}" "${_abs}"
                    DEPENDS "${_abs}"
                    COMMENT "Assembling x86: ${_target}_src${_src_idx}"
                )
            else()
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_CL}" /utf-8 /nologo /c "${_abs}" /Fo"${_obj}" ${_inc_flags} ${_def_flags} ${_CRT} ${_OPT}
                    DEPENDS "${_abs}"
                    COMMENT "Compiling x86: ${_target}_src${_src_idx}"
                )
            endif()
            math(EXPR _src_idx "${_src_idx} + 1")
        endforeach()

        # Compile .rc files individually
        foreach(_src ${WDK_SOURCES})
            get_filename_component(_abs "${_src}" ABSOLUTE)
            get_filename_component(_name "${_src}" NAME_WE)
            get_filename_component(_ext "${_src}" EXT)
            if(_ext STREQUAL ".rc")
                set(_res "${_obj_dir}/${_name}.res")
                list(APPEND _all_objs "${_res}")
                add_custom_command(
                    OUTPUT "${_res}"
                    COMMAND "${X86_RC}" /nologo /Fo"${_res}" ${_inc_flags} "${_abs}"
                    DEPENDS "${_abs}"
                    COMMENT "Compiling x86 resource: ${_name}"
                )
            endif()
        endforeach()

        add_custom_command(
            OUTPUT "${_output}" "${_implib}"
            COMMAND "${X86_LINK}" /nologo /DLL /OUT:"${_output}" /IMPLIB:"${_implib}" ${_all_objs} /MACHINE:X86 ${_libpaths} ${_link_libs}
            DEPENDS ${_all_objs}
            COMMENT "Linking x86: ${_target}.dll"
        )

        add_custom_target(${_target} ALL DEPENDS "${_output}")
    endif()
endfunction()

function(um_lib_x86 _target)
    cmake_parse_arguments(WDK "" "" "SOURCES;INCLUDE_DIRS;LINK_DIRS;DEFINITIONS" ${ARGN})

    if(CMAKE_C_COMPILER STREQUAL X86_CL)
        # Native mode
        add_library(${_target} STATIC ${WDK_SOURCES})
        set_target_properties(${_target} PROPERTIES
            COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;DWIN32;_WINDOWS${WDK_DEFINITIONS}"
            MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>"
        )
        target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
        target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/std:c++latest>)
        foreach(_lib_dir ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
        endforeach()
    else()
        # Custom command mode
        set(_output "${CMAKE_CURRENT_BINARY_DIR}/${_target}.lib")
        set(_obj_dir "${CMAKE_CURRENT_BINARY_DIR}/${_target}_objs")
        file(MAKE_DIRECTORY "${_obj_dir}")

        if(CMAKE_BUILD_TYPE STREQUAL "Debug")
            set(_CRT "/MTd")
            set(_OPT "/Od" "/Ob0" "/DDEBUG")
        else()
            set(_CRT "/MT")
            set(_OPT "/O2" "/Ob2" "/DNDEBUG")
        endif()

        set(_inc_flags)
        foreach(_d ${WDK_UM_INCLUDE_DIRS_X86} ${WDK_INCLUDE_DIRS})
            list(APPEND _inc_flags "/I\"${_d}\"")
        endforeach()

        set(_def_flags "/D_WIN32_WINNT=${WDK_WINVER}" "/DWIN32" "/D_WINDOWS" "/DUNICODE" "/D_UNICODE")
        foreach(_def ${WDK_DEFINITIONS})
            list(APPEND _def_flags "/D${_def}")
        endforeach()

        set(_libpaths)
        foreach(_d ${WDK_UM_LIB_DIRS_X86} ${WDK_LINK_DIRS})
            list(APPEND _libpaths "/LIBPATH:\"${_d}\"")
        endforeach()

        set(_all_objs "")
        set(_src_idx 0)

        # Compile each source individually (.cpp/.c with cl.exe, .asm with ml.exe)
        foreach(_src ${WDK_SOURCES})
            get_filename_component(_abs "${_src}" ABSOLUTE)
            get_filename_component(_ext "${_src}" EXT)
            set(_obj "${_obj_dir}/${_target}_src${_src_idx}.obj")
            list(APPEND _all_objs "${_obj}")
            if(_ext STREQUAL ".asm")
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_ML}" /nologo /c /Fo"${_obj}" "${_abs}"
                    DEPENDS "${_abs}"
                    COMMENT "Assembling x86: ${_target}_src${_src_idx}"
                )
            else()
                add_custom_command(
                    OUTPUT "${_obj}"
                    COMMAND "${X86_CL}" /utf-8 /nologo /c "${_abs}" /Fo"${_obj}" ${_inc_flags} ${_def_flags} ${_CRT} ${_OPT}
                    DEPENDS "${_abs}"
                    COMMENT "Compiling x86: ${_target}_src${_src_idx}"
                )
            endif()
            math(EXPR _src_idx "${_src_idx} + 1")
        endforeach()

        add_custom_command(
            OUTPUT "${_output}"
            COMMAND "${X86_LINK}" /nologo /LIB /OUT:"${_output}" ${_all_objs}
            DEPENDS ${_all_objs}
            COMMENT "Linking x86: ${_target}.lib"
        )

        add_custom_target(${_target} ALL DEPENDS "${_output}")
    endif()
endfunction()

function(um_exe_mfc _target)
    cmake_parse_arguments(WDK "" "" "SOURCES;INCLUDE_DIRS;LINK_DIRS;LINK_LIBS;DEFINITIONS" ${ARGN})

    add_executable(${_target} ${WDK_SOURCES})

    target_compile_options(${_target} PRIVATE $<$<COMPILE_LANGUAGE:C,CXX>:/utf-8> $<$<COMPILE_LANGUAGE:C>:/TC> $<$<COMPILE_LANGUAGE:CXX>:/EHsc /std:c++latest>)
    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "_WIN32_WINNT=${WDK_WINVER};UNICODE;_UNICODE;_AFX_STATIC${WDK_DEFINITIONS}")
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")

    target_include_directories(${_target} PRIVATE
        ${WDK_UM_INCLUDE_DIRS}
        ${MFC_INCLUDE_DIR}
        ${WDK_INCLUDE_DIRS})
    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()
    target_link_options(${_target} PRIVATE "/LIBPATH:${MFC_LIB_DIR_X64}")
    if(WDK_LINK_DIRS)
        foreach(_lib_dir ${WDK_LINK_DIRS})
            target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
        endforeach()
    endif()

    set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:WINDOWS /MACHINE:X64")
    target_link_libraries(${_target} ${WDK_UM_SDK_LIBS} ${WDK_UM_EXTRA_LIBS} ${WDK_LINK_LIBS})
endfunction()

function(um_exe_mfc_x86 _target)
    cmake_parse_arguments(WDK "" "" "SOURCES;INCLUDE_DIRS;LINK_DIRS;LINK_LIBS;DEFINITIONS" ${ARGN})

    set(_output "${CMAKE_CURRENT_BINARY_DIR}/${_target}.exe")
    set(_obj_dir "${CMAKE_CURRENT_BINARY_DIR}/${_target}_objs")
    file(MAKE_DIRECTORY "${_obj_dir}")

    set(_cc "${X86_CL}")
    set(_link "${X86_LINK}")

    if(CMAKE_BUILD_TYPE STREQUAL "Debug")
        set(_CRT "/MTd")
        set(_OPT "/Od /Ob0 /DDEBUG")
    else()
        set(_CRT "/MT")
        set(_OPT "/O2 /Ob2 /DNDEBUG")
    endif()

    set(_inc_flags)
    foreach(_d ${WDK_UM_INCLUDE_DIRS_X86} ${MFC_INCLUDE_DIR} ${WDK_INCLUDE_DIRS})
        list(APPEND _inc_flags "/I\"${_d}\"")
    endforeach()

    set(_def_flags "/D_WIN32_WINNT=${WDK_WINVER}" "/DWIN32" "/D_WINDOWS" "/D_AFX_STATIC" "/DUNICODE" "/D_UNICODE")
    foreach(_def ${WDK_DEFINITIONS})
        list(APPEND _def_flags "/D${_def}")
    endforeach()

    set(_libpaths)
    foreach(_d ${WDK_UM_LIB_DIRS_X86} ${MFC_LIB_DIR_X86} ${WDK_LINK_DIRS})
        list(APPEND _libpaths "/LIBPATH:\"${_d}\"")
    endforeach()

    set(_link_libs kernel32.lib user32.lib ${WDK_LINK_LIBS})

    set(_all_objs "")
    set(_src_idx 0)

    # Compile each source individually (.cpp/.c with cl.exe, .asm with ml.exe)
    foreach(_src ${WDK_SOURCES})
        get_filename_component(_abs "${_src}" ABSOLUTE)
        get_filename_component(_ext "${_src}" EXT)
        set(_obj "${_obj_dir}/${_target}_src${_src_idx}.obj")
        list(APPEND _all_objs "${_obj}")
        if(_ext STREQUAL ".asm")
            add_custom_command(
                OUTPUT "${_obj}"
                COMMAND "${X86_ML}" /nologo /c /Fo"${_obj}" "${_abs}"
                DEPENDS "${_abs}"
                COMMENT "Assembling x86: ${_target}_src${_src_idx}"
            )
        else()
            add_custom_command(
                OUTPUT "${_obj}"
                COMMAND "${_cc}" /utf-8 /nologo /c "${_abs}" /Fo"${_obj}" ${_inc_flags} ${_def_flags} ${_CRT} ${_OPT}
                DEPENDS "${_abs}"
                COMMENT "Compiling x86: ${_target}_src${_src_idx}"
            )
        endif()
        math(EXPR _src_idx "${_src_idx} + 1")
    endforeach()

    add_custom_command(
        OUTPUT "${_output}"
        COMMAND "${_link}" /nologo /OUT:"${_output}" ${_all_objs} /SUBSYSTEM:WINDOWS /MACHINE:X86 ${_libpaths} ${_link_libs}
        DEPENDS ${_all_objs}
        COMMENT "Linking x86 MFC: ${_target}.exe"
    )

    add_custom_target(${_target} ALL DEPENDS "${_output}")
endfunction()

# 自动引入 unity.cmake（collect_sources / generate_unity）
include("${CMAKE_CURRENT_LIST_DIR}/unity.cmake" OPTIONAL)
`)

	return os.WriteFile(outputPath, []byte(b.String()), 0644)
}
