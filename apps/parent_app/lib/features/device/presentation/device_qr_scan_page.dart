import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../application/device_binding_controller.dart';

/// Scans the one-time provisioning QR shown by a display-equipped device.
class DeviceQrScanPage extends ConsumerStatefulWidget {
  const DeviceQrScanPage({super.key});

  @override
  ConsumerState<DeviceQrScanPage> createState() => _DeviceQrScanPageState();
}

class _DeviceQrScanPageState extends ConsumerState<DeviceQrScanPage> {
  final _controller = MobileScannerController(
    formats: const [BarcodeFormat.qrCode],
    detectionSpeed: DetectionSpeed.noDuplicates,
  );
  bool _isHandlingResult = false;
  String? _errorMessage;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _handleBarcode(BarcodeCapture capture) async {
    if (_isHandlingResult) {
      return;
    }
    final rawValue = capture.barcodes.firstOrNull?.rawValue;
    if (rawValue == null || rawValue.isEmpty) {
      return;
    }
    _isHandlingResult = true;
    await _controller.stop();

    final payload = _parsePayload(rawValue);
    if (payload == null) {
      setState(() => _errorMessage = '这不是初芽设备上的绑定码，请重新扫描');
      _isHandlingResult = false;
      await _controller.start();
      return;
    }
    try {
      await ref
          .read(deviceBindingControllerProvider.notifier)
          .bindToken(
            token: payload.token,
            deviceName: payload.deviceName,
            hardwareModel: payload.hardwareModel,
          );
      if (mounted) {
        context.go('/devices');
      }
    } on Object {
      if (mounted) {
        setState(() => _errorMessage = '绑定没有完成，请确认绑定码仍然有效');
      }
      _isHandlingResult = false;
      await _controller.start();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('扫描设备绑定码')),
      body: Stack(
        children: [
          MobileScanner(
            controller: _controller,
            onDetect: _handleBarcode,
            errorBuilder: (context, error) => const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Text('无法使用相机，请在系统设置中允许相机权限后重试。'),
              ),
            ),
          ),
          const Center(child: _ScannerFrame()),
          if (_errorMessage != null)
            Align(
              alignment: Alignment.bottomCenter,
              child: Container(
                width: double.infinity,
                margin: const EdgeInsets.all(24),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.errorContainer,
                  borderRadius: BorderRadius.circular(18),
                ),
                child: Text(_errorMessage!),
              ),
            ),
        ],
      ),
    );
  }
}

class _ScannerFrame extends StatelessWidget {
  const _ScannerFrame();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 240,
      height: 240,
      decoration: BoxDecoration(
        border: Border.all(
          color: Theme.of(context).colorScheme.primary,
          width: 3,
        ),
        borderRadius: BorderRadius.circular(24),
      ),
    );
  }
}

class _ProvisioningPayload {
  const _ProvisioningPayload({
    required this.token,
    required this.deviceName,
    required this.hardwareModel,
  });

  final String token;
  final String deviceName;
  final String hardwareModel;
}

_ProvisioningPayload? _parsePayload(String rawValue) {
  final uri = Uri.tryParse(rawValue);
  if (uri == null || uri.scheme != 'sprout') {
    return null;
  }
  if (uri.host != 'device' && uri.path != '/device') {
    return null;
  }
  final token = uri.queryParameters['token'] ?? '';
  final deviceID = uri.queryParameters['device_id'] ?? '';
  if (token.isEmpty || deviceID.isEmpty) {
    return null;
  }
  return _ProvisioningPayload(
    token: token,
    deviceName: uri.queryParameters['name'] ?? '初芽',
    hardwareModel: uri.queryParameters['hardware'] ?? '',
  );
}
