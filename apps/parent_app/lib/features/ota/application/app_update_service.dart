import 'dart:async';
import 'dart:io';

import 'package:crypto/crypto.dart';
import 'package:dio/dio.dart';
import 'package:open_filex/open_filex.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:path/path.dart' as path;
import 'package:path_provider/path_provider.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/error/app_exception.dart';
import '../../auth/data/auth_api.dart';
import 'resource_archive_extractor.dart';

/// Download and install progress shown by the update page.
class AppUpdateProgress {
  const AppUpdateProgress({
    required this.stage,
    required this.fraction,
    this.receivedBytes = 0,
    this.totalBytes = 0,
  });

  final AppUpdateStage stage;
  final double fraction;
  final int receivedBytes;
  final int totalBytes;
}

enum AppUpdateStage { downloading, verifying, installing, extracting }

/// Installs client and resource updates after integrity verification.
///
/// Client updates are handed to the Android package installer, which keeps
/// application data in place. Resource updates are extracted into a private
/// versioned directory and become visible only after the current-version
/// marker is atomically replaced.
class AppUpdateService {
  AppUpdateService({Dio? downloadClient})
    : _downloadClient =
          downloadClient ??
          Dio(
            BaseOptions(
              connectTimeout: const Duration(seconds: 20),
              receiveTimeout: const Duration(minutes: 10),
              sendTimeout: const Duration(seconds: 20),
            ),
          );

  static const _resourceRootName = 'resources';
  static const _currentVersionFileName = 'current_version';
  final Dio _downloadClient;

  /// Describes why an advertised update cannot be installed, if any.
  ///
  /// The platform still returns update metadata during infrastructure
  /// migrations, so callers show an actionable state instead of failing only
  /// after the guardian starts the download.
  String? updateProblem(AppUpdateInfo update) {
    if (update.downloadUrl.isEmpty) {
      return '更新还没有准备好，请稍后再试';
    }
    final uri = Uri.tryParse(update.downloadUrl);
    if (uri == null || uri.scheme != 'https' || uri.host.isEmpty) {
      return '更新地址不安全，请联系客服处理';
    }
    if (!RegExp(r'^[0-9a-fA-F]{64}$').hasMatch(update.sha256)) {
      return '更新校验信息不完整，请稍后再试';
    }
    return null;
  }

  /// Returns the installed package version instead of a compile-time fallback.
  Future<String> installedVersion() async {
    final packageInfo = await PackageInfo.fromPlatform();
    return packageInfo.version;
  }

  /// Downloads and opens the APK installer for a verified client package.
  Future<void> installClientUpdate({
    required AppUpdateInfo update,
    required void Function(AppUpdateProgress progress) onProgress,
  }) async {
    if (!Platform.isAndroid) {
      throw const AppException(
        kind: AppErrorKind.serviceUnavailable,
        message: '请在应用商店更新到新版本',
        retryable: false,
      );
    }
    _validateDownloadSpec(update);
    final temporaryDirectory = await getTemporaryDirectory();
    final safeVersion = _safeVersionDirectory(update.version);
    final downloadedFile = File(
      path.join(
        temporaryDirectory.path,
        'sprout-update-$safeVersion-${DateTime.now().microsecondsSinceEpoch}.apk',
      ),
    );

    try {
      await _downloadAndVerify(
        update: update,
        destination: downloadedFile,
        stage: AppUpdateStage.downloading,
        onProgress: onProgress,
      );
      onProgress(
        const AppUpdateProgress(stage: AppUpdateStage.installing, fraction: 1),
      );
      final result = await OpenFilex.open(
        downloadedFile.path,
        type: 'application/vnd.android.package-archive',
      );
      if (result.type != ResultType.done) {
        throw const AppException(
          kind: AppErrorKind.serviceUnavailable,
          message: '没有打开安装页面，请稍后重试',
          retryable: true,
        );
      }
    } finally {
      // The installer copies the package before this future completes on
      // supported Android versions; removing a missing file is harmless.
      if (await downloadedFile.exists()) {
        try {
          await downloadedFile.delete();
        } on FileSystemException {
          // A failed cleanup must not hide the installation outcome.
        }
      }
    }
  }

  /// Downloads, verifies, and activates a resource package.
  ///
  /// A failed download, hash check, or extraction leaves the previous
  /// `current_version` marker untouched.
  Future<void> installResourceUpdate({
    required AppUpdateInfo update,
    required void Function(AppUpdateProgress progress) onProgress,
  }) async {
    _validateDownloadSpec(update);
    final supportDirectory = await getApplicationSupportDirectory();
    final resourcesRoot = Directory(
      path.join(supportDirectory.path, _resourceRootName),
    );
    final stagingDirectory = Directory(
      path.join(
        resourcesRoot.path,
        '.staging-${DateTime.now().microsecondsSinceEpoch}',
      ),
    );
    final versionDirectory = Directory(
      path.join(resourcesRoot.path, _safeVersionDirectory(update.version)),
    );
    final archiveFile = File(
      path.join(
        (await getTemporaryDirectory()).path,
        'sprout-resource-${DateTime.now().microsecondsSinceEpoch}.zip',
      ),
    );

    await resourcesRoot.create(recursive: true);
    await stagingDirectory.create(recursive: true);

    try {
      await _downloadAndVerify(
        update: update,
        destination: archiveFile,
        stage: AppUpdateStage.downloading,
        onProgress: onProgress,
      );
      onProgress(
        const AppUpdateProgress(stage: AppUpdateStage.extracting, fraction: 0),
      );
      await ResourceArchiveExtractor().extract(
        archiveFile: archiveFile,
        destination: stagingDirectory,
      );

      if (await versionDirectory.exists()) {
        await versionDirectory.delete(recursive: true);
      }
      await stagingDirectory.rename(versionDirectory.path);
      await _writeCurrentVersion(resourcesRoot, update.version);
      onProgress(
        const AppUpdateProgress(stage: AppUpdateStage.installing, fraction: 1),
      );
    } on Object {
      if (await stagingDirectory.exists()) {
        await stagingDirectory.delete(recursive: true);
      }
      rethrow;
    } finally {
      if (await archiveFile.exists()) {
        try {
          await archiveFile.delete();
        } on FileSystemException {
          // Cleanup failure is not part of the update result.
        }
      }
    }
  }

  Future<void> _downloadAndVerify({
    required AppUpdateInfo update,
    required File destination,
    required AppUpdateStage stage,
    required void Function(AppUpdateProgress progress) onProgress,
  }) async {
    try {
      await destination.parent.create(recursive: true);
      await _downloadClient.download(
        update.downloadUrl,
        destination.path,
        deleteOnError: true,
        options: Options(
          responseType: ResponseType.stream,
          headers: const {'Accept': 'application/octet-stream'},
        ),
        onReceiveProgress: (receivedBytes, totalBytes) {
          final fraction = totalBytes <= 0
              ? 0.0
              : (receivedBytes / totalBytes).clamp(0.0, 1.0);
          onProgress(
            AppUpdateProgress(
              stage: stage,
              fraction: fraction,
              receivedBytes: receivedBytes,
              totalBytes: totalBytes,
            ),
          );
        },
      );
    } on DioException catch (error) {
      throw AppException(
        kind: AppErrorKind.network,
        message: '下载没有完成，请检查网络后重试',
        retryable: true,
        cause: error,
      );
    }

    onProgress(
      const AppUpdateProgress(stage: AppUpdateStage.verifying, fraction: 1),
    );
    await _verifySha256(destination, update.sha256);
  }

  Future<void> _verifySha256(File file, String expectedSha256) async {
    final digest = await sha256.bind(file.openRead()).first;
    final actual = digest.toString().toLowerCase();
    if (actual != expectedSha256.toLowerCase()) {
      throw const AppException(
        kind: AppErrorKind.serviceUnavailable,
        message: '更新包校验没有通过，请稍后重试',
        retryable: true,
      );
    }
  }

  Future<void> _writeCurrentVersion(
    Directory resourcesRoot,
    String version,
  ) async {
    final marker = File(path.join(resourcesRoot.path, _currentVersionFileName));
    final stagingMarker = File(
      path.join(
        resourcesRoot.path,
        '.current_version-${DateTime.now().microsecondsSinceEpoch}',
      ),
    );
    await stagingMarker.writeAsString(version, flush: true);
    try {
      await stagingMarker.rename(marker.path);
    } on FileSystemException {
      // Android does not replace an existing target on every filesystem.
      if (await marker.exists()) {
        await marker.delete();
      }
      await stagingMarker.rename(marker.path);
    }
  }

  void _validateDownloadSpec(AppUpdateInfo update) {
    final problem = updateProblem(update);
    if (problem != null) {
      throw const AppException(
        kind: AppErrorKind.serviceUnavailable,
        message: '更新信息不完整，请稍后重试',
        retryable: true,
      );
    }
  }

  String _safeVersionDirectory(String version) {
    final sanitized = version.replaceAll(RegExp(r'[^0-9A-Za-z._-]'), '_');
    return sanitized.isEmpty ? 'unknown' : sanitized;
  }
}

/// Shared updater used by the software update page.
final appUpdateServiceProvider = Provider<AppUpdateService>((ref) {
  return AppUpdateService();
});
