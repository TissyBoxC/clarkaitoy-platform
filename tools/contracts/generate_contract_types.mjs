import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import {
  InputData,
  JSONSchemaInput,
  JSONSchemaStore,
  quicktype,
} from 'quicktype-core';

const toolsRoot = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(toolsRoot, '..', '..');
const contractsRoot = resolve(repositoryRoot, 'packages', 'contracts');
const outputRoot = resolve(contractsRoot, 'generated');
const parentAppContractRoot = resolve(
  repositoryRoot,
  'apps',
  'parent_app',
  'lib',
  'core',
  'contracts',
  'generated',
);

const schemaBaseURI = 'https://sprout.example/contracts/';
const sourceSchemaPath = resolve(contractsRoot, 'schemas', 'envelope.schema.json');

const sourceSchema = JSON.parse(await readFile(sourceSchemaPath, 'utf8'));

class LocalSchemaStore extends JSONSchemaStore {
  async fetch(address) {
    const relativePath = address.startsWith(schemaBaseURI)
      ? address.slice(schemaBaseURI.length)
      : null;
    if (relativePath === null || relativePath.includes('..')) {
      return undefined;
    }
    return JSON.parse(await readFile(resolve(contractsRoot, relativePath), 'utf8'));
  }
}

async function generateTypes(language) {
  const inputData = new InputData();
  const schemaInput = new JSONSchemaInput(new LocalSchemaStore(), [], [
    sourceSchema.$id,
  ]);
  await inputData.addSource(
    'schema',
    {
      name: sourceSchema.title,
      schema: JSON.stringify(sourceSchema),
      uris: [sourceSchema.$id],
    },
    () => schemaInput,
  );

  const result = await quicktype({
    inputData,
    lang: language,
    rendererOptions:
      language === 'typescript'
        ? {
            'just-types': 'true',
            'nice-property-names': 'false',
            'prefer-unions': 'true',
          }
        : {
            'just-types': 'true',
            'nice-property-names': 'false',
          },
  });

  let contents = result.lines.join('\n').trimEnd() + '\n';
  if (language === 'dart') {
    contents = normalizeDartSchemaVersion(contents);
  }
  if (!contents.includes('Generated')) {
    contents = `// Generated from ${sourceSchema.$id}; do not edit.\n\n${contents}`;
  }
  return contents;
}

function normalizeDartSchemaVersion(contents) {
  return contents
    .replaceAll('SchemaVersion schemaVersion', 'String schemaVersion')
    .replace(
      /enum SchemaVersion \{\n    THE_100\n\}\n?/,
      'class SchemaVersion {\n'
        + "    static const String the100 = '1.0.0';\n"
        + '}\n',
    );
}

async function writeGeneratedFile(relativePath, contents) {
  const outputPath = resolve(outputRoot, relativePath);
  await mkdir(dirname(outputPath), { recursive: true });
  await writeFile(outputPath, contents, 'utf8');
}

function formatDartFile(filePath) {
  const result = spawnSync('dart', ['format', filePath], {
    encoding: 'utf8',
    stdio: 'pipe',
    // Windows resolves the Dart launcher through a batch file.
    shell: process.platform === 'win32',
  });
  if (result.error !== undefined) {
    throw result.error;
  }
  if (result.status !== 0) {
    throw new Error(result.stderr.trim() || result.stdout.trim());
  }
}

await writeGeneratedFile(
  'typescript/envelope.ts',
  await generateTypes('typescript'),
);
const dartOutputPath = resolve(outputRoot, 'dart', 'envelope.dart');
await writeGeneratedFile(
  'dart/envelope.dart',
  await generateTypes('dart'),
);
// Dart is the only generated language with a stable official formatter.
formatDartFile(dartOutputPath);
await mkdir(parentAppContractRoot, { recursive: true });
await writeFile(
  resolve(parentAppContractRoot, 'envelope.dart'),
  await readFile(dartOutputPath, 'utf8'),
  'utf8',
);

console.log('Generated TypeScript and Dart contract types.');
