Unicode true

####
## Vinta NSIS Windows Installer Configuration
####

!define WAILS_INSTALL_SCOPE "user"
!define REQUEST_EXECUTION_LEVEL "user"

!include "wails_tools.nsh"
!include "MUI2.nsh"
!include "FileFunc.nsh"
!include "LogicLib.nsh"

# Version information
!ifndef INFO_PRODUCTVERSION
    !define INFO_PRODUCTVERSION "0.1.0"
!endif

!ifndef INFO_PRODUCTVERSION_NUMERIC
    !define INFO_PRODUCTVERSION_NUMERIC "0.1.0.0"
!endif

VIProductVersion "${INFO_PRODUCTVERSION_NUMERIC}"
VIFileVersion    "${INFO_PRODUCTVERSION_NUMERIC}"

VIAddVersionKey "CompanyName"      "${INFO_COMPANYNAME}"
VIAddVersionKey "Publisher"        "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription"  "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"   "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"      "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"   "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"      "${INFO_PRODUCTNAME}"
VIAddVersionKey "OriginalFilename" "Vinta-Setup-x64.exe"

# Execution Level: Run as standard user (installs into %LOCALAPPDATA%\Programs\Vinta without elevated UAC warnings)
RequestExecutionLevel user

# Enable HiDPI support
ManifestDPIAware true

# UI Configuration
!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_NOAUTOCLOSE

# Installer Pages
!insertmacro MUI_PAGE_WELCOME

# Directory Selection Page with root folder guard
!define MUI_PAGE_CUSTOMFUNCTION_LEAVE DirectoryLeave
!insertmacro MUI_PAGE_DIRECTORY

# Component / Shortcut Selection Page
!define MUI_COMPONENTSPAGE_NODESC
!insertmacro MUI_PAGE_COMPONENTS

!insertmacro MUI_PAGE_INSTFILES

# Finish page with optional launch checkbox
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!define MUI_FINISHPAGE_RUN_TEXT "Launch ${INFO_PRODUCTNAME}"
!insertmacro MUI_PAGE_FINISH

# Uninstaller Pages
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\Vinta-Setup-x64.exe"

# Default install directory: per-user LocalAppData (or ProgramFiles if running elevated)
InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
InstallDirRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "InstallLocation"
ShowInstDetails show

Var /GLOBAL PathTail
Var /GLOBAL DetectedPath

# Function to ensure selected directory always ends with \Vinta subfolder
Function DirectoryLeave
    # If the executable already exists in the chosen folder, do not append subfolder
    ${If} ${FileExists} "$INSTDIR\${PRODUCT_EXECUTABLE}"
        Return
    ${EndIf}

    # Strip trailing backslash if present
    StrCpy $0 $INSTDIR "" -1
    ${If} $0 == "\"
        StrCpy $INSTDIR $INSTDIR -1
    ${EndIf}

    # Extract last 5 characters to check for "Vinta"
    StrCpy $PathTail $INSTDIR "" -5
    ${If} $PathTail != "${INFO_PRODUCTNAME}"
    ${AndIf} $PathTail != "vinta"
    ${AndIf} $PathTail != "VINTA"
        StrCpy $INSTDIR "$INSTDIR\${INFO_PRODUCTNAME}"
    ${EndIf}
FunctionEnd

# Macro to verify and extract install directory from registry string
!macro CheckRegistryPath REG_ROOT REG_KEY REG_VAL
    ${If} $DetectedPath == ""
        ReadRegStr $0 ${REG_ROOT} "${REG_KEY}" "${REG_VAL}"
        ${If} $0 != ""
            # Strip enclosing quotes if present
            StrCpy $1 $0 1 0
            ${If} $1 == '"'
                StrCpy $0 $0 "" 1
                StrCpy $1 $0 "" -1
                ${If} $1 == '"'
                    StrCpy $0 $0 -1
                ${EndIf}
            ${EndIf}

            # Check if directory directly contains executable
            ${If} ${FileExists} "$0\${PRODUCT_EXECUTABLE}"
                StrCpy $DetectedPath $0
            ${Else}
                # If path pointed to an exe file, check its parent directory
                ${GetParent} "$0" $2
                ${If} $2 != ""
                ${AndIf} ${FileExists} "$2\${PRODUCT_EXECUTABLE}"
                    StrCpy $DetectedPath $2
                ${EndIf}
            ${EndIf}
        ${EndIf}
    ${EndIf}
!macroend

Function .onInit
    !insertmacro wails.checkArchitecture

    # 1. If INSTDIR was specified via /D= or already contains executable, keep it
    ${If} $INSTDIR != ""
    ${AndIf} ${FileExists} "$INSTDIR\${PRODUCT_EXECUTABLE}"
        Return
    ${EndIf}

    StrCpy $DetectedPath ""

    # 2. Check 64-bit registry entries in HKCU and HKLM
    SetRegView 64
    !insertmacro CheckRegistryPath HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallPath"
    !insertmacro CheckRegistryPath HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "DisplayIcon"
    !insertmacro CheckRegistryPath HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "UninstallString"
    !insertmacro CheckRegistryPath HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}" "DisplayIcon"
    !insertmacro CheckRegistryPath HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}" "UninstallString"

    !insertmacro CheckRegistryPath HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallPath"
    !insertmacro CheckRegistryPath HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "DisplayIcon"
    !insertmacro CheckRegistryPath HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "UninstallString"
    !insertmacro CheckRegistryPath HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}" "DisplayIcon"
    !insertmacro CheckRegistryPath HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}" "UninstallString"

    # 3. Check 32-bit registry entries
    SetRegView 32
    !insertmacro CheckRegistryPath HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallPath"
    !insertmacro CheckRegistryPath HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UNINST_KEY_NAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallLocation"
    !insertmacro CheckRegistryPath HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallPath"
    SetRegView 64

    # 4. If registry check discovered existing installation path, adopt it
    ${If} $DetectedPath != ""
        StrCpy $INSTDIR $DetectedPath
        Return
    ${EndIf}

    # 5. Check standard filesystem locations
    ${If} ${FileExists} "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}\${PRODUCT_EXECUTABLE}"
        StrCpy $INSTDIR "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
        Return
    ${EndIf}

    ${If} ${FileExists} "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}\${PRODUCT_EXECUTABLE}"
        StrCpy $INSTDIR "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
        Return
    ${EndIf}

    ${If} ${FileExists} "$PROGRAMFILES64\${INFO_PRODUCTNAME}\${PRODUCT_EXECUTABLE}"
        StrCpy $INSTDIR "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
        Return
    ${EndIf}

    ${If} ${FileExists} "$PROGRAMFILES\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}\${PRODUCT_EXECUTABLE}"
        StrCpy $INSTDIR "$PROGRAMFILES\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
        Return
    ${EndIf}

    ${If} ${FileExists} "$PROGRAMFILES\${INFO_PRODUCTNAME}\${PRODUCT_EXECUTABLE}"
        StrCpy $INSTDIR "$PROGRAMFILES\${INFO_PRODUCTNAME}"
        Return
    ${EndIf}

    # 6. Default destination for new installations
    UserInfo::GetAccountType
    Pop $1
    ${If} $1 == "Admin"
        StrCpy $INSTDIR "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
    ${Else}
        StrCpy $INSTDIR "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
    ${EndIf}
FunctionEnd

# Section 1: Core application files (Required)
Section "!${INFO_PRODUCTNAME} Core Application" SecCore
    SectionIn RO # Mandatory component

    !insertmacro wails.setShellContext
    !insertmacro wails.webview2runtime

    # Terminate running instances before overwriting files
    DetailPrint "Closing existing ${INFO_PRODUCTNAME} instances..."
    nsExec::Exec 'cmd.exe /c "taskkill /IM ${PRODUCT_EXECUTABLE} /F >nul 2>&1 || exit 0"'

    SetOutPath $INSTDIR

    # Install main executable and bundled files
    !insertmacro wails.files
    File "..\..\..\backend\services\vintaservices.dll"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller

    # Record installation details in HKCU registry for detection and Windows settings
    SetRegView 64
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "DisplayName" "${INFO_PRODUCTNAME}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "DisplayVersion" "${INFO_PRODUCTVERSION}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "Publisher" "${INFO_COMPANYNAME}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "DisplayIcon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"

    WriteRegStr HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallPath" "$INSTDIR"

    # Also record in HKLM if running with administrator privileges
    UserInfo::GetAccountType
    Pop $1
    ${If} $1 == "Admin"
        WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "DisplayName" "${INFO_PRODUCTNAME}"
        WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "DisplayVersion" "${INFO_PRODUCTVERSION}"
        WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "Publisher" "${INFO_COMPANYNAME}"
        WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "DisplayIcon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
        WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "InstallLocation" "$INSTDIR"
        WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
        WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"

        WriteRegStr HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallLocation" "$INSTDIR"
        WriteRegStr HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" "InstallPath" "$INSTDIR"
    ${EndIf}
SectionEnd

# Section 2: Start Menu Shortcut (Checked by default)
Section "Start Menu Shortcut" SecStartMenu
    !insertmacro wails.setShellContext
    CreateDirectory "$SMPROGRAMS\${INFO_COMPANYNAME}"
    CreateShortcut "$SMPROGRAMS\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
SectionEnd

# Section 3: Desktop Shortcut (Checked by default)
Section "Desktop Shortcut" SecDesktop
    !insertmacro wails.setShellContext
    CreateShortcut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    # Terminate running instances before uninstallation
    DetailPrint "Closing active ${INFO_PRODUCTNAME} processes..."
    nsExec::Exec 'cmd.exe /c "taskkill /IM ${PRODUCT_EXECUTABLE} /F >nul 2>&1 || exit 0"'

    # Remove shortcuts if they exist
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"
    Delete "$SMPROGRAMS\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}.lnk"
    RMDir "$SMPROGRAMS\${INFO_COMPANYNAME}"

    # Remove installed files
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}"
    Delete "$INSTDIR\vintaservices.dll"
    Delete "$INSTDIR\uninstall.exe"

    # Remove application directory
    RMDir /r "$INSTDIR"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller

    # Clean up registry keys
    SetRegView 64
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "${INFO_PRODUCTNAME}"
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run" "${INFO_PRODUCTNAME}"
    DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}"
    DeleteRegKey HKCU "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
    DeleteRegKey /ifempty HKCU "Software\${INFO_COMPANYNAME}"

    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${INFO_PRODUCTNAME}"
    DeleteRegKey HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
    DeleteRegKey /ifempty HKLM "Software\${INFO_COMPANYNAME}"
SectionEnd
