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
const schemaFiles = [];
const fixtureFiles = [];

async function collectJsonFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = join(directory, entry.name);
    if (entry.isDirectory()) {
      await collectJsonFiles(fullPath);
      continue;
    }
    if (extname(entry.name) !== '.json') {
      continue;
    }
    if (entry.name.endsWith('.schema.json')) {
      schemaFiles.push(fullPath);
      continue;
    }
    if (entry.name.endsWith('.example.json') || entry.name === 'device_capabilities.json') {
      fixtureFiles.push(fullPath);
      continue;
    }
    throw new Error(`${fullPath}: contract JSON files must be schemas or examples`);
  }
}

await collectJsonFiles(contractsRoot);

const schemas = [];

for (const filePath of schemaFiles.sort()) {
  const content = await readFile(filePath, 'utf8');
  let document;
  try {
    document = JSON.parse(content);
  } catch (error) {
    throw new Error(`${filePath}: invalid JSON: ${error.message}`);
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

for (const filePath of fixtureFiles.sort()) {
  const content = await readFile(filePath, 'utf8');
  try {
    JSON.parse(content);
  } catch (error) {
    throw new Error(`${filePath}: invalid JSON: ${error.message}`);
  }
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

function fixtureSchemaPath(fixturePath) {
  if (fixturePath.endsWith('device_capabilities.json')) {
    return fixturePath.replace(/device_capabilities\.json$/, 'device_capabilities.schema.json');
  }
  return fixturePath.replace(/\.example\.json$/, '.schema.json');
}

for (const fixturePath of fixtureFiles.sort()) {
  const schemaPath = fixtureSchemaPath(fixturePath);
  const registeredSchema = schemas.find(({ filePath }) => filePath === schemaPath);
  if (!registeredSchema) {
    throw new Error(`${fixturePath}: matching schema not found at ${schemaPath}`);
  }

  const validate = ajv.getSchema(registeredSchema.schema.$id);
  if (!validate) {
    throw new Error(`${fixturePath}: schema was not compiled`);
  }

  const document = JSON.parse(await readFile(fixturePath, 'utf8'));
  if (!validate(document)) {
    throw new Error(`${fixturePath}: ${ajv.errorsText(validate.errors)}`);
  }
}

console.log(
  `Validated ${schemas.length} schemas and ${fixtureFiles.length} contract examples.`,
);
