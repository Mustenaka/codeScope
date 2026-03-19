import 'dart:convert';
import 'dart:io';

import '../../modules/file/file_content_record.dart';
import '../../modules/file/file_tree_node.dart';
import '../../modules/log/log_event.dart';
import '../../modules/message/thread_message_record.dart';
import '../../modules/project/project_record.dart';
import '../../modules/prompt/prompt_command_task.dart';
import '../../modules/session/session_record.dart';
import '../../modules/thread/thread_record.dart';
import '../rest_client.dart';
import 'server_api_mapper.dart';

typedef HttpClientFactory = HttpClient Function();

class ServerApiException implements Exception {
  ServerApiException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;

  @override
  String toString() {
    if (statusCode == null) {
      return 'ServerApiException: $message';
    }
    return 'ServerApiException($statusCode): $message';
  }
}

class ServerCodeScopeRestClient implements CodeScopeRestClient {
  ServerCodeScopeRestClient(
    this.baseUrl, {
    HttpClientFactory? httpClientFactory,
  }) : _httpClientFactory = httpClientFactory ?? HttpClient.new;

  final String baseUrl;
  final HttpClientFactory _httpClientFactory;

  @override
  Future<List<ProjectRecord>> fetchProjects() async {
    final body = await _getJson('projects');
    if (body is! List<Object?>) {
      throw const FormatException('Expected a JSON array for projects.');
    }
    return ServerApiMapper.projectListFromJson(body);
  }

  @override
  Future<ProjectRecord> fetchProjectDetail(String projectId) async {
    final body = await _getJson('projects/$projectId');
    return ServerApiMapper.projectFromJson(_asJsonMap(body));
  }

  @override
  Future<List<ThreadRecord>> fetchProjectThreads(String projectId) async {
    final body = await _getJson('projects/$projectId/threads');
    if (body is! List<Object?>) {
      throw const FormatException('Expected a JSON array for threads.');
    }
    return ServerApiMapper.threadListFromJson(body);
  }

  @override
  Future<ThreadRecord> fetchThreadDetail(String threadId) async {
    final body = await _getJson('threads/$threadId');
    return ServerApiMapper.threadFromJson(_asJsonMap(body));
  }

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) async {
    final body = await _getJson('threads/$threadId/messages');
    if (body is! List<Object?>) {
      throw const FormatException('Expected a JSON array for thread messages.');
    }
    return ServerApiMapper.threadMessageListFromJson(body);
  }

  @override
  Future<List<SessionRecord>> fetchSessions() async {
    final body = await _getJson('sessions');
    if (body is! List<Object?>) {
      throw const FormatException('Expected a JSON array for sessions.');
    }
    return ServerApiMapper.sessionListFromJson(body);
  }

  @override
  Future<SessionRecord> fetchSessionDetail(String sessionId) async {
    final body = await _getJson('sessions/$sessionId');
    return ServerApiMapper.sessionFromJson(_asJsonMap(body));
  }

  @override
  Future<List<LogEvent>> fetchSessionEvents(String sessionId) async {
    final body = await _getJson('sessions/$sessionId/events');
    if (body is! List<Object?>) {
      throw const FormatException('Expected a JSON array for events.');
    }
    return ServerApiMapper.logEventListFromJson(body);
  }

  @override
  Future<List<PromptCommandTask>> fetchSessionCommands(String sessionId) async {
    final body = await _getJson('sessions/$sessionId/commands');
    if (body is! List<Object?>) {
      throw const FormatException('Expected a JSON array for command tasks.');
    }
    return ServerApiMapper.commandTaskListFromJson(body);
  }

  @override
  Future<PromptCommandTask> sendPrompt(String sessionId, String content) async {
    final body = await _postJson(
      'sessions/$sessionId/commands/prompt',
      <String, Object?>{'content': content},
    );
    return ServerApiMapper.commandTaskFromJson(_asJsonMap(body));
  }

  @override
  Future<List<FileTreeNode>> fetchSessionFileTree(String sessionId) async {
    final body = await _getJson('sessions/$sessionId/files/tree');
    if (body is! List<Object?>) {
      throw const FormatException('Expected a JSON array for file tree.');
    }
    return ServerApiMapper.fileTreeFromJson(body);
  }

  @override
  Future<FileContentRecord> fetchSessionFileContent(
    String sessionId,
    String path,
  ) async {
    final body = await _getJson(
      'sessions/$sessionId/files/content?path=${Uri.encodeQueryComponent(path)}',
    );
    return ServerApiMapper.fileContentFromJson(_asJsonMap(body));
  }

  Future<Object?> _getJson(String path) async {
    final client = _httpClientFactory();
    try {
      final request = await client.getUrl(_buildUri(path));
      request.headers.set(HttpHeaders.acceptHeader, ContentType.json.mimeType);
      final response = await request.close();
      final responseBody = await utf8.decoder.bind(response).join();

      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw ServerApiException(
          _errorMessageFromBody(responseBody) ??
              'Request failed with status ${response.statusCode}.',
          statusCode: response.statusCode,
        );
      }

      if (responseBody.isEmpty) {
        throw const FormatException('Server returned an empty response body.');
      }

      return jsonDecode(responseBody);
    } on SocketException catch (error) {
      throw ServerApiException('Network request failed: ${error.message}');
    } on HttpException catch (error) {
      throw ServerApiException('HTTP request failed: ${error.message}');
    } finally {
      client.close(force: true);
    }
  }

  Future<Object?> _postJson(String path, Object body) async {
    final client = _httpClientFactory();
    try {
      final request = await client.postUrl(_buildUri(path));
      request.headers.contentType = ContentType.json;
      request.headers.set(HttpHeaders.acceptHeader, ContentType.json.mimeType);
      request.write(jsonEncode(body));
      final response = await request.close();
      final responseBody = await utf8.decoder.bind(response).join();

      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw ServerApiException(
          _errorMessageFromBody(responseBody) ??
              'Request failed with status ${response.statusCode}.',
          statusCode: response.statusCode,
        );
      }

      if (responseBody.isEmpty) {
        throw const FormatException('Server returned an empty response body.');
      }

      return jsonDecode(responseBody);
    } on SocketException catch (error) {
      throw ServerApiException('Network request failed: ${error.message}');
    } on HttpException catch (error) {
      throw ServerApiException('HTTP request failed: ${error.message}');
    } finally {
      client.close(force: true);
    }
  }

  Uri _buildUri(String path) {
    final normalizedBase = baseUrl.endsWith('/') ? baseUrl : '$baseUrl/';
    return Uri.parse(normalizedBase).resolve(path);
  }

  String? _errorMessageFromBody(String responseBody) {
    if (responseBody.isEmpty) {
      return null;
    }

    try {
      final json = jsonDecode(responseBody);
      final payload = _asJsonMap(json);
      final error = payload['error'];
      if (error is String && error.isNotEmpty) {
        return error;
      }
    } on FormatException {
      return responseBody;
    }

    return responseBody;
  }

  Map<String, Object?> _asJsonMap(Object? value) {
    if (value is Map<String, Object?>) {
      return value;
    }
    if (value is Map) {
      return value.map(
        (Object? key, Object? entryValue) =>
            MapEntry(key.toString(), entryValue),
      );
    }
    throw const FormatException('Expected a JSON object.');
  }
}
