import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

class ApiClient {
  final String baseUrl;
  final http.Client client;

  ApiClient({
    String? baseUrl,
    http.Client? client,
  })  : baseUrl = baseUrl ??
            const String.fromEnvironment(
              'API_URL',
              defaultValue: kIsWeb
                  ? '/api/v1'
                  : 'http://10.0.2.2:8080/api/v1',
            ),
        client = client ?? http.Client();

  /// Resolve absolute or relative URI
  Uri _buildUri(String path) {
    if (baseUrl.startsWith('http://') || baseUrl.startsWith('https://')) {
      return Uri.parse('$baseUrl$path');
    }
    // Relative path for Flutter Web through proxy
    return Uri.parse('$baseUrl$path');
  }

  Future<Map<String, dynamic>> checkHealth() async {
    try {
      final response = await client
          .get(
            _buildUri('/healthz'),
            headers: {'Accept': 'application/json'},
          )
          .timeout(const Duration(seconds: 5));

      if (response.statusCode == 200) {
        final decoded = json.decode(response.body);
        return decoded is Map<String, dynamic> ? decoded : {'status': 'OK'};
      }
      return {'status': 'ERROR', 'code': response.statusCode};
    } catch (e) {
      return {'status': 'OFFLINE', 'error': e.toString()};
    }
  }

  Future<Map<String, dynamic>> login(String email, String password) async {
    try {
      final response = await client
          .post(
            _buildUri('/auth/login'),
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
        'message': decoded['error']?['message'] ?? 'Login failed',
      };
    } catch (e) {
      return {'success': false, 'message': e.toString()};
    }
  }

  Future<Map<String, dynamic>> getProfile(String accessToken) async {
    try {
      final response = await client
          .get(
            _buildUri('/auth/me'),
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
