package main

import (
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

const powershell = "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"

type ewdkCommonEnv struct {
	WDKContentRoot               string
	WindowsTargetPlatformVersion string
	VCToolsInstallDir            string
	WDKBinRoot                   string
	DiaRoot                      string
	VSINSTALLDIR                 string
	BuildLabSetupRoot            string
	CC                           string
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
	IncludeDirs []string
	LibDirs     []string
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

	stream.CopyFile("ninja.exe", filepath.Join(cmake.BinDir, "ninja.exe"))

	ensureTestCertificate()

	envData := mylog.Check2(json.MarshalIndent(env, "", "  "))
	mylog.Check(os.WriteFile(cmake.EwdkEnvFile, envData, 0644))
	mylog.Success("Generated: ", cmake.EwdkEnvFile)

	mylog.Success("Environment ready. Run build.bat to start building.")
}

const isoPath = `D:\ux\examples\ewdk\EWDK_br_release_28000_251103-1709.iso`

const testSignCertName = "WDKTestCert"

func ensureTestCertificate() {
	script := fmt.Sprintf(`$cert = Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Where-Object { $_.Subject -match 'CN=%s' }; if ($cert) { Write-Host 'EXISTS' } else { Write-Host 'NOT_FOUND' }`, testSignCertName)
	output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
	if err != nil {
		fmt.Printf("  [WARN] check certificate: %v\n", err)
		return
	}
	if strings.TrimSpace(string(output)) == "EXISTS" {
		fmt.Printf("  [OK]   Test certificate '%s' already exists\n", testSignCertName)
		return
	}
	createScript := fmt.Sprintf(`New-SelfSignedCertificate -Type CodeSigningCert -Subject "CN=%s" -CertStoreLocation "Cert:\CurrentUser\My" | Out-Null; Write-Host 'CREATED'`, testSignCertName)
	out, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", createScript).CombinedOutput()
	if err != nil {
		fmt.Printf("  [WARN] create certificate: %v, output: %s\n", err, strings.TrimSpace(string(out)))
		return
	}
	fmt.Printf("  [OK]   Created test certificate '%s'\n", testSignCertName)
}

func resolveISOPath() string {
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		iso := filepath.Join(os.Getenv("TEMP"), "ewdk.iso")
		if !stream.FileExists(iso) {
			panic("CI env: ewdk.iso not found in TEMP")
		}
		return iso
	}
	if !stream.FileExists(isoPath) {
		panic("ISO file not found: " + isoPath)
	}
	return isoPath
}

func cleanGenerated() {
	files := []string{cmake.EwdkCmakeFile}
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

	cc := filepath.Join(result[EnvVCToolsInstallDir], "bin\\Hostx64\\x64\\cl.exe")
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
			filepath.Join(result[EnvWDKContentRoot], "Include", "wdf", "kmdf", "1.35"),
		},
		LibDirs: []string{
			filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "km", "x64"),
		},
	}

	um := ewdkUMEnv{
		IncludeDirs: append([]string{
			filepath.Join(pre, "shared"),
			filepath.Join(pre, "um"),
			ucrtInc,
		}, vcInc...),
		LibDirs: append([]string{umLib, ucrtLib}, vcLib...),
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

	b.WriteString("\n# ---- KM (Kernel-Mode) ----\n")
	writeList(&b, "WDK_KM_INCLUDE_DIRS", env.KM.IncludeDirs)
	writeList(&b, "WDK_KM_LIB_DIRS", env.KM.LibDirs)

	b.WriteString("\n# ---- UM (User-Mode) ----\n")
	writeList(&b, "WDK_UM_INCLUDE_DIRS", env.UM.IncludeDirs)
	writeList(&b, "WDK_UM_LIB_DIRS", env.UM.LibDirs)

	b.WriteString("\n# ---- Compiler ----\n")
	b.WriteString(fmt.Sprintf("set(CMAKE_C_COMPILER \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.CC)))
	b.WriteString(fmt.Sprintf("set(CMAKE_CXX_COMPILER \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.CC)))
	b.WriteString(fmt.Sprintf("set(CMAKE_RC_COMPILER \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.RC)))
	b.WriteString("set(CMAKE_C_COMPILER_WORKS 1 CACHE INTERNAL \"\")\n")
	b.WriteString("set(CMAKE_CXX_COMPILER_WORKS 1 CACHE INTERNAL \"\")\n")
	b.WriteString("set(CMAKE_C_STANDARD_LIBRARIES \"\")\n")
	b.WriteString("set(CMAKE_CXX_STANDARD_LIBRARIES \"\")\n")
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
file(WRITE ${KM_ADDITIONAL_FLAGS_FILE} "#pragma runtime_checks(\"suc\", off)")

set(KM_COMPILE_FLAGS
    "/Zp8"
    "/GF"
    "/GR-"
    "/Gz"
    "/kernel"
    "/FIwarning.h"
    "/FI${KM_ADDITIONAL_FLAGS_FILE}"
    "/Oi"
    )

set(KM_COMPILE_DEFINITIONS "WINNT=1;_AMD64_;AMD64")
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

# ---- KM/UM Functions ----

function(km_sys _target)
    cmake_parse_arguments(WDK "" "KMDF;WINVER;NTDDI_VERSION" "" ${ARGN})

    add_executable(${_target} ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES SUFFIX ".sys")
    set_target_properties(${_target} PROPERTIES COMPILE_OPTIONS "${KM_COMPILE_FLAGS}")
    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "${KM_COMPILE_DEFINITIONS};$<$<CONFIG:Debug>:${KM_COMPILE_DEFINITIONS_DEBUG}>;_WIN32_WINNT=${WDK_WINVER}"
        )
    set_target_properties(${_target} PROPERTIES LINK_FLAGS "${KM_LINK_FLAGS}")
    set_target_properties(${_target} PROPERTIES LINK_INTERFACE_LIBRARIES "")
    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_KM_INCLUDE_DIRS})

    target_link_libraries(${_target} WDK::NTOSKRNL WDK::HAL WDK::WMILIB)

    if(WDK::BUFFEROVERFLOWK)
        target_link_libraries(${_target} WDK::BUFFEROVERFLOWK)
    else()
        target_link_libraries(${_target} WDK::BUFFEROVERFLOWFASTFAILK)
    endif()

    if(DEFINED WDK_KMDF)
        target_include_directories(${_target} PRIVATE "${WDK_ROOT}/Include/wdf/kmdf/${WDK_KMDF}")
        target_link_libraries(${_target}
            "${WDK_ROOT}/Lib/wdf/kmdf/${WDK_PLATFORM}/${WDK_KMDF}/WdfDriverEntry.lib"
            "${WDK_ROOT}/Lib/wdf/kmdf/${WDK_PLATFORM}/${WDK_KMDF}/WdfLdr.lib"
            )
        set_property(TARGET ${_target} APPEND_STRING PROPERTY LINK_FLAGS "/ENTRY:FxDriverEntry")
    else()
        set_property(TARGET ${_target} APPEND_STRING PROPERTY LINK_FLAGS "/ENTRY:GsDriverEntry")
    endif()

    if(KM_TEST_SIGN AND DEFINED KM_SIGNTOOL_PATH)
        add_custom_command(TARGET ${_target} POST_BUILD
            COMMAND ${KM_SIGNTOOL_PATH} sign /fd SHA256 /s My /n ${KM_TEST_SIGN_NAME} /t http://timestamp.digicert.com $<TARGET_FILE:${_target}>
            WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
            COMMENT "Signing driver with test certificate: ${KM_TEST_SIGN_NAME}"
            VERBATIM
        )
    endif()
endfunction()

function(km_lib _target)
    cmake_parse_arguments(WDK "" "KMDF;WINVER;NTDDI_VERSION" "" ${ARGN})

    add_library(${_target} ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES COMPILE_OPTIONS "${KM_COMPILE_FLAGS}")
    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "${KM_COMPILE_DEFINITIONS};$<$<CONFIG:Debug>:${KM_COMPILE_DEFINITIONS_DEBUG};>_WIN32_WINNT=${WDK_WINVER}"
        )
    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_KM_INCLUDE_DIRS})

    if(DEFINED WDK_KMDF)
        target_include_directories(${_target} PRIVATE "${WDK_ROOT}/Include/wdf/kmdf/${WDK_KMDF}")
    endif()
endfunction()

function(um_exe _target)
    cmake_parse_arguments(WDK "" "SUBSYSTEM;WINVER;NTDDI_VERSION" "" ${ARGN})

    if(NOT WDK_SUBSYSTEM)
        set(WDK_SUBSYSTEM "CONSOLE")
    endif()

    add_executable(${_target} ${WDK_UNPARSED_ARGUMENTS})

    string(TOUPPER "${WDK_SUBSYSTEM}" _subsystem_upper)

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER}")
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})

    if(_subsystem_upper STREQUAL "CONSOLE" OR _subsystem_upper STREQUAL "WINCON")
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:CONSOLE")
    elseif(_subsystem_upper STREQUAL "WINDOWS" OR _subsystem_upper STREQUAL "WIN")
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:WINDOWS")
    endif()

    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()
    target_link_libraries(${_target} kernel32.lib user32.lib)
endfunction()

function(um_lib _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "" ${ARGN})

    add_library(${_target} ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER}")
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})
    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()
endfunction()

function(um_dll _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "" ${ARGN})

    add_library(${_target} SHARED ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "_WIN32_WINNT=${WDK_WINVER};_USRDLL;_WINDLL"
        )
    set_target_properties(${_target} PROPERTIES MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE ${WDK_UM_INCLUDE_DIRS})
    foreach(_lib_dir ${WDK_UM_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_lib_dir}")
    endforeach()
    target_link_libraries(${_target} kernel32.lib)
endfunction()
`)

	return os.WriteFile(outputPath, []byte(b.String()), 0644)
}
