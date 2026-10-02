import 'package:esp_provisioning_wifi/esp_provisioning_wifi.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/error/app_exception.dart';
import '../application/device_binding_controller.dart';
import '../domain/device_payload.dart';

/// Guided first-run flow for a display-equipped or nearby 初芽 device.
class DeviceProvisioningPage extends ConsumerStatefulWidget {
  const DeviceProvisioningPage({super.key, required this.setup});

  final DeviceSetupPayload setup;

  @override
  ConsumerState<DeviceProvisioningPage> createState() =>
      _DeviceProvisioningPageState();
}

enum _ProvisioningStep { finding, wifi, password, binding, done }

class _DeviceProvisioningPageState
    extends ConsumerState<DeviceProvisioningPage> {
  _ProvisioningStep _step = _ProvisioningStep.finding;
  List<String> _devices = const [];
  List<EspWifiNetwork> _networks = const [];
  String? _selectedDevice;
  String? _selectedSSID;
  String? _errorMessage;
  final _passwordController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _findDevice();
  }

  @override
  void dispose() {
    _passwordController.dispose();
    ref.read(deviceBindingControllerProvider.notifier).cancelProvisioning();
    super.dispose();
  }

  Future<void> _findDevice() async {
    setState(() {
      _step = _ProvisioningStep.finding;
      _errorMessage = null;
    });
    try {
      final devices = await ref
          .read(deviceBindingControllerProvider.notifier)
          .scanProvisioningDevices();
      if (!mounted) {
        return;
      }
      final matching = devices
          .where((name) => name == widget.setup.serviceName)
          .toList(growable: false);
      setState(() {
        _devices = matching.isEmpty ? devices : matching;
        _selectedDevice = _devices.length == 1 ? _devices.single : null;
      });
    } on Object catch (error) {
      if (mounted) {
        setState(() => _errorMessage = _messageFor(error));
      }
    }
  }

  Future<void> _selectDevice(String deviceName) async {
    setState(() {
      _selectedDevice = deviceName;
      _step = _ProvisioningStep.wifi;
      _errorMessage = null;
      _networks = const [];
    });
    try {
      final networks = await ref
          .read(deviceBindingControllerProvider.notifier)
          .scanWifiNetworks(widget.setup);
      if (!mounted) {
        return;
      }
      setState(() => _networks = networks);
    } on Object catch (error) {
      if (mounted) {
        setState(() => _errorMessage = _messageFor(error));
      }
    }
  }

  Future<void> _selectNetwork(EspWifiNetwork network) async {
    setState(() {
      _selectedSSID = network.ssid;
      _step = network.security == EspWifiSecurity.open
          ? _ProvisioningStep.password
          : _ProvisioningStep.password;
      _errorMessage = null;
    });
    if (network.security == EspWifiSecurity.open) {
      await _connect();
    }
  }

  Future<void> _connect() async {
    final deviceName = _selectedDevice;
    final ssid = _selectedSSID;
    if (deviceName == null || ssid == null) {
      setState(() => _errorMessage = '请先选择要连接的设备');
      return;
    }
    setState(() {
      _step = _ProvisioningStep.binding;
      _errorMessage = null;
    });
    try {
      await ref
          .read(deviceBindingControllerProvider.notifier)
          .provisionWifi(
            setup: widget.setup,
            ssid: ssid,
            password: _passwordController.text,
          );
      final bindingPayload = await ref
          .read(deviceBindingControllerProvider.notifier)
          .readBindingPayload(widget.setup);
      await ref
          .read(deviceBindingControllerProvider.notifier)
          .bindToken(
            token: bindingPayload.bindingToken,
            deviceName: bindingPayload.deviceName,
          );
      if (mounted) {
        setState(() => _step = _ProvisioningStep.done);
      }
    } on Object catch (error) {
      if (mounted) {
        setState(() {
          _step = _ProvisioningStep.password;
          _errorMessage = _messageFor(error);
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('连接新设备')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: switch (_step) {
            _ProvisioningStep.finding => _buildFinding(),
            _ProvisioningStep.wifi => _buildWifiList(),
            _ProvisioningStep.password => _buildPassword(),
            _ProvisioningStep.binding => _buildProgress('正在完成连接…'),
            _ProvisioningStep.done => _buildDone(),
          },
        ),
      ),
    );
  }

  Widget _buildFinding() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const Text('正在寻找初芽'),
        const SizedBox(height: 12),
        const LinearProgressIndicator(),
        if (_devices.isNotEmpty) ...[
          const SizedBox(height: 24),
          Text('找到这些设备', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          ..._devices.map(
            (device) => Card(
              child: ListTile(
                leading: const Icon(Icons.toys_outlined),
                title: Text(device),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => _selectDevice(device),
              ),
            ),
          ),
        ],
        if (_devices.isEmpty && _errorMessage == null) ...[
          const SizedBox(height: 16),
          const Text('请让初芽保持开机，并停留在配网页面。'),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            onPressed: _findDevice,
            icon: const Icon(Icons.refresh),
            label: const Text('重新寻找'),
          ),
        ],
        if (_errorMessage != null) ...[
          const SizedBox(height: 16),
          Text(
            _errorMessage!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
          const SizedBox(height: 12),
          FilledButton(onPressed: _findDevice, child: const Text('重新寻找')),
        ],
      ],
    );
  }

  Widget _buildWifiList() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text('选择要连接的无线网络', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        if (_networks.isEmpty && _errorMessage == null)
          const LinearProgressIndicator(),
        if (_errorMessage != null) ...[
          const SizedBox(height: 16),
          Text(
            _errorMessage!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
          const SizedBox(height: 12),
          FilledButton(
            onPressed: _selectedDevice == null
                ? null
                : () => _selectDevice(_selectedDevice!),
            child: const Text('重新读取网络'),
          ),
        ],
        Expanded(
          child: ListView(
            children: _networks
                .map(
                  (network) => Card(
                    child: ListTile(
                      leading: const Icon(Icons.wifi),
                      title: Text(network.ssid),
                      subtitle: network.rssi == null
                          ? null
                          : Text('信号 ${network.rssi} dBm'),
                      trailing: const Icon(Icons.chevron_right),
                      onTap: () => _selectNetwork(network),
                    ),
                  ),
                )
                .toList(growable: false),
          ),
        ),
      ],
    );
  }

  Widget _buildPassword() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          '输入“${_selectedSSID ?? ''}”的密码',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 16),
        TextField(
          controller: _passwordController,
          obscureText: true,
          autofocus: true,
          decoration: const InputDecoration(labelText: '无线网络密码'),
          onSubmitted: (_) => _connect(),
        ),
        if (_errorMessage != null) ...[
          const SizedBox(height: 12),
          Text(
            _errorMessage!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ],
        const SizedBox(height: 20),
        FilledButton(onPressed: _connect, child: const Text('连接无线网络')),
        const SizedBox(height: 10),
        OutlinedButton(
          onPressed: () => setState(() {
            _step = _ProvisioningStep.wifi;
            _errorMessage = null;
          }),
          child: const Text('选择其他网络'),
        ),
      ],
    );
  }

  Widget _buildProgress(String message) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const CircularProgressIndicator(),
          const SizedBox(height: 16),
          Text(message),
        ],
      ),
    );
  }

  Widget _buildDone() {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.check_circle_outline, size: 56),
          const SizedBox(height: 16),
          const Text('初芽已连接'),
          const SizedBox(height: 20),
          FilledButton(
            onPressed: () => context.go('/devices'),
            child: const Text('查看我的设备'),
          ),
        ],
      ),
    );
  }

  String _messageFor(Object error) {
    if (error is AppException) {
      return error.message;
    }
    return '连接没有完成，请确认初芽已开机并靠近手机';
  }
}
