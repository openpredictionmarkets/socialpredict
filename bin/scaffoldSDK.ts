#!/usr/bin/env ts-node

/**
 * Simple scaffolding tool to generate a TypeScript SDK surface
 * from the OpenAPI specification in `backend/docs/openapi.yaml`.
 *
 * This script intentionally generates *stub* functions only – it does not
 * implement any HTTP client logic. The goal is to provide a typed, discoverable
 * surface area that can be filled in by hand or wired up to a client later.
 *
 * Requirements:
 * - Node 18+ (for modern JS features).
 * - The `yaml` package installed and resolvable from this script:
 *   - e.g. `pnpm install yaml` in a package that includes this file.
 *
 * Example usage (from repo root):
 *   pnpx ts-node bin/scaffoldSDK.ts
 *   pnpx ts-node bin/scaffoldSDK.ts --spec backend/docs/openapi.yaml --out sdk
 */

import * as fs from "fs";
import * as path from "path";
// `yaml` is intentionally the only external dependency we rely on here.
// It must be installed separately (see header comment).
import * as YAML from "yaml";

type HttpMethod =
  | "get"
  | "put"
  | "post"
  | "delete"
  | "options"
  | "head"
  | "patch"
  | "trace";

interface OpenAPIOperation {
  operationId?: string;
  tags?: string[];
  summary?: string;
  description?: string;
  // The rest of the OpenAPI fields are intentionally left untyped for now,
  // since this script only needs minimal metadata to scaffold stubs.
  [key: string]: unknown;
}

interface OpenAPISpec {
  paths?: {
    [route: string]: {
      [method in HttpMethod]?: OpenAPIOperation;
    };
  };
  components?: {
    schemas?: {
      // We keep this intentionally loose; the generator inspects the raw shapes.
      [name: string]: any;
    };
  };
  [key: string]: unknown;
}

interface CliOptions {
  specPath: string;
  outDir: string;
  filename: string;
  typesOutDir: string;
  typesFilename: string;
}

function parseArgs(argv: string[]): CliOptions {
  const defaults: CliOptions = {
    specPath: "backend/docs/openapi.yaml",
    outDir: "sdk",
    filename: "index.ts",
    typesOutDir: "sdk",
    typesFilename: "types.ts",
  };

  const args = [...argv];
  const opts: Partial<CliOptions> = {};

  while (args.length > 0) {
    const arg = args.shift();
    if (!arg) continue;
    if (arg === "--spec" || arg === "--specPath") {
      opts.specPath = args.shift() ?? defaults.specPath;
    } else if (arg === "--out" || arg === "--outDir") {
      opts.outDir = args.shift() ?? defaults.outDir;
    } else if (arg === "--filename") {
      opts.filename = args.shift() ?? defaults.filename;
    } else if (arg === "--typesOutDir" || arg === "--typesOut") {
      opts.typesOutDir = args.shift() ?? defaults.typesOutDir;
    } else if (arg === "--typesFilename") {
      opts.typesFilename = args.shift() ?? defaults.typesFilename;
    }
  }

  return {
    specPath: opts.specPath ?? defaults.specPath,
    outDir: opts.outDir ?? defaults.outDir,
    filename: opts.filename ?? defaults.filename,
    typesOutDir: opts.typesOutDir ?? defaults.typesOutDir,
    typesFilename: opts.typesFilename ?? defaults.typesFilename,
  };
}

function readOpenAPISpec(specPath: string): OpenAPISpec {
  const resolved = path.resolve(specPath);
  if (!fs.existsSync(resolved)) {
    throw new Error(`OpenAPI spec not found at ${resolved}`);
  }

  const raw = fs.readFileSync(resolved, "utf8");
  const doc = YAML.parse(raw) as OpenAPISpec;

  if (!doc || typeof doc !== "object" || !doc.paths) {
    throw new Error("Parsed OpenAPI document is missing a 'paths' object");
  }

  return doc;
}

function toCamelCase(input: string): string {
  const cleaned = input
    .trim()
    .replace(/[\s\-]+/g, "_")
    .replace(/[^a-zA-Z0-9_]/g, "_");

  return cleaned
    .split("_")
    .filter(Boolean)
    .map((part, index) => {
      const lower = part.toLowerCase();
      if (index === 0) {
        return lower;
      }
      return lower.charAt(0).toUpperCase() + lower.slice(1);
    })
    .join("");
}

function generateFunctionName(
  route: string,
  method: HttpMethod,
  operation: OpenAPIOperation,
): string {
  if (operation.operationId) {
    return toCamelCase(operation.operationId);
  }

  const tag = operation.tags && operation.tags.length > 0
    ? toCamelCase(operation.tags[0])
    : "";

  // Convert `/v0/markets/{id}/resolve` into something like `v0MarketsByIdResolve`.
  const routeDescriptor = route
    .replace(/^\//, "")
    .replace(/\{([^}]+)\}/g, "by_$1")
    .replace(/\/+/g, "_");

  const methodPrefix = method.toLowerCase();
  const baseName = [tag, methodPrefix, routeDescriptor]
    .filter(Boolean)
    .join("_");

  return toCamelCase(baseName);
}

function toPascalCase(input: string): string {
  const camel = toCamelCase(input);
  if (!camel) return "";
  return camel.charAt(0).toUpperCase() + camel.slice(1);
}

function deriveEndpointLayout(
  route: string,
  method: HttpMethod,
  operation: OpenAPIOperation,
): { version: string; topic: string; functionName: string } {
  const cleaned = route.replace(/^\//, "");
  const segments = cleaned.split("/").filter(Boolean);

  const versionSegment = segments.find((s) => /^v[0-9]+/.test(s)) ?? "v0";

  const topicTag = operation.tags && operation.tags.length > 0
    ? operation.tags[0]
    : "misc";
  const topic = toCamelCase(topicTag);

  const indexOfVersion = segments.indexOf(versionSegment);
  const afterVersion = indexOfVersion >= 0
    ? segments.slice(indexOfVersion + 1)
    : segments;

  const tokens: string[] = [];
  for (const seg of afterVersion) {
    const match = seg.match(/^\{([^}]+)\}$/);
    if (match) {
      tokens.push(`by-${match[1]}`);
    } else {
      tokens.push(seg);
    }
  }

  const slug = tokens.length > 0 ? tokens.join("_") : topicTag;
  const suffix = toPascalCase(slug);
  const prefix = method.toLowerCase();
  const functionName = `${prefix}${suffix}`;

  return {
    version: versionSegment,
    topic,
    functionName,
  };
}

function generateSdkStubs(spec: OpenAPISpec, outBaseDir: string): string {
  const indexLines: string[] = [];

  indexLines.push(
    "// AUTO-GENERATED SDK ENTRYPOINT",
    "// This file was created by bin/scaffoldSDK.ts based on backend/docs/openapi.yaml.",
    "//",
    "// It re-exports per-endpoint helpers generated under sdk/api/*.",
    "",
  );

  const nameCounts = new Map<string, number>();

  for (const [route, pathItem] of Object.entries(spec.paths ?? {})) {
    if (!pathItem) continue;

    for (const method of Object.keys(pathItem) as HttpMethod[]) {
      const lowerMethod = method.toLowerCase() as HttpMethod;
      const operation = pathItem[lowerMethod];
      if (!operation) continue;

      // Keep the health check helper in sdk/config/getHealth.ts.
      if (route === "/health" && lowerMethod === "get") {
        continue;
      }

      const layout = deriveEndpointLayout(route, lowerMethod, operation);

      const baseName = layout.functionName;
      const count = nameCounts.get(baseName) ?? 0;
      nameCounts.set(baseName, count + 1);

      const functionName = count === 0 ? baseName : `${baseName}${count + 1}`;

      const summaryComment = operation.summary || operation.description;

      const responseType = getOperationResponseType(operation);
      const requestType = getOperationRequestType(operation);

      const fileDir = path.join(outBaseDir, "api", layout.version, layout.topic);
      const fileName = `${functionName}.ts`;

      const lines: string[] = [];
      lines.push(
        'import type * as Types from "../../../types.ts";',
        'import { GeneratedSdkContext, sdkRequest } from "../../../helpers.ts";',
        "",
      );

      if (summaryComment) {
        lines.push("/**");
        lines.push(` * ${summaryComment}`);
        lines.push(" *");
        lines.push(
          ` * Generated from ${lowerMethod.toUpperCase()} ${route}`,
        );
        lines.push(" */");
      } else {
        lines.push(
          `// Generated from ${lowerMethod.toUpperCase()} ${route}`,
        );
      }

      lines.push(
        `export async function ${functionName}(`,
        "  ctx: GeneratedSdkContext,",
      );

      if (requestType) {
        lines.push(
          `  params: ${requestType},`,
        );
      } else {
        lines.push(
          "  params?: unknown,",
        );
      }

      lines.push(
        `): Promise<${responseType}> {`,
        `  // TODO: implement via sdkRequest and wire params into path, query, and body.`,
        `  throw new Error("Not implemented: ${functionName} (${lowerMethod.toUpperCase()} ${route})");`,
        "}",
        "",
      );

      writeOutputFile(fileDir, fileName, lines.join("\n"));

      indexLines.push(
        `export { ${functionName} } from "./api/${layout.version}/${layout.topic}/${functionName}.ts";`,
      );
    }
  }

  return indexLines.join("\n");
}

function schemaToTsType(schema: any): string {
  if (!schema || typeof schema !== "object") {
    return "unknown";
  }

  if (schema.$ref && typeof schema.$ref === "string") {
    // We only support local component refs like "#/components/schemas/Foo".
    const ref = schema.$ref;
    const parts = ref.split("/");
    return parts[parts.length - 1] || "unknown";
  }

  const type = schema.type as string | undefined;

  if (schema.enum && Array.isArray(schema.enum) && schema.enum.length > 0) {
    const literals = schema.enum.map((v: unknown) =>
      typeof v === "string" ? JSON.stringify(v) : String(v)
    );
    return literals.join(" | ");
  }

  switch (type) {
    case "string":
      return "string";
    case "integer":
    case "number":
      return "number";
    case "boolean":
      return "boolean";
    case "array": {
      const itemType = schema.items ? schemaToTsType(schema.items) : "unknown";
      return `${itemType}[]`;
    }
    case "object": {
      // If this is a generic map type like { [key: string]: number }.
      if (schema.additionalProperties) {
        const valueType = schemaToTsType(schema.additionalProperties);
        return `{ [key: string]: ${valueType} }`;
      }
      // For inline object shapes without named schema, fall back to a generic record.
      if (!schema.properties) {
        return "Record<string, unknown>";
      }
      // When used as a property type, we keep things simple and treat object
      // with properties as a generic record.
      return "Record<string, unknown>";
    }
    default:
      return "unknown";
  }
}

function schemaToSdkType(schema: any): string | undefined {
  if (!schema || typeof schema !== "object") {
    return undefined;
  }

  if (schema.$ref && typeof schema.$ref === "string") {
    const ref = schema.$ref;
    const parts = ref.split("/");
    const name = parts[parts.length - 1];
    if (name) {
      return `Types.${name}`;
    }
  }

  const type = schema.type as string | undefined;

  if (type === "array" && schema.items) {
    const itemType = schemaToSdkType(schema.items) ?? "unknown";
    return `${itemType}[]`;
  }

   // Fall back to primitive/inline types when no $ref is present.
  if (type === "string" || type === "integer" || type === "number" || type === "boolean") {
    return schemaToTsType(schema);
  }

  return undefined;
}

function getOperationResponseType(operation: any): string {
  const responses = operation?.responses;
  if (!responses || typeof responses !== "object") {
    return "unknown";
  }

  const preferredStatuses = ["200", "201", "202", "204"];
  for (const status of preferredStatuses) {
    const res = responses[status];
    if (!res) continue;

    if (status === "204") {
      return "void";
    }

    const content = res.content;
    if (!content || typeof content !== "object") continue;

    const media =
      content["application/json"] ??
      Object.values<any>(content)[0];
    if (!media || !media.schema) continue;

    const type = schemaToSdkType(media.schema);
    if (type) {
      return type;
    }
  }

  return "unknown";
}

function getOperationRequestType(operation: any): string | undefined {
  const requestBody = operation?.requestBody;
  if (!requestBody) {
    return undefined;
  }

  const content = requestBody.content;
  if (!content || typeof content !== "object") {
    return undefined;
  }

  const media =
    content["application/json"] ??
    Object.values<any>(content)[0];
  if (!media || !media.schema) {
    return undefined;
  }

  return schemaToSdkType(media.schema);
}

function generateTypesFromSchemas(spec: OpenAPISpec): string {
  const schemas = spec.components?.schemas;
  const lines: string[] = [];

  lines.push(
    "// AUTO-GENERATED TYPES",
    "// This file was created by bin/scaffoldSDK.ts based on backend/docs/openapi.yaml.",
    "//",
    "// It mirrors the shapes under components/schemas as TypeScript interfaces",
    "// and types for use by the generated SDK stubs.",
    "",
  );

  if (!schemas || Object.keys(schemas).length === 0) {
    lines.push("// No components.schemas found in the OpenAPI document.");
    return lines.join("\n");
  }

  for (const [name, schema] of Object.entries(schemas)) {
    if (!schema || typeof schema !== "object") continue;

    const description: string | undefined = schema.description;
    const type: string | undefined = schema.type;

    if (description) {
      lines.push("/**");
      lines.push(` * ${description}`);
      lines.push(" */");
    }

    // Object-like schemas become interfaces; everything else becomes a type alias.
    if (type === "object" || schema.properties || schema.additionalProperties) {
      const required = new Set<string>(schema.required ?? []);
      lines.push(`export interface ${name} {`);

      const properties = schema.properties ?? {};
      for (const [propName, propSchema] of Object.entries<any>(properties)) {
        const propType = schemaToTsType(propSchema);
        const optional = required.has(propName) ? "" : "?";
        lines.push(`  ${propName}${optional}: ${propType};`);
      }

      if (schema.additionalProperties) {
        const valueType = schemaToTsType(schema.additionalProperties);
        lines.push(`  [key: string]: ${valueType};`);
      }

      lines.push("}", "");
    } else {
      const tsType = schemaToTsType(schema);
      lines.push(`export type ${name} = ${tsType};`, "");
    }
  }

  return lines.join("\n");
}

function writeOutputFile(outDir: string, filename: string, contents: string) {
  const resolvedDir = path.resolve(outDir);
  fs.mkdirSync(resolvedDir, { recursive: true });

  const outPath = path.join(resolvedDir, filename);
  fs.writeFileSync(outPath, contents, "utf8");
  // eslint-disable-next-line no-console
  console.log(`Wrote SDK stubs to ${outPath}`);
}

function main() {
  const [, , ...argv] = process.argv;
  const opts = parseArgs(argv);

  const spec = readOpenAPISpec(opts.specPath);
  const typeDefs = generateTypesFromSchemas(spec);
  writeOutputFile(opts.typesOutDir, opts.typesFilename, typeDefs);

  const sdkStubs = generateSdkStubs(spec, opts.outDir);
  writeOutputFile(opts.outDir, opts.filename, sdkStubs);
}

// This script is intended to be executed directly (CLI-style), not imported.
// Always run `main()` when the module is evaluated.
void main();
