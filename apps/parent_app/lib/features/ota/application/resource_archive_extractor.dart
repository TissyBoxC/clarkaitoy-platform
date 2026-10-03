import 'dart:io';

import 'package:archive/archive.dart';
import 'package:path/path.dart' as path;

import '../../../core/error/app_exception.dart';

/// Extracts a verified resource ZIP inside a caller-owned directory.
///
/// The extractor rejects absolute paths, parent traversal, symlinks, and
/// oversized payloads so a compromised package cannot escape its destination.
class ResourceArchiveExtractor {
  ResourceArchiveExtractor({this.maxExtractedBytes = 512 * 1024 * 1024});

  final int maxExtractedBytes;

  Future<void> extract({
    required File archiveFile,
    required Directory destination,
  }) async {
    final archive = ZipDecoder().decodeBytes(await archiveFile.readAsBytes());
    final rootPath = path.normalize(destination.absolute.path);
    var extractedBytes = 0;

    for (final entry in archive) {
      final relativePath = _safeRelativeArchivePath(entry.name);
      if (relativePath == null || entry.isSymbolicLink) {
        throw const AppException(
          kind: AppErrorKind.serviceUnavailable,
          message: '更新包内容不完整，请稍后重试',
          retryable: false,
        );
      }
      final outputPath = path.normalize(
        path.join(rootPath, path.joinAll(relativePath)),
      );
      if (outputPath != rootPath && !path.isWithin(rootPath, outputPath)) {
        throw const AppException(
          kind: AppErrorKind.serviceUnavailable,
          message: '更新包内容不完整，请稍后重试',
          retryable: false,
        );
      }

      if (entry.isDirectory) {
        await Directory(outputPath).create(recursive: true);
        continue;
      }
      extractedBytes += entry.size;
      if (extractedBytes > maxExtractedBytes) {
        throw const AppException(
          kind: AppErrorKind.serviceUnavailable,
          message: '更新包内容过大，请稍后重试',
          retryable: false,
        );
      }
      final outputFile = File(outputPath);
      await outputFile.parent.create(recursive: true);
      await outputFile.writeAsBytes(entry.content, flush: true);
    }
  }

  List<String>? _safeRelativeArchivePath(String archivePath) {
    final normalized = archivePath.replaceAll('\\', '/');
    if (normalized.startsWith('/') || normalized.contains('\u0000')) {
      return null;
    }
    final parts = normalized.split('/');
    if (parts.isEmpty || parts.any((part) => part == '..')) {
      return null;
    }
    return parts.where((part) => part.isNotEmpty).toList(growable: false);
  }
}
