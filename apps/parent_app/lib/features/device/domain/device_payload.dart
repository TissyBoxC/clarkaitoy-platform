/// Device link parsed from a QR code shown by a 初芽 device.
sealed class DevicePayload {
  const DevicePayload();

  /// Parses both first-run setup codes and completed binding codes.
  static DevicePayload? tryParse(String rawValue) {
    final uri = Uri.tryParse(rawValue.trim());
    if (uri == null || uri.scheme != 'sprout') {
      return null;
    }
    if (uri.host == 'setup' || uri.path == '/setup') {
      final serviceName = uri.queryParameters['service']?.trim() ?? '';
      final proofOfPossession = uri.queryParameters['pop']?.trim() ?? '';
      final username =
          uri.queryParameters['user']?.trim() ?? 'sprout-provisioning';
      if (serviceName.isEmpty || proofOfPossession.isEmpty) {
        return null;
      }
      return DeviceSetupPayload(
        serviceName: serviceName,
        proofOfPossession: proofOfPossession,
        username: username,
      );
    }
    if (uri.host == 'device' || uri.path == '/device') {
      final deviceID = uri.queryParameters['device_id']?.trim() ?? '';
      final bindingToken = uri.queryParameters['token']?.trim() ?? '';
      if (deviceID.isEmpty || bindingToken.isEmpty) {
        return null;
      }
      return DeviceBindingPayload(
        deviceID: deviceID,
        bindingToken: bindingToken,
        deviceName: uri.queryParameters['name']?.trim() ?? '初芽',
      );
    }
    return null;
  }
}

/// First-run payload used to join the local device and configure Wi-Fi.
class DeviceSetupPayload extends DevicePayload {
  const DeviceSetupPayload({
    required this.serviceName,
    required this.proofOfPossession,
    required this.username,
  });

  final String serviceName;
  final String proofOfPossession;
  final String username;
}

/// Completed payload used to bind the device to the signed-in guardian.
class DeviceBindingPayload extends DevicePayload {
  const DeviceBindingPayload({
    required this.deviceID,
    required this.bindingToken,
    required this.deviceName,
  });

  final String deviceID;
  final String bindingToken;
  final String deviceName;
}
