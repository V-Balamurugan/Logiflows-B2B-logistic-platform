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
  bool _isLoading = false;

  final TextEditingController _emailController =
      TextEditingController(text: 'courier@speedycourier.com');
  final TextEditingController _passwordController =
      TextEditingController(text: 'Courier@Speedy2026!');
  Map<String, dynamic>? _user;
  Map<String, dynamic>? _tokens;
  String? _authError;
  bool _isLoggingIn = false;

  @override
  void initState() {
    super.initState();
    _refreshStatus();
  }

  Future<void> _refreshStatus() async {
    setState(() {
      _isLoading = true;
    });

    final res = await _apiClient.checkHealth();
    setState(() {
      _apiStatus = res['data']?['status']?.toString() ??
          res['status']?.toString() ??
          'OFFLINE';
      _isLoading = false;
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
  }

  void _handleLogout() {
    setState(() {
      _user = null;
      _tokens = null;
      _authError = null;
    });
  }

  @override
  void dispose() {
    _apiClient.dispose();
    _emailController.dispose();
    _passwordController.dispose();
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
            icon: _isLoading
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.refresh),
            onPressed: _isLoading ? null : _refreshStatus,
            tooltip: 'Refresh Connection Pulse',
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
                  color: isOnline
                      ? const Color(0x1F22C55E)
                      : const Color(0x1FF59E0B),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: isOnline
                        ? const Color(0x4D22C55E)
                        : const Color(0x4DF59E0B),
                  ),
                ),
                child: Row(
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
                            ? 'Authoritative Backend Connected (/api/v1/healthz)'
                            : 'Standby / Gateway Offline',
                        style: TextStyle(
                          color: isOnline ? Colors.greenAccent : Colors.amberAccent,
                          fontWeight: FontWeight.w600,
                          fontSize: 13,
                        ),
                      ),
                    ),
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
                                border: Border.all(
                                    color: const Color(0x6610B981)),
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

              // Architecture Features Card
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
