import type { Template } from "../api/templates";

/**
 * Converts a Template object to a YAML string.
 * Handles the template fields: id, name, description, image,
 * env, ports, volumes, and restart.
 */
export function templateToYaml(t: Partial<Template>): string {
  const lines: string[] = [];

  if (t.id !== undefined) lines.push(`id: ${t.id}`);
  if (t.name !== undefined) lines.push(`name: ${t.name}`);
  if (t.description !== undefined) {
    const desc = t.description;
    if (desc.includes(":") || desc.includes("#") || desc.includes("\n")) {
      lines.push(`description: "${desc.replace(/"/g, '\\"')}"`);
    } else {
      lines.push(`description: ${desc}`);
    }
  }
  if (t.image !== undefined) lines.push(`image: ${t.image}`);

  if (t.env && Object.keys(t.env).length > 0) {
    lines.push("env:");
    for (const [k, v] of Object.entries(t.env)) {
      lines.push(`  ${k}: "${v}"`);
    }
  }

  if (t.ports && t.ports.length > 0) {
    lines.push("ports:");
    for (const p of t.ports) {
      lines.push(`  - host: ${p.host}`);
      lines.push(`    container: ${p.container}`);
    }
  }

  if (t.volumes && t.volumes.length > 0) {
    lines.push("volumes:");
    for (const v of t.volumes) {
      lines.push(`  - name: ${v.name}`);
      lines.push(`    container: ${v.container}`);
    }
  }

  if (t.restart) {
    lines.push(`restart: ${t.restart}`);
  }

  return lines.join("\n");
}

/**
 * Parses a simple YAML string back into a partial Template object.
 * Only handles the specific fields we use: id, name, description, image,
 * env, ports, volumes, restart.
 */
export function yamlToTemplate(yaml: string): Partial<Template> {
  const t: Partial<Template> = {};
  const lines = yaml.split("\n");

  let currentSection: string | null = null;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let currentListItem: any = null;

  for (const rawLine of lines) {
    if (!rawLine) continue;
    const trimmed = rawLine.trim();
    if (trimmed === "" || trimmed.startsWith("#")) continue;

    // Top-level key:value (no leading whitespace)
    const topKeyMatch = rawLine.match(/^(\w[\w-]*):\s*(.*)$/);
    if (topKeyMatch && !rawLine.startsWith(" ")) {
      const key = topKeyMatch[1]!;
      const rawVal = topKeyMatch[2] || "";
      let val = rawVal.trim();
      val = val.replace(/^"(.*)"$/, "$1").replace(/^'(.*)'$/, "$1");

      if (val === "") {
        currentSection = key;
        currentListItem = null;
      } else {
        currentSection = null;
        currentListItem = null;
        switch (key) {
          case "id": t.id = val; break;
          case "name": t.name = val; break;
          case "description": t.description = val.replace(/^"(.*)"$/, "$1"); break;
          case "image": t.image = val; break;
          case "restart": t.restart = val; break;
        }
      }
      continue;
    }

    // Array item under a section: `  - key: value`
    const arrayItemMatch = rawLine.match(/^ {2}- +(\w[\w-]*):\s*(.*)$/);
    if (arrayItemMatch && currentSection) {
      const itemKey = arrayItemMatch[1]!;
      const itemVal = arrayItemMatch[2]!.trim();

      if (currentSection === "ports" && itemKey === "host") {
        if (!t.ports) t.ports = [];
        currentListItem = { host: parseInt(itemVal) || 0, container: 0 };
        t.ports.push(currentListItem);
      } else if (currentSection === "volumes" && itemKey === "name") {
        if (!t.volumes) t.volumes = [];
        currentListItem = { name: itemVal, container: "" };
        t.volumes.push(currentListItem);
      }
      continue;
    }

    // Sub-key under an array item: `      key: value` or `    key: value`
    const subItemMatch = rawLine.match(/^ {4,}(\w[\w-]*):\s*(.*)$/);
    if (subItemMatch && currentListItem) {
      const subKey = subItemMatch[1]!;
      const subVal = subItemMatch[2]!.trim();
      if (currentSection === "ports") {
        if (subKey === "host") currentListItem.host = parseInt(subVal) || 0;
        if (subKey === "container") currentListItem.container = parseInt(subVal) || 0;
      } else if (currentSection === "volumes") {
        if (subKey === "name") currentListItem.name = subVal;
        if (subKey === "container") currentListItem.container = subVal;
      }
      continue;
    }

    // Sub-key under a section: `  key: value` (for env:)
    const sectionKeyMatch = rawLine.match(/^ {2}(\w[\w-]*):\s*(.*)$/);
    if (sectionKeyMatch && currentSection === "env") {
      if (!t.env) t.env = {};
      const envKey = sectionKeyMatch[1]!;
      const envVal = sectionKeyMatch[2]!.trim().replace(/^"(.*)"$/, "$1");
      if (envKey) {
        t.env[envKey] = envVal;
      }
      continue;
    }
  }

  return t;
}