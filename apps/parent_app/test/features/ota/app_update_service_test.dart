import 'dart:io';

import 'package:archive/archive.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/core/error/app_exception.dart';
import 'package:parent_app/features/auth/data/auth_api.dart';
import 'package:parent_app/features/ota/application/app_update_service.dart';
import 'package:parent_app/features/ota/application/resource_archive_extractor.dart';

void main() {
  group('AppUpdateService.updateProblem', () {
    test('accepts a complete HTTPS package specification', () {
      final service = AppUpdateService();

      expect(service.updateProblem(_update()), isNull);
    });

    test('rejects a non-HTTPS download address', () {
      final service = AppUpdateService();

      expect(
        service.updateProblem(_update(downloadUrl: 'http://example.com/a.zip')),
        '更新地址不安全，请联系客服处理',
      );
    });

    test('rejects a missing checksum', () {
      final service = AppUpdateService();

      expect(service.updateProblem(_update(sha256: '')), '更新校验信息不完整，请稍后再试');
    });
  });

  group('resource archive boundary', () {
    test('rejects a traversing archive entry before writing it', () async {
      final supportDirectory = await Directory.systemTemp.createTemp(
        'sprout-update-test-',
      );
      addTearDown(() async {
        if (await supportDirectory.exists()) {
          await supportDirectory.delete(recursive: true);
        }
      });

      final archive = Archive()
        ..add(ArchiveFile.string('../outside.txt', 'do not write'));
      final archiveBytes = ZipEncoder().encode(archive);
      final packageFile = File('${supportDirectory.path}/package.zip');
      await packageFile.writeAsBytes(archiveBytes);
      final destination = Directory('${supportDirectory.path}/unpacked')
        ..createSync();

      await expectLater(
        ResourceArchiveExtractor().extract(
          archiveFile: packageFile,
          destination: destination,
        ),
        throwsA(isA<AppException>()),
      );
      expect(
        File('${supportDirectory.path}/outside.txt').existsSync(),
        isFalse,
      );
    });

    test('extracts a normal archive inside the destination', () async {
      final supportDirectory = await Directory.systemTemp.createTemp(
        'sprout-update-test-',
      );
      addTearDown(() async {
        if (await supportDirectory.exists()) {
          await supportDirectory.delete(recursive: true);
        }
      });

      final archive = Archive()
        ..add(ArchiveFile.string('content/story.txt', 'new story'));
      final archiveBytes = ZipEncoder().encode(archive);
      final packageFile = File('${supportDirectory.path}/package.zip');
      await packageFile.writeAsBytes(archiveBytes);
      final destination = Directory('${supportDirectory.path}/unpacked')
        ..createSync();

      await ResourceArchiveExtractor().extract(
        archiveFile: packageFile,
        destination: destination,
      );

      expect(
        File('${destination.path}/content/story.txt').readAsStringSync(),
        'new story',
      );
    });
  });
}

AppUpdateInfo _update({String? downloadUrl, String? sha256}) {
  return AppUpdateInfo(
    kind: 'resource',
    version: '1.0.1',
    downloadUrl: downloadUrl ?? 'https://example.com/resource.zip',
    sha256: sha256 ?? 'A' * 64,
    releaseNotes: '更新内容',
    isMandatory: false,
    minSupportedVersion: '1.0.0',
  );
}
