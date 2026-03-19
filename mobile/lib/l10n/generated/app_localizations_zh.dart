// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Chinese (`zh`).
class AppLocalizationsZh extends AppLocalizations {
  AppLocalizationsZh([String locale = 'zh']) : super(locale);

  @override
  String get appTitle => 'codeScope Mobile';

  @override
  String get sessionsTitle => 'codeScope 会话';

  @override
  String get settingsTitle => '设置';

  @override
  String get settingsTooltip => '设置';

  @override
  String get remoteSessionsHeadline => '远程 AI 编程会话';

  @override
  String sessionsReady(int count) {
    return '已准备 $count 个会话';
  }

  @override
  String get mockTransportSuffix => '当前使用 Mock 传输';

  @override
  String get retry => '重试';

  @override
  String get files => '文件';

  @override
  String get prompt => 'Prompt';

  @override
  String get sessionDetailsTitle => '会话详情';

  @override
  String get fileBrowserPlaceholderTitle => '文件浏览占位页';

  @override
  String get fileBrowserPlaceholderBody =>
      '预留给 GET /api/sessions/:id/files/tree 与文件内容接口。';

  @override
  String get promptPlaceholderTitle => 'Prompt 下发占位页';

  @override
  String get promptPlaceholderBody =>
      '预留给 GET /api/prompts 和 POST /api/sessions/:id/commands/prompt。';

  @override
  String currentSessionLabel(String sessionId) {
    return '当前会话：$sessionId';
  }

  @override
  String get settingsSectionGeneral => '通用设置';

  @override
  String get settingsSectionConnection => '连接配置';

  @override
  String get languageLabel => '语言';

  @override
  String get languageChinese => '中文';

  @override
  String get languageEnglish => 'English';

  @override
  String get dataSourceLabel => '数据源';

  @override
  String get dataSourceMock => 'Mock';

  @override
  String get dataSourceServer => 'Server';

  @override
  String get apiBaseUrlLabel => 'REST API 地址';

  @override
  String get webSocketUrlLabel => 'WebSocket 地址';

  @override
  String get settingsHintRuntimeOnly => '当前改动仅对本次运行生效。';

  @override
  String get settingsTestConnection => '测试连接';

  @override
  String get settingsTestingConnection => '正在测试连接...';

  @override
  String get settingsConnectionSuccess => '连接成功';

  @override
  String get settingsConnectionMockOnly => 'Mock 模式无需测试服务端连接。';

  @override
  String settingsConnectionFailure(String message) {
    return '连接失败：$message';
  }

  @override
  String get saveSettings => '保存设置';

  @override
  String get settingsSaved => '设置已应用';

  @override
  String get machineLabel => '机器';

  @override
  String get workspaceLabel => '工作区';

  @override
  String get statusLabel => '状态';

  @override
  String get createdStatus => '已创建';

  @override
  String get runningStatus => '运行中';

  @override
  String get stoppedStatus => '已停止';

  @override
  String get failedStatus => '失败';

  @override
  String get logTypeAi => 'AI';

  @override
  String get logTypeTerminal => '终端';

  @override
  String get logTypeCommand => '命令';

  @override
  String get logTypeFile => '文件';

  @override
  String get logTypeDiff => 'Diff';

  @override
  String get logTypeError => '错误';

  @override
  String get logTypeHeartbeat => '心跳';

  @override
  String loadingErrorPrefix(String message) {
    return '加载失败：$message';
  }

  @override
  String get settingsSectionAbout => '关于';

  @override
  String get appVersionLabel => '版本';

  @override
  String get appBuildLabel => '构建号';
}
