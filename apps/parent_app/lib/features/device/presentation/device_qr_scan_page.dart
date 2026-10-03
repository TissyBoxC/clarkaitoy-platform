import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../../../core/theme/app_motion.dart';
import '../../../shared/widgets/app_reveal.dart';
import '../application/device_binding_controller.dart';
import '../domain/device_payload.dart';

typedef DeviceScannerViewBuilder =
    Widget Function(BuildContext context, MobileScannerController controller);

/// Scans the one-time provisioning QR shown by a display-equipped device.
class DeviceQrScanPage extends ConsumerStatefulWidget {
  const DeviceQrScanPage({super.key, this.scannerViewBuilder});

  /// Test seam that keeps camera platform channels out of widget tests.
  final DeviceScannerViewBuilder? scannerViewBuilder;

  @override
  ConsumerState<DeviceQrScanPage> createState() => _DeviceQrScanPageState();
}

class _DeviceQrScanPageState extends ConsumerState<DeviceQrScanPage> {
  final _controller = MobileScannerController(
    formats: const [BarcodeFormat.qrCode],
    detectionSpeed: DetectionSpeed.noDuplicates,
  );
  bool _isHandlingResult = false;
  bool _isLeaving = false;
  String? _errorMessage;

  @override
  void dispose() {
    _controller.stop().then((_) => _controller.dispose());
    super.dispose();
  }

  Future<void> _handleBarcode(BarcodeCapture capture) async {
    if (_isHandlingResult || _isLeaving) {
      return;
    }
    final rawValue = capture.barcodes.firstOrNull?.rawValue;
    if (rawValue == null || rawValue.isEmpty) {
      return;
    }
    _isHandlingResult = true;
    await _controller.stop();
    if (!mounted || _isLeaving) {
      return;
    }

    final payload = DevicePayload.tryParse(rawValue);
    if (payload == null) {
      setState(() => _errorMessage = '这不是初芽设备上的绑定码，请重新扫描');
      _isHandlingResult = false;
      if (mounted && !_isLeaving) {
        await _controller.start();
      }
      return;
    }
    try {
      if (payload is DeviceSetupPayload) {
        if (mounted) {
          await context.push('/devices/provision', extra: payload);
        }
        _isHandlingResult = false;
        if (mounted && !_isLeaving) {
          await _controller.start();
        }
        return;
      }
      final bindingPayload = payload as DeviceBindingPayload;
      await ref
          .read(deviceBindingControllerProvider.notifier)
          .bindToken(
            token: bindingPayload.bindingToken,
            deviceName: bindingPayload.deviceName,
          );
      if (mounted) {
        context.go('/devices');
      }
    } on Object {
      if (mounted && !_isLeaving) {
        setState(() => _errorMessage = '绑定没有完成，请确认绑定码仍然有效');
      }
      _isHandlingResult = false;
      if (mounted && !_isLeaving) {
        await _controller.start();
      }
    }
  }

  Future<void> _leave() async {
    if (_isLeaving) {
      return;
    }
    _isLeaving = true;
    await _controller.stop();
    if (mounted) {
      Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, _) {
        if (!didPop) {
          _leave();
        }
      },
      child: Scaffold(
        appBar: AppBar(
          title: const Text('扫描设备绑定码'),
          leading: IconButton(
            tooltip: '返回',
            onPressed: _leave,
            icon: const Icon(Icons.arrow_back_rounded),
          ),
        ),
        body: Stack(
          children: [
            _scannerView(context),
            const Center(child: _ScannerFrame()),
            AnimatedSlide(
              offset: _errorMessage == null
                  ? const Offset(0, 0.18)
                  : Offset.zero,
              duration: AppMotion.standard,
              curve: AppMotion.enterCurve,
              child: AnimatedOpacity(
                opacity: _errorMessage == null ? 0 : 1,
                duration: AppMotion.fast,
                child: Align(
                  alignment: Alignment.bottomCenter,
                  child: Container(
                    width: double.infinity,
                    margin: const EdgeInsets.all(24),
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.errorContainer,
                      borderRadius: BorderRadius.circular(18),
                    ),
                    child: Text(_errorMessage ?? ''),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _scannerView(BuildContext context) {
    final builder = widget.scannerViewBuilder;
    if (builder != null) {
      return builder(context, _controller);
    }
    return MobileScanner(
      controller: _controller,
      onDetect: _handleBarcode,
      errorBuilder: (context, error) => const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Text('无法使用相机，请在系统设置中允许相机权限后重试。'),
        ),
      ),
    );
  }
}

class _ScannerFrame extends StatelessWidget {
  const _ScannerFrame();

  @override
  Widget build(BuildContext context) {
    return AppReveal(
      child: Container(
        width: 240,
        height: 240,
        decoration: BoxDecoration(
          border: Border.all(
            color: Theme.of(context).colorScheme.primary,
            width: 3,
          ),
          borderRadius: BorderRadius.circular(24),
          boxShadow: [
            BoxShadow(
              color: Theme.of(
                context,
              ).colorScheme.primary.withValues(alpha: 0.18),
              blurRadius: 28,
              spreadRadius: 2,
            ),
          ],
        ),
      ),
    );
  }
}
