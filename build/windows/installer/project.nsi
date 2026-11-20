Unicode true

####
## Reka Shade installer (NSIS). Based on the Wails template, customised to:
##  - install per user (no admin) so the in-app updater can replace the exe;
##  - ask a running copy to quit cleanly first (it restores your displays);
##  - speak English and French; offer to launch Reka Shade at the end.
## Built by `wails build -nsis -installscope user` (see .github/workflows).
####

!include "wails_tools.nsh"
!include "WinMessages.nsh"

VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

ManifestDPIAware true

!include "MUI2.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!define MUI_FINISHPAGE_RUN_TEXT "$(LaunchText)"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "French"

LangString LaunchText ${LANG_ENGLISH} "Launch Reka Shade"
LangString LaunchText ${LANG_FRENCH} "Lancer Reka Shade"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"

!ifdef WAILS_INSTALL_SCOPE
  !if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
  !else
    InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
  !endif
!else
  InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
!endif

ShowInstDetails show

; Ask a running Reka Shade to exit (it restores gamma, vibrance and resolution
; on the way out), then make sure it is gone before files are replaced.
!macro rekashade.closeRunning
    FindWindow $0 "RekaShadeHost"
    StrCmp $0 0 rs_closed
    SendMessage $0 ${WM_CLOSE} 0 0
    Sleep 2500
    FindWindow $0 "RekaShadeHost"
    StrCmp $0 0 rs_closed
    nsExec::Exec 'taskkill /F /IM "${PRODUCT_EXECUTABLE}"'
    Pop $0
    Sleep 500
    rs_closed:
!macroend

Function .onInit
    !insertmacro wails.checkArchitecture
FunctionEnd

Section
    !insertmacro wails.setShellContext
    !insertmacro rekashade.closeRunning
    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}.old"
    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext
    !insertmacro rekashade.closeRunning

    ; Startup entry written by the app itself.
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "RekaShade"
    ; WebView2 cache. Your settings in %APPDATA%\RekaShade are kept.
    RMDir /r "$LOCALAPPDATA\RekaShade"

    RMDir /r $INSTDIR
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.deleteUninstaller
SectionEnd
