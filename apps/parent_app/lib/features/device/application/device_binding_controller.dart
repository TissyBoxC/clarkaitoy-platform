import 'package:flutter_blue_plus/flutter_blue_plus.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/error/app_exception.dart';
import '../../../providers.dart';
import '../data/device_binding_api.dart';

/// State for the nearby-device and scanned-token binding flow.
class DeviceBindingState {
  const DeviceBindingState({
    required this.isScanning,
    required this.devices,
    required this.bindings,
    this.errorMessage,
  });

  const DeviceBindingState.initial()
    : isScanning = false,
      devices = const [],
      bindings = const [],
      errorMessage = null;

  final bool isScanning;
  final List<DiscoveredDevice> devices;
  final List<BoundDevice> bindings;
  final String? errorMessage;

  DeviceBindingState copyWith({
    bool? isScanning,
    List<DiscoveredDevice>? devices,
    List<BoundDevice>? bindings,
    String? errorMessage,
  }) {
    return DeviceBindingState(
      isScanning: isScanning ?? this.isScanning,
      devices: devices ?? this.devices,
      bindings: bindings ?? this.bindings,
      errorMessage: errorMessage,
    );
  }
}

/// Device announced over BLE with the minimum metadata needed for selection.
class DiscoveredDevice {
  const DiscoveredDevice({
    required this.id,
    required this.name,
    required this.rssi,
    required this.serviceData,
  });

  final String id;
  final String name;
  final int rssi;
  final Map<String, List<int>> serviceData;
}

/// Owns discovery, scanned-token binding, and bound-device refresh.
class DeviceBindingController extends AsyncNotifier<DeviceBindingState> {
  late final DeviceBindingApi _api;

  @override
  Future<DeviceBindingState> build() async {
    _api = ref.read(deviceBindingApiProvider);
    final bindings = await _api.list();
    return DeviceBindingState(
      isScanning: false,
      devices: const [],
      bindings: bindings,
    );
  }

  Future<void> refresh() async {
    state = await AsyncValue.guard(() async {
      final bindings = await _api.list();
      return state.value?.copyWith(bindings: bindings) ??
          DeviceBindingState(
            isScanning: false,
            devices: const [],
            bindings: bindings,
          );
    });
  }

  Future<void> scanNearby() async {
    final current = state.value ?? const DeviceBindingState.initial();
    state = AsyncData(current.copyWith(isScanning: true, errorMessage: null));
    try {
      await FlutterBluePlus.startScan(timeout: const Duration(seconds: 8));
      final results = await FlutterBluePlus.scanResults.firstWhere(
        (items) => items.isNotEmpty,
        orElse: () => const <ScanResult>[],
      );
      final devices = results
          .where((result) => result.device.platformName.trim().isNotEmpty)
          .map(
            (result) => DiscoveredDevice(
              id: result.device.remoteId.str,
              name: result.device.platformName.trim(),
              rssi: result.rssi,
              serviceData: <String, List<int>>{
                for (final entry
                    in result.advertisementData.serviceData.entries)
                  entry.key.toString(): entry.value,
              },
            ),
          )
          .toList(growable: false);
      state = AsyncData(current.copyWith(isScanning: false, devices: devices));
    } on Object catch (error) {
      state = AsyncData(
        current.copyWith(isScanning: false, errorMessage: _messageFor(error)),
      );
    } finally {
      await FlutterBluePlus.stopScan();
    }
  }

  /// Binds a token extracted from a QR code or BLE provisioning payload.
  Future<BoundDevice> bindToken({
    required String token,
    required String deviceName,
    String hardwareModel = '',
    String firmwareVersion = '',
    List<String> capabilities = const [],
  }) async {
    final binding = await _api.bind(
      token: token,
      deviceName: deviceName,
      hardwareModel: hardwareModel,
      firmwareVersion: firmwareVersion,
      capabilities: capabilities,
    );
    await refresh();
    return binding;
  }

  Future<void> remove(String deviceId) async {
    await _api.remove(deviceId);
    await refresh();
  }
}

String _messageFor(Object error) {
  if (error is AppException) {
    return error.message;
  }
  return '没有找到附近设备，请确认初芽已开机并处于配网状态';
}

final deviceBindingControllerProvider =
    AsyncNotifierProvider<DeviceBindingController, DeviceBindingState>(
      DeviceBindingController.new,
    );
