; MuxueTools Windows Installer Script
; Inno Setup 6.x

#define MyAppName "MuxueTools"
#define MyAppVersion GetEnv('VERSION')
#if MyAppVersion == ""
  #define MyAppVersion "0.3.1"
#endif
#define MyAppPublisher "muxueliunian"
#define MyAppURL "https://github.com/muxueliunian/muxueTools"
#define MyAppExeName "muxueTools.exe"

[Setup]
; 应用程序唯一标识符 (PowerShell [guid]::NewGuid() 生成)
AppId={{D2E16969-12AA-43B0-9CE7-D9BB27A546AC}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
; 许可文件 (暂时禁用，项目未提供 LICENSE 文件)
; LicenseFile=..\..\..\LICENSE
OutputDir=..\..\..\dist
OutputBaseFilename=MuxueTools-Setup-{#MyAppVersion}
SetupIconFile=..\..\..\assets\icon.ico
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

; 权限设置 (不需要管理员权限，安装到用户目录)
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog

; 卸载设置
UninstallDisplayIcon={app}\{#MyAppExeName}
UninstallDisplayName={#MyAppName}

[Languages]
; 简体中文 (默认)
Name: "chinesesimplified"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"
; 英语
Name: "english"; MessagesFile: "compiler:Default.isl"
; 日语
Name: "japanese"; MessagesFile: "compiler:Languages\Japanese.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
; 主程序
Source: "..\..\..\bin\muxueTools.exe"; DestDir: "{app}"; Flags: ignoreversion

; 前端资源
Source: "..\..\..\web\dist\*"; DestDir: "{app}\web\dist"; Flags: ignoreversion recursesubdirs createallsubdirs

; 配置模板
Source: "..\..\..\configs\config.example.yaml"; DestDir: "{app}"; DestName: "config.example.yaml"; Flags: ignoreversion

[Icons]
; 开始菜单快捷方式
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
; 桌面快捷方式 (可选)
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
; 安装完成后可选运行程序
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Code]
const
  AppExeName = 'muxueTools.exe';

var
  ShouldDeleteUserData: Boolean;

// ============================================================
// Windows API 声明
// ============================================================
function FindWindowByClassName(lpClassName, lpWindowName: String): HWND;
  external 'FindWindowW@user32.dll stdcall';

function ShellExecuteW(hwnd: HWND; lpOperation, lpFile, lpParameters, lpDirectory: String; nShowCmd: Integer): Integer;
  external 'ShellExecuteW@shell32.dll stdcall';

// ============================================================
// 多语言消息
// ============================================================
function GetMsg(MsgId: String): String;
begin
  case MsgId of
    'AppRunning':
      case ActiveLanguage() of
        'chinesesimplified': Result := 'MuxueTools 正在运行，请先关闭程序再继续。' + #13#10 + #13#10 + '点击"确定"自动关闭程序，或点击"取消"手动关闭。';
        'japanese': Result := 'MuxueTools が実行中です。続行する前にプログラムを閉じてください。' + #13#10 + #13#10 + '「OK」をクリックして自動的に閉じるか、「キャンセル」をクリックして手動で閉じてください。';
        else Result := 'MuxueTools is running. Please close it before continuing.' + #13#10 + #13#10 + 'Click "OK" to close it automatically, or "Cancel" to close it manually.';
      end;
    'OldVersionFound':
      case ActiveLanguage() of
        'chinesesimplified': Result := '检测到已安装的旧版本。' + #13#10 + #13#10 + '点击"是"将打开卸载向导，请完成卸载后继续安装新版本。' + #13#10 + '（在卸载向导中您可以选择是否删除用户数据）';
        'japanese': Result := '古いバージョンが検出されました。' + #13#10 + #13#10 + '「はい」をクリックするとアンインストールウィザードが開きます。' + #13#10 + '（アンインストール時にユーザーデータを削除するかどうかを選択できます）';
        else Result := 'An older version is installed.' + #13#10 + #13#10 + 'Click "Yes" to open the uninstall wizard.' + #13#10 + '(You can choose whether to delete user data during uninstall)';
      end;
    'UninstallFailed':
      case ActiveLanguage() of
        'chinesesimplified': Result := '无法自动卸载旧版本。请手动卸载后重试。';
        'japanese': Result := '古いバージョンを自動的にアンインストールできませんでした。手動でアンインストールしてから再試行してください。';
        else Result := 'Could not uninstall the old version automatically. Please uninstall it manually and try again.';
      end;
    'DeleteUserData':
      case ActiveLanguage() of
        'chinesesimplified': Result := '同时删除用户数据（API密钥、对话历史、配置文件）';
        'japanese': Result := 'ユーザーデータも削除する（APIキー、会話履歴、設定ファイル）';
        else Result := 'Also delete user data (API keys, chat history, config files)';
      end;
    'DeleteFailed':
      case ActiveLanguage() of
        'chinesesimplified': Result := '无法删除用户数据，程序可能仍在运行。' + #13#10 + #13#10 + '是否尝试结束 MuxueTools 进程并重试？';
        'japanese': Result := 'ユーザーデータを削除できませんでした。プログラムがまだ実行中の可能性があります。' + #13#10 + #13#10 + 'MuxueTools プロセスを終了して再試行しますか？';
        else Result := 'Could not delete user data. The program may still be running.' + #13#10 + #13#10 + 'Do you want to terminate MuxueTools and retry?';
      end;
    'DataDeleted':
      case ActiveLanguage() of
        'chinesesimplified': Result := '用户数据已成功删除。';
        'japanese': Result := 'ユーザーデータが正常に削除されました。';
        else Result := 'User data has been successfully deleted.';
      end;
    'AskDeleteData':
      case ActiveLanguage() of
        'chinesesimplified': Result := '是否同时删除用户数据？' + #13#10 + #13#10 + '用户数据包括：' + #13#10 + '• API 密钥' + #13#10 + '• 对话历史记录' + #13#10 + '• 配置文件' + #13#10 + #13#10 + '点击"是"删除数据，点击"否"保留数据。';
        'japanese': Result := 'ユーザーデータも削除しますか？' + #13#10 + #13#10 + 'ユーザーデータには以下が含まれます：' + #13#10 + '• APIキー' + #13#10 + '• 会話履歴' + #13#10 + '• 設定ファイル' + #13#10 + #13#10 + '「はい」をクリックするとデータが削除されます。';
        else Result := 'Do you also want to delete user data?' + #13#10 + #13#10 + 'User data includes:' + #13#10 + '• API Keys' + #13#10 + '• Chat History' + #13#10 + '• Config Files' + #13#10 + #13#10 + 'Click "Yes" to delete, "No" to keep.';
      end;
  end;
end;

// ============================================================
// 进程检测与终止
// ============================================================
function IsAppRunning(): Boolean;
var
  ResultCode: Integer;
begin
  // 使用 tasklist 检查进程是否存在
  Result := Exec('cmd.exe', '/c tasklist /FI "IMAGENAME eq ' + AppExeName + '" | find /i "' + AppExeName + '" >nul', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) and (ResultCode = 0);
end;

function KillApp(): Boolean;
var
  ResultCode: Integer;
begin
  Result := Exec('cmd.exe', '/c taskkill /F /IM ' + AppExeName, '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Sleep(500); // 等待进程完全退出
end;

function HandleRunningApp(): Boolean;
begin
  Result := True;
  while IsAppRunning() do
  begin
    if MsgBox(GetMsg('AppRunning'), mbConfirmation, MB_OKCANCEL) = IDOK then
    begin
      KillApp();
      Sleep(500);
    end
    else
    begin
      Result := False;
      Exit;
    end;
  end;
end;

// ============================================================
// 旧版本检测与卸载
// ============================================================
function GetUninstallString(): String;
var
  UninstallKey: String;
begin
  Result := '';
  UninstallKey := 'Software\Microsoft\Windows\CurrentVersion\Uninstall\{D2E16969-12AA-43B0-9CE7-D9BB27A546AC}_is1';
  
  // 先检查当前用户
  if not RegQueryStringValue(HKCU, UninstallKey, 'UninstallString', Result) then
    // 再检查所有用户
    RegQueryStringValue(HKLM, UninstallKey, 'UninstallString', Result);
end;

function UninstallOldVersion(): Boolean;
var
  UninstallString: String;
  ResultCode: Integer;
begin
  Result := True;
  UninstallString := GetUninstallString();
  
  if UninstallString <> '' then
  begin
    // 询问用户是否卸载旧版本
    if MsgBox(GetMsg('OldVersionFound'), mbConfirmation, MB_YESNO) = IDYES then
    begin
      // 打开卸载向导（不使用 /SILENT，让用户可以选择是否删除数据）
      UninstallString := RemoveQuotes(UninstallString);
      if not Exec(UninstallString, '', '', SW_SHOW, ewWaitUntilTerminated, ResultCode) then
      begin
        MsgBox(GetMsg('UninstallFailed'), mbError, MB_OK);
        Result := False;
      end;
    end
    else
    begin
      Result := False;
    end;
  end;
end;

// ============================================================
// 用户数据删除
// ============================================================
function TryDeleteUserData(): Boolean;
var
  UserDataPath: String;
  FindRec: TFindRec;
  FilePath: String;
begin
  Result := True;
  UserDataPath := ExpandConstant('{userappdata}\MuxueTools');
  
  if not DirExists(UserDataPath) then
    Exit;
  
  // 尝试删除所有文件
  if FindFirst(UserDataPath + '\*', FindRec) then
  begin
    try
      repeat
        if (FindRec.Name <> '.') and (FindRec.Name <> '..') then
        begin
          FilePath := UserDataPath + '\' + FindRec.Name;
          if FindRec.Attributes and FILE_ATTRIBUTE_DIRECTORY <> 0 then
          begin
            if not DelTree(FilePath, True, True, True) then
              Result := False;
          end
          else
          begin
            if not DeleteFile(FilePath) then
              Result := False;
          end;
        end;
      until not FindNext(FindRec);
    finally
      FindClose(FindRec);
    end;
  end;
  
  // 删除主目录
  if Result then
    Result := RemoveDir(UserDataPath);
end;

procedure DeleteUserDataWithRetry();
begin
  if not TryDeleteUserData() then
  begin
    // 删除失败，询问是否结束进程
    if MsgBox(GetMsg('DeleteFailed'), mbConfirmation, MB_YESNO) = IDYES then
    begin
      KillApp();
      Sleep(1000);
      // 重试删除
      if TryDeleteUserData() then
        MsgBox(GetMsg('DataDeleted'), mbInformation, MB_OK);
    end;
  end
  else
  begin
    MsgBox(GetMsg('DataDeleted'), mbInformation, MB_OK);
  end;
end;

// ============================================================
// 安装初始化
// ============================================================
function InitializeSetup(): Boolean;
begin
  Result := True;
  
  // 检查程序是否正在运行
  if not HandleRunningApp() then
  begin
    Result := False;
    Exit;
  end;
  
  // 检查并卸载旧版本
  if GetUninstallString() <> '' then
  begin
    Result := UninstallOldVersion();
  end;
end;

// ============================================================
// 卸载初始化 - 在卸载开始前询问是否删除用户数据
// ============================================================
function InitializeUninstall(): Boolean;
var
  UserDataPath: String;
begin
  Result := True;
  ShouldDeleteUserData := False;
  
  // 检查程序是否正在运行
  if not HandleRunningApp() then
  begin
    Result := False;
    Exit;
  end;
  
  // 检查用户数据是否存在，如果存在则询问是否删除
  UserDataPath := ExpandConstant('{userappdata}\MuxueTools');
  if DirExists(UserDataPath) then
  begin
    if MsgBox(GetMsg('AskDeleteData'), mbConfirmation, MB_YESNO) = IDYES then
    begin
      ShouldDeleteUserData := True;
    end;
  end;
end;

// ============================================================
// 卸载完成后处理
// ============================================================
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    // 根据之前的选择删除用户数据
    if ShouldDeleteUserData then
    begin
      DeleteUserDataWithRetry();
    end;
  end;
end;

// ============================================================
// 安装完成后处理
// ============================================================
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    // 可以添加后处理逻辑
  end;
end;

