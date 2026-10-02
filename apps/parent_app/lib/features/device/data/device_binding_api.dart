import '../../../core/network/api_client.dart';

/// Device information returned by the platform binding API.
class BoundDevice {
  const BoundDevice({
    required this.deviceId,
    required this.deviceName,
    required this.hardwareModel,
    required this.firmwareVersion,
    required this.capabilities,
    required this.boundAt,
  });

  final String deviceId;
  final String deviceName;
  final String hardwareModel;
  final String firmwareVersion;
  final List<String> capabilities;
  final DateTime boundAt;

  factory BoundDevice.fromJson(Map<String, Object?> json) {
    final boundAtValue = json['bound_at'];
    return BoundDevice(
      deviceId: _requiredString(json['device_id'], 'device_id'),
      deviceName: _requiredString(json['device_name'], 'device_name'),
      hardwareModel: _asString(json['hardware_model']),
      firmwareVersion: _asString(json['firmware_version']),
      capabilities: _stringList(json['capabilities']),
      boundAt: boundAtValue is String
          ? DateTime.tryParse(boundAtValue) ?? DateTime.now()
          : DateTime.now(),
    );
  }
}

/// Binding API used by the scan and BLE provisioning flows.
class DeviceBindingApi {
  const DeviceBindingApi(this._apiClient);

  final ApiClient _apiClient;

  Future<List<BoundDevice>> list() async {
    final response = await _apiClient.get('/api/v1/devices');
    final data = response['data'] as Map<String, Object?>;
    final devices = data['devices'];
    if (devices is! List) {
      return const [];
    }
    return devices
        .whereType<Map>()
        .map((item) => BoundDevice.fromJson(Map<String, Object?>.from(item)))
        .toList(growable: false);
  }

  /// Requests a one-time binding code from a device that has authenticated
  /// over its local BLE or Wi-Fi provisioning channel.
  Future<DeviceProvisioningTicket> requestProvisioningTicket(
    String deviceId,
    String deviceSessionToken,
  ) async {
    final response = await _apiClient.postWithBearerToken(
      '/api/v1/devices/$deviceId/provisioning-token',
      bearerToken: deviceSessionToken,
    );
    return DeviceProvisioningTicket.fromJson(
      response['data'] as Map<String, Object?>,
    );
  }

  Future<BoundDevice> bind({
    required String token,
    required String deviceName,
  }) async {
    final response = await _apiClient.post(
      '/api/v1/devices/bind',
      body: {'token': token, 'device_name': deviceName},
    );
    return BoundDevice.fromJson(response['data'] as Map<String, Object?>);
  }

  Future<void> remove(String deviceId) async {
    await _apiClient.delete('/api/v1/devices/$deviceId');
  }
}

/// One-time code and expiry issued by an authenticated device.
class DeviceProvisioningTicket {
  const DeviceProvisioningTicket({
    required this.token,
    required this.deviceId,
    required this.expiresAt,
  });

  final String token;
  final String deviceId;
  final DateTime expiresAt;

  factory DeviceProvisioningTicket.fromJson(Map<String, Object?> json) {
    final expiresAtValue = json['expires_at'];
    return DeviceProvisioningTicket(
      token: _requiredString(json['token'], 'token'),
      deviceId: _requiredString(json['device_id'], 'device_id'),
      expiresAt: expiresAtValue is String
          ? DateTime.tryParse(expiresAtValue) ?? DateTime.now()
          : DateTime.now(),
    );
  }
}

String _requiredString(Object? value, String field) {
  if (value is String && value.isNotEmpty) {
    return value;
  }
  throw FormatException('missing $field');
}

String _asString(Object? value) {
  return value is String ? value : '';
}

List<String> _stringList(Object? value) {
  if (value is! List) {
    return const [];
  }
  return value.whereType<String>().toList(growable: false);
}
