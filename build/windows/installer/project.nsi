Unicode true

####
## Vinta NSIS Windows Installer Configuration
####

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
ShowInstDetails show

# Function to ensure selected directory always ends with \Vinta subfolder
Function DirectoryLeave
    Var /GLOBAL PathTail

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

Function .onInit
    !insertmacro wails.checkArchitecture
    # If running with administrator privileges, prefer Program Files
    UserInfo::GetAccountType
    Pop $0
    ${If} $0 == "Admin"
        StrCpy $INSTDIR "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
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
SectionEnd
