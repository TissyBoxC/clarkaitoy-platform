import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/providers.dart';
import 'features/device/data/device_binding_api.dart';

/// Device binding API used by scanning and BLE setup.
final deviceBindingApiProvider = Provider<DeviceBindingApi>((ref) {
  return DeviceBindingApi(ref.read(apiClientProvider));
});
