import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:logiflows_mobile/core/network/api_client.dart';

void main() {
  group('ApiClient Core Tests', () {
    test('buildUri formats paths correctly', () {
      final client = ApiClient(baseUrl: 'http://localhost:8080/api/v1');
      expect(client.buildUri('/healthz').toString(), 'http://localhost:8080/api/v1/healthz');
      expect(client.buildUri('vehicles').toString(), 'http://localhost:8080/api/v1/vehicles');
    });

    test('getVehicles returns list of fleet vehicles', () async {
      final mockClient = MockClient((request) async {
        expect(request.url.path, '/api/v1/vehicles');
        expect(request.headers['Authorization'], 'Bearer mock-jwt-token');

        return http.Response(
          json.encode({
            'data': [
              {
                'id': '00000000-0000-0000-0001-000000000001',
                'license_plate': 'KA-01-EV-1001',
                'make': 'Tata',
                'model': 'Ace EV',
                'vehicle_type': 'ELECTRIC_VAN',
                'status': 'AVAILABLE',
                'battery_or_fuel_level_percent': 90.0
              }
            ]
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final apiClient = ApiClient(baseUrl: 'http://localhost:8080/api/v1', client: mockClient);
      final res = await apiClient.getVehicles('mock-jwt-token');

      expect(res['success'], true);
      final data = res['data'] as List;
      expect(data.length, 1);
      expect(data[0]['license_plate'], 'KA-01-EV-1001');
      expect(data[0]['vehicle_type'], 'ELECTRIC_VAN');
    });

    test('sendTelematicsPing posts coordinates and returns 201', () async {
      final mockClient = MockClient((request) async {
        expect(request.url.path, '/api/v1/vehicles/veh-123/telematics');
        expect(request.method, 'POST');
        final body = json.decode(request.body);
        expect(body['latitude'], 12.9716);
        expect(body['longitude'], 77.5946);
        expect(body['speed_kmh'], 45.0);

        return http.Response(
          json.encode({
            'data': {
              'vehicle_id': 'veh-123',
              'latitude': 12.9716,
              'longitude': 77.5946,
              'speed_kmh': 45.0
            }
          }),
          201,
          headers: {'content-type': 'application/json'},
        );
      });

      final apiClient = ApiClient(baseUrl: 'http://localhost:8080/api/v1', client: mockClient);
      final res = await apiClient.sendTelematicsPing(
        accessToken: 'mock-jwt-token',
        vehicleId: 'veh-123',
        latitude: 12.9716,
        longitude: 77.5946,
        speedKmh: 45.0,
        headingDegrees: 180.0,
        batteryOrFuelPercent: 85.0,
        odometerKm: 1200.0,
      );

      expect(res['success'], true);
      expect(res['data']['vehicle_id'], 'veh-123');
    });

    test('getLiveFleet fetches active fleet positions', () async {
      final mockClient = MockClient((request) async {
        expect(request.url.path, '/api/v1/fleet/live');
        return http.Response(
          json.encode({
            'data': [
              {
                'vehicle_id': 'veh-123',
                'license_plate': 'KA-01-EV-1001',
                'latitude': 12.9716,
                'longitude': 77.5946,
                'status': 'ON_ROUTE'
              }
            ]
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final apiClient = ApiClient(baseUrl: 'http://localhost:8080/api/v1', client: mockClient);
      final res = await apiClient.getLiveFleet('mock-jwt-token');

      expect(res['success'], true);
      final data = res['data'] as List;
      expect(data.length, 1);
      expect(data[0]['license_plate'], 'KA-01-EV-1001');
    });
  });
}
