#define MyAppName "Just Install"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "Lorenzo Villani"
#define MyAppURL "http://lorenzo.villani.me/project/just-install"
#define MyAppExeName "just-install.exe"

[Setup]
AppId={{82A92C6B-6C6A-4AA6-8A28-CC553B785F2F}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
ChangesEnvironment=true
DefaultDirName={pf}\JustInstall
DisableDirPage=yes
DisableReadyPage=yes
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
OutputDir="."
OutputBaseFilename=just-install-setup
Compression=lzma
SolidCompression=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "dist\just-install.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"

[Code]
const
    ModPathName = '';
    ModPathType = 'user';

function ModPathDir(): TArrayOfString;
begin
    setArrayLength(Result, 1)
    Result[0] := ExpandConstant('{app}');
end;
#include "modpath.iss"
