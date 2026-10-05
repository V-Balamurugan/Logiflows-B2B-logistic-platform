import 'package:flutter/material.dart';
import 'core/network/api_client.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const LogiFlowsEmployeeApp());
}

class LogiFlowsEmployeeApp extends StatelessWidget {
  const LogiFlowsEmployeeApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'LogiFlows Courier',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        colorSchemeSeed: Colors.indigo,
        scaffoldBackgroundColor: const Color(0xFF090D16),
        useMaterial3: true,
      ),
      home: const EmployeeHomeScreen(),
    );
  }
}

class EmployeeHomeScreen extends StatefulWidget {
  const EmployeeHomeScreen({super.key});

  @override
  State<EmployeeHomeScreen> createState() => _EmployeeHomeScreenState();
}

class _EmployeeHomeScreenState extends State<EmployeeHomeScreen> {
  final ApiClient _apiClient = ApiClient();
  String _apiStatus = 'CHECKING';
  String? _statusError;
  int? _latencyMs;
  String _activeTargetUrl = '';
  bool _isLoading = false;

  final TextEditingController _emailController =
      TextEditingController(text: 'courier@speedycourier.com');
  final TextEditingController _passwordController =
      TextEditingController(text: 'Courier@Speedy2026!');
  final TextEditingController _urlController = TextEditingController();

  Map<String, dynamic>? _user;
  Map<String, dynamic>? _tokens;
  String? _authError;
  bool _isLoggingIn = false;

  List<dynamic> _vehicles = [];
  Map<String, dynamic>? _selectedVehicle;
  bool _isLoadingVehicles = false;
  bool _isSendingTelem = false;
  String? _telemStatusMsg;

  @override
  void initState() {
    super.initState();
    _urlController.text = _apiClient.baseUrl;
    _refreshStatus();
  }

  Future<void> _refreshStatus() async {
    setState(() {
      _isLoading = true;
      _statusError = null;
    });

    final res = await _apiClient.checkHealth();
    setState(() {
      _apiStatus = res['data']?['status']?.toString() ??
          res['status']?.toString() ??
          'OFFLINE';
      _statusError = res['error']?.toString();
      _latencyMs = res['latency_ms'] as int?;
      _activeTargetUrl = res['target_url']?.toString() ?? _apiClient.baseUrl;
      _isLoading = false;
    });
  }

  Future<void> _fetchVehicles() async {
    final token = _tokens?['access_token']?.toString();
    if (token == null) return;
    setState(() => _isLoadingVehicles = true);
    final res = await _apiClient.getVehicles(token);
    setState(() {
      _isLoadingVehicles = false;
      if (res['success'] == true) {
        _vehicles = res['data'] is List ? res['data'] as List : [];
        if (_vehicles.isNotEmpty && _selectedVehicle == null) {
          _selectedVehicle = _vehicles.first as Map<String, dynamic>;
        }
      }
    });
  }

  Future<void> _sendLivePing() async {
    final token = _tokens?['access_token']?.toString();
    final vehId = _selectedVehicle?['id']?.toString();
    if (token == null || vehId == null) return;

    setState(() {
      _isSendingTelem = true;
      _telemStatusMsg = null;
    });

    final currentMileage = (_selectedVehicle?['current_mileage_km'] as num?)?.toDouble() ?? 100.0;
    final currentBattery = (_selectedVehicle?['battery_or_fuel_level_percent'] as num?)?.toDouble() ?? 95.0;

    final res = await _apiClient.sendTelematicsPing(
      accessToken: token,
      vehicleId: vehId,
      latitude: 12.9716,
      longitude: 77.5946,
      speedKmh: 42.0,
      headingDegrees: 120.0,
      batteryOrFuelPercent: (currentBattery - 0.5).clamp(0.0, 100.0),
      odometerKm: currentMileage + 1.2,
    );

    setState(() {
      _isSendingTelem = false;
      if (res['success'] == true) {
        _telemStatusMsg = 'Ping delivered! 12.9716, 77.5946 • 42 km/h';
        if (_selectedVehicle != null) {
          _selectedVehicle!['current_mileage_km'] = currentMileage + 1.2;
          _selectedVehicle!['battery_or_fuel_level_percent'] = (currentBattery - 0.5).clamp(0.0, 100.0);
        }
      } else {
        _telemStatusMsg = 'Error: ${res['message']}';
      }
    });
  }

  Future<void> _handleLogin() async {
    setState(() {
      _isLoggingIn = true;
      _authError = null;
    });

    final res = await _apiClient.login(
      _emailController.text.trim(),
      _passwordController.text,
    );

    setState(() {
      _isLoggingIn = false;
      if (res['success'] == true) {
        _user = res['data']['user'];
        _tokens = res['data']['tokens'];
        _authError = null;
      } else {
        _authError = res['message'] ?? 'Authentication failed';
      }
    });

    if (res['success'] == true) {
      _fetchVehicles();
    }
  }

  void _handleLogout() {
    setState(() {
      _user = null;
      _tokens = null;
      _authError = null;
      _vehicles = [];
      _selectedVehicle = null;
      _telemStatusMsg = null;
    });
  }

  void _showGatewaySettings() {
    _urlController.text = _apiClient.baseUrl;
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF0F172A),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) {
        return Padding(
          padding: const EdgeInsets.all(20.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Row(
                children: [
                  Icon(Icons.settings, color: Colors.indigoAccent, size: 20),
                  SizedBox(width: 8),
                  Text(
                    'Backend Gateway Configuration',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              TextField(
                controller: _urlController,
                decoration: InputDecoration(
                  labelText: 'API Base URL',
                  labelStyle: const TextStyle(color: Colors.white60, fontSize: 13),
                  filled: true,
                  fillColor: const Color(0xFF1E293B),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(10),
                    borderSide: const BorderSide(color: Color(0x1AFFFFFF)),
                  ),
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                ),
                style: const TextStyle(color: Colors.white, fontSize: 13),
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  ActionChip(
                    label: const Text('Localhost (8080)', style: TextStyle(fontSize: 11)),
                    onPressed: () {
                      _urlController.text = 'http://localhost:8080/api/v1';
                    },
                  ),
                  ActionChip(
                    label: const Text('Android (10.0.2.2)', style: TextStyle(fontSize: 11)),
                    onPressed: () {
                      _urlController.text = 'http://10.0.2.2:8080/api/v1';
                    },
                  ),
                  ActionChip(
                    label: const Text('Web Proxy (/api/v1)', style: TextStyle(fontSize: 11)),
                    onPressed: () {
                      _urlController.text = '/api/v1';
                    },
                  ),
                ],
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () {
                  Navigator.pop(ctx);
                  _apiClient.updateBaseUrl(_urlController.text);
                  _refreshStatus();
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.indigoAccent,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                ),
                child: const Text('Save & Reconnect Gateway', style: TextStyle(fontWeight: FontWeight.bold)),
              ),
            ],
          ),
        );
      },
    );
  }

  @override
  void dispose() {
    _apiClient.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _urlController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isOnline = _apiStatus == 'OK';

    return Scaffold(
      appBar: AppBar(
        backgroundColor: const Color(0xFF0F172A),
        title: const Row(
          children: [
            Icon(Icons.local_shipping, color: Colors.indigoAccent),
            SizedBox(width: 8),
            Text(
              'LogiFlows Courier',
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.tune, color: Colors.indigoAccent),
            onPressed: _showGatewaySettings,
            tooltip: 'Configure Backend Gateway',
          ),
          IconButton(
            icon: _isLoading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.refresh),
            onPressed: _isLoading ? null : _refreshStatus,
            tooltip: 'Refresh Pulse',
          ),
          if (_user != null)
            IconButton(
              icon: const Icon(Icons.logout, color: Colors.redAccent),
              onPressed: _handleLogout,
              tooltip: 'Sign Out',
            ),
        ],
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Connection Pulse Banner
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                decoration: BoxDecoration(
                  color: isOnline ? const Color(0x1F22C55E) : const Color(0x1FF59E0B),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: isOnline ? const Color(0x4D22C55E) : const Color(0x4DF59E0B),
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(
                          isOnline ? Icons.cloud_done : Icons.cloud_off,
                          color: isOnline ? Colors.greenAccent : Colors.amberAccent,
                          size: 20,
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Text(
                            isOnline
                                ? 'Authoritative Backend Connected (${_latencyMs ?? 0}ms)'
                                : 'Backend Gateway Standby / Unreachable',
                            style: TextStyle(
                              color: isOnline ? Colors.greenAccent : Colors.amberAccent,
                              fontWeight: FontWeight.w600,
                              fontSize: 13,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 6),
                    Text(
                      'Target: ${_activeTargetUrl.isNotEmpty ? _activeTargetUrl : _apiClient.baseUrl}',
                      style: const TextStyle(
                        color: Colors.white54,
                        fontSize: 11,
                        fontFamily: 'monospace',
                      ),
                    ),
                    if (!isOnline && _statusError != null) ...[
                      const SizedBox(height: 4),
                      Text(
                        _statusError!,
                        style: const TextStyle(color: Colors.amberAccent, fontSize: 11),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 16),

              // Authenticated Courier Session or Login Form
              if (_user != null) ...[
                Card(
                  color: const Color(0xFF131D31),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                    side: const BorderSide(color: Colors.indigo, width: 1.5),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(20.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              '${_user!['first_name']} ${_user!['last_name']}',
                              style: const TextStyle(
                                fontSize: 20,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 10, vertical: 4),
                              decoration: BoxDecoration(
                                color: const Color(0x3310B981),
                                borderRadius: BorderRadius.circular(20),
                                border: Border.all(color: const Color(0x6610B981)),
                              ),
                              child: Text(
                                _user!['role']?.toString() ?? 'EMPLOYEE',
                                style: const TextStyle(
                                  color: Colors.greenAccent,
                                  fontSize: 11,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 6),
                        Text(
                          _user!['email']?.toString() ?? '',
                          style: const TextStyle(color: Colors.white70, fontSize: 13),
                        ),
                        const Divider(height: 24, color: Colors.white12),
                        Row(
                          children: [
                            const Icon(Icons.verified, size: 16, color: Colors.indigoAccent),
                            const SizedBox(width: 6),
                            Text(
                              'JWT Bearer Session Active (${_tokens?['token_type'] ?? 'Bearer'})',
                              style: const TextStyle(
                                color: Colors.indigoAccent,
                                fontSize: 12,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                Card(
                  color: const Color(0xFF1E293B),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                    side: const BorderSide(color: Color(0x1AFFFFFF)),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(20.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Row(
                              children: [
                                Icon(Icons.directions_car, color: Colors.blueAccent, size: 20),
                                SizedBox(width: 8),
                                Text(
                                  'Fleet Vehicle & Telematics',
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.white,
                                  ),
                                ),
                              ],
                            ),
                            IconButton(
                              icon: const Icon(Icons.refresh, size: 18, color: Colors.white60),
                              onPressed: _isLoadingVehicles ? null : _fetchVehicles,
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        if (_isLoadingVehicles) ...[
                          const Center(child: CircularProgressIndicator()),
                        ] else if (_vehicles.isEmpty) ...[
                          const Text(
                            'No vehicles registered or assigned to this tenant yet.',
                            style: TextStyle(color: Colors.white60, fontSize: 13),
                          ),
                        ] else ...[
                          DropdownButtonFormField<String>(
                            value: _selectedVehicle?['id']?.toString(),
                            dropdownColor: const Color(0xFF0F172A),
                            decoration: InputDecoration(
                              labelText: 'Active Delivery Vehicle',
                              labelStyle: const TextStyle(color: Colors.white60, fontSize: 12),
                              filled: true,
                              fillColor: const Color(0xFF0F172A),
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(10),
                                borderSide: const BorderSide(color: Color(0x1AFFFFFF)),
                              ),
                              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                            ),
                            items: _vehicles.map<DropdownMenuItem<String>>((v) {
                              final map = v as Map<String, dynamic>;
                              return DropdownMenuItem<String>(
                                value: map['id']?.toString(),
                                child: Text(
                                  '${map['license_plate']} • ${map['make']} ${map['model']} (${map['vehicle_type']})',
                                  style: const TextStyle(color: Colors.white, fontSize: 13),
                                ),
                              );
                            }).toList(),
                            onChanged: (val) {
                              setState(() {
                                _selectedVehicle = _vehicles.firstWhere(
                                  (v) => (v as Map<String, dynamic>)['id']?.toString() == val,
                                  orElse: () => _vehicles.first,
                                ) as Map<String, dynamic>;
                              });
                            },
                          ),
                          const SizedBox(height: 16),
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: const Color(0xFF0F172A),
                              borderRadius: BorderRadius.circular(10),
                              border: Border.all(color: const Color(0x1AFFFFFF)),
                            ),
                            child: Column(
                              children: [
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    const Text('Power / Battery Level', style: TextStyle(color: Colors.white60, fontSize: 12)),
                                    Text(
                                      '${(_selectedVehicle?['battery_or_fuel_level_percent'] as num?)?.toStringAsFixed(1) ?? '100'}%',
                                      style: const TextStyle(color: Colors.greenAccent, fontWeight: FontWeight.bold, fontSize: 13),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 8),
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    const Text('Odometer Reading', style: TextStyle(color: Colors.white60, fontSize: 12)),
                                    Text(
                                      '${(_selectedVehicle?['current_mileage_km'] as num?)?.toStringAsFixed(1) ?? '0'} km',
                                      style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 8),
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    const Text('Operational Status', style: TextStyle(color: Colors.white60, fontSize: 12)),
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                                      decoration: BoxDecoration(
                                        color: Colors.blueAccent.withOpacity(0.2),
                                        borderRadius: BorderRadius.circular(6),
                                      ),
                                      child: Text(
                                        _selectedVehicle?['status']?.toString() ?? 'AVAILABLE',
                                        style: const TextStyle(color: Colors.blueAccent, fontSize: 11, fontWeight: FontWeight.bold),
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 16),
                          SizedBox(
                            width: double.infinity,
                            child: ElevatedButton.icon(
                              onPressed: _isSendingTelem ? null : _sendLivePing,
                              icon: const Icon(Icons.gps_fixed, size: 16),
                              label: _isSendingTelem
                                  ? const SizedBox(
                                      width: 16,
                                      height: 16,
                                      child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                                    )
                                  : const Text('Send GPS Telemetry Ping'),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: Colors.teal,
                                foregroundColor: Colors.white,
                                padding: const EdgeInsets.symmetric(vertical: 12),
                                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                              ),
                            ),
                          ),
                          if (_telemStatusMsg != null) ...[
                            const SizedBox(height: 8),
                            Text(
                              _telemStatusMsg!,
                              style: TextStyle(
                                color: _telemStatusMsg!.startsWith('Error') ? Colors.redAccent : Colors.tealAccent,
                                fontSize: 12,
                              ),
                            ),
                          ],
                        ],
                      ],
                    ),
                  ),
                ),
              ] else ...[
                Card(
                  color: const Color(0xFF1E293B),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                    side: const BorderSide(color: Color(0x1AFFFFFF)),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(20.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Row(
                          children: [
                            Icon(Icons.badge, color: Colors.indigoAccent, size: 20),
                            SizedBox(width: 8),
                            Text(
                              'Courier Authentication',
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 14),
                        TextField(
                          controller: _emailController,
                          decoration: InputDecoration(
                            labelText: 'Employee Email',
                            labelStyle: const TextStyle(color: Colors.white60, fontSize: 13),
                            filled: true,
                            fillColor: const Color(0xFF0F172A),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(10),
                              borderSide: const BorderSide(color: Color(0x1AFFFFFF)),
                            ),
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                          ),
                          style: const TextStyle(color: Colors.white, fontSize: 14),
                        ),
                        const SizedBox(height: 12),
                        TextField(
                          controller: _passwordController,
                          obscureText: true,
                          decoration: InputDecoration(
                            labelText: 'Password',
                            labelStyle: const TextStyle(color: Colors.white60, fontSize: 13),
                            filled: true,
                            fillColor: const Color(0xFF0F172A),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(10),
                              borderSide: const BorderSide(color: Color(0x1AFFFFFF)),
                            ),
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                          ),
                          style: const TextStyle(color: Colors.white, fontSize: 14),
                        ),
                        if (_authError != null) ...[
                          const SizedBox(height: 10),
                          Text(
                            _authError!,
                            style: const TextStyle(color: Colors.redAccent, fontSize: 12),
                          ),
                        ],
                        const SizedBox(height: 16),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton(
                            onPressed: _isLoggingIn ? null : _handleLogin,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.indigoAccent,
                              foregroundColor: Colors.white,
                              padding: const EdgeInsets.symmetric(vertical: 12),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(10),
                              ),
                            ),
                            child: _isLoggingIn
                                ? const SizedBox(
                                    width: 18,
                                    height: 18,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                      color: Colors.white,
                                    ),
                                  )
                                : const Text(
                                    'Connect & Authorize Courier',
                                    style: TextStyle(fontWeight: FontWeight.bold),
                                  ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
              const SizedBox(height: 20),

              // Foundation Card
              Card(
                color: const Color(0xFF1E293B),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                  side: const BorderSide(color: Color(0x0FFFFFFF)),
                ),
                child: const Padding(
                  padding: EdgeInsets.all(20.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Phase 0 Foundation Ready',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
                      SizedBox(height: 8),
                      Text(
                        'Clean architecture mobile node with JWT authentication, telemetry heartbeat, QR custody tracking, and offline sync.',
                        style: TextStyle(fontSize: 13, color: Colors.white70),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 32),
              const Center(
                child: Text(
                  'LogiFlows v0.1.0 • Connected Mobile Node',
                  style: TextStyle(color: Colors.white38, fontSize: 12),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
