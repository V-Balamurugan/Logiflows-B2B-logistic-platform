import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

class ApiClient {
  String baseUrl;
  final http.Client client;

  ApiClient({
    String? baseUrl,
    http.Client? client,
  })  : baseUrl = baseUrl ?? detectDefaultBaseUrl(),
        client = client ?? http.Client();

  /// Automatically detect the optimal API gateway URL depending on runtime platform
  static String detectDefaultBaseUrl() {
    const envUrl = String.fromEnvironment('API_URL');
    if (envUrl.isNotEmpty) {
      return envUrl;
    }

    if (kIsWeb) {
      final base = Uri.base;
      if ((base.scheme == 'http' || base.scheme == 'https') && base.port == 5174) {
        // When running in the mobile docker container, proxy via nginx
        return '${base.origin}/api/v1';
      }
      return 'http://localhost:8080/api/v1';
    }

    // Native target platform detection
    switch (defaultTargetPlatform) {
      case TargetPlatform.android:
        // Android Emulator uses 10.0.2.2 to loop back to host localhost
        return 'http://10.0.2.2:8080/api/v1';
      case TargetPlatform.windows:
      case TargetPlatform.macOS:
      case TargetPlatform.linux:
      case TargetPlatform.iOS:
      default:
        // Windows Desktop, macOS, Linux, and iOS Simulator connect directly to localhost
        return 'http://localhost:8080/api/v1';
    }
  }

  void updateBaseUrl(String newUrl) {
    var trimmed = newUrl.trim();
    if (trimmed.endsWith('/')) {
      trimmed = trimmed.substring(0, trimmed.length - 1);
    }
    baseUrl = trimmed;
  }

  /// Construct guaranteed absolute URI with valid HTTP/HTTPS scheme
  Uri buildUri(String path) {
    final cleanPath = path.startsWith('/') ? path : '/$path';
    final cleanBase = baseUrl.endsWith('/')
        ? baseUrl.substring(0, baseUrl.length - 1)
        : baseUrl;

    if (cleanBase.startsWith('http://') || cleanBase.startsWith('https://')) {
      return Uri.parse('$cleanBase$cleanPath');
    }

    if (kIsWeb) {
      // Resolve relative path against browser window location
      return Uri.base.resolve('$cleanBase$cleanPath');
    }

    // Fallback default for native
    return Uri.parse('http://localhost:8080$cleanBase$cleanPath');
  }

  Future<Map<String, dynamic>> checkHealth() async {
    final targetUri = buildUri('/healthz');
    final stopwatch = Stopwatch()..start();
    try {
      final response = await client
          .get(
            targetUri,
            headers: {'Accept': 'application/json'},
          )
          .timeout(const Duration(seconds: 5));

      stopwatch.stop();
      if (response.statusCode == 200) {
        final decoded = json.decode(response.body);
        final data = decoded is Map<String, dynamic> ? decoded : {'status': 'OK'};
        return {
          'status': 'OK',
          'data': data,
          'latency_ms': stopwatch.elapsedMilliseconds,
          'target_url': targetUri.toString(),
        };
      }
      return {
        'status': 'ERROR',
        'code': response.statusCode,
        'target_url': targetUri.toString(),
        'error': 'Server returned HTTP ${response.statusCode}',
      };
    } catch (e) {
      stopwatch.stop();
      return {
        'status': 'OFFLINE',
        'target_url': targetUri.toString(),
        'error': e.toString(),
      };
    }
  }

  Future<Map<String, dynamic>> login(String email, String password) async {
    final targetUri = buildUri('/auth/login');
    try {
      final response = await client
          .post(
            targetUri,
            headers: {
              'Content-Type': 'application/json',
              'Accept': 'application/json',
            },
            body: json.encode({'email': email, 'password': password}),
          )
          .timeout(const Duration(seconds: 8));

      final decoded = json.decode(response.body);
      if (response.statusCode == 200) {
        return {'success': true, 'data': decoded['data']};
      }
      return {
        'success': false,
        'message': decoded['error']?['message'] ?? 'Login failed (${response.statusCode})',
      };
    } catch (e) {
      return {'success': false, 'message': e.toString()};
    }
  }

  Future<Map<String, dynamic>> getProfile(String accessToken) async {
    final targetUri = buildUri('/auth/me');
    try {
      final response = await client
          .get(
            targetUri,
            headers: {
              'Accept': 'application/json',
              'Authorization': 'Bearer $accessToken',
            },
          )
          .timeout(const Duration(seconds: 5));

      final decoded = json.decode(response.body);
      if (response.statusCode == 200) {
        return {'success': true, 'data': decoded['data']};
      }
      return {'success': false, 'message': decoded['error']?['message'] ?? 'Failed to load profile'};
    } catch (e) {
      return {'success': false, 'message': e.toString()};
    }
  }

  void dispose() {
    client.close();
  }
}
