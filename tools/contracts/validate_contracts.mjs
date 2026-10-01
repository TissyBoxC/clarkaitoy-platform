import { readFile, readdir } from 'node:fs/promises';
import { dirname, extname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import Ajv2020 from 'ajv/dist/2020.js';
import addFormats from 'ajv-formats';

const contractsRoot = resolve(
  dirname(fileURLToPath(import.meta.url)),
  '..',
  '..',
  'packages',
  'contracts',
);

const contractNamespace = 'https://sprout.example/contracts/';
const jsonFiles = [];

async function collectJsonFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = join(directory, entry.name);
    if (entry.isDirectory()) {
      await collectJsonFiles(fullPath);
      continue;
    }
    if (extname(entry.name) === '.json') {
      jsonFiles.push(fullPath);
    }
  }
}

await collectJsonFiles(contractsRoot);

const schemas = [];
const documents = [];

for (const filePath of jsonFiles.sort()) {
  const content = await readFile(filePath, 'utf8');
  let document;
  try {
    document = JSON.parse(content);
  } catch (error) {
    throw new Error(`${filePath}: invalid JSON: ${error.message}`);
  }

  if (filePath.endsWith('device_capabilities.json')) {
    documents.push({ filePath, document });
    continue;
  }

  if (
    document.$schema !== 'https://json-schema.org/draft/2020-12/schema' ||
    typeof document.$id !== 'string' ||
    !document.$id.startsWith(contractNamespace)
  ) {
    throw new Error(`${filePath}: expected a draft 2020-12 schema in the contract namespace`);
  }

  schemas.push({ filePath, schema: document });
}

const ajv = new Ajv2020({
  allErrors: true,
  strict: true,
});
addFormats(ajv);

for (const { filePath, schema } of schemas) {
  try {
    ajv.addSchema(schema, schema.$id);
  } catch (error) {
    throw new Error(`${filePath}: failed to register schema: ${error.message}`);
  }
}

for (const { filePath, schema } of schemas) {
  try {
    ajv.compile(schema);
  } catch (error) {
    throw new Error(`${filePath}: invalid schema: ${error.message}`);
  }
}

const capabilityFile = documents.find(({ filePath }) =>
  filePath.endsWith('device_capabilities.json'),
);
if (!capabilityFile) {
  throw new Error('device_capabilities.json was not found');
}

const capabilitySchema = {
  $schema: 'https://json-schema.org/draft/2020-12/schema',
  $id: `${contractNamespace}capabilities/device_capabilities.schema.json`,
  type: 'object',
  additionalProperties: false,
  required: ['schema_version', 'capabilities'],
  properties: {
    schema_version: {
      type: 'string',
      const: '1.0.0',
    },
    capabilities: {
      type: 'array',
      minItems: 1,
      uniqueItems: true,
      items: {
        type: 'string',
        pattern: '^[a-z][a-z0-9_]{1,63}$',
      },
    },
  },
};

const validateCapabilities = ajv.compile(capabilitySchema);
if (!validateCapabilities(capabilityFile.document)) {
  throw new Error(
    `${capabilityFile.filePath}: ${ajv.errorsText(validateCapabilities.errors)}`,
  );
}

console.log(
  `Validated ${schemas.length} schemas and ${documents.length} contract document.`,
);
