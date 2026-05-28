import { createRequire } from 'module';
const require = createRequire(import.meta.url);

// Redirect @actions/core logging to stderr so stdout is pure JSON.
// @actions/core uses process.stdout.write for ::commands and info().
const origStdoutWrite = process.stdout.write.bind(process.stdout);
process.stdout.write = function (chunk, encoding, callback) {
  // Redirect all stdout writes to stderr except our final JSON output
  return process.stderr.write(chunk, encoding, callback);
};

const { DefaultArtifactClient } = require('@actions/artifact');
import { readdirSync, statSync } from 'fs';
import { join } from 'path';

function getAllFiles(dir) {
  const results = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      results.push(...getAllFiles(full));
    } else {
      results.push(full);
    }
  }
  return results;
}

async function main() {
  const [shardDir, artifactName, retentionDays] = process.argv.slice(2);

  if (!shardDir || !artifactName) {
    console.error('Usage: node upload.mjs <shardDir> <artifactName> [retentionDays]');
    process.exit(1);
  }

  const client = new DefaultArtifactClient();
  const options = {};
  if (retentionDays) {
    options.retentionDays = parseInt(retentionDays, 10);
  }

  const files = getAllFiles(shardDir);
  const result = await client.uploadArtifact(artifactName, files, shardDir, options);

  // Restore stdout for the final JSON result
  process.stdout.write = origStdoutWrite;
  process.stdout.write(JSON.stringify({
    id: String(result.id),
    name: artifactName,
    size: result.size,
  }));
}

main().catch(e => {
  console.error(e.message);
  process.exit(1);
});