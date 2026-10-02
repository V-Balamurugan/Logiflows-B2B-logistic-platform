import 'dart:convert';
import 'package:http/http.dart' as http;

class ApiClient {
  final String baseUrl;
  final http.Client client;

  ApiClient({
    this.baseUrl = 'http://10.0.2.2:8080/api/v1',
    http.Client? client,
  }) : client = client ?? http.Client();

  Future<Map<String, dynamic>> checkHealth() async {
    try {
      final response = await client
          .get(
            Uri.parse('$baseUrl/healthz'),
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

  void dispose() {
    client.close();
  }
}
