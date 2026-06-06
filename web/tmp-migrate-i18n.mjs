import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const localesRoot = path.join(root, 'src', 'i18n', 'locales');
const uiCopyFiles = ['docs', 'login', 'chat', 'agentList', 'agentSkillsPanel', 'adminSettings', 'enterpriseUsage', 'platformUsage', 'imSettings'];
const sourceTransforms = [
  {
    file: path.join(root, 'src', 'domains', 'chat', 'ChatPage.tsx'),
    prefix: 'chat',
    variable: 'chatCopy',
    ensureT: [
      { from: "const { i18n } = useTranslation();", to: "const { t, i18n } = useTranslation();" },
    ],
    removeLines: [
      "  const chatCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).chat;",
      "  const chatCopy = getUiCopy(currentLanguage).chat;",
    ],
  },
  {
    file: path.join(root, 'src', 'domains', 'agent', 'AgentList.tsx'),
    prefix: 'agent_list',
    variable: 'agentListCopy',
    removeLines: ["  const agentListCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).agentList;"],
  },
  {
    file: path.join(root, 'src', 'domains', 'agent', 'AgentSkillsPanel.tsx'),
    prefix: 'agent_skills_panel',
    variable: 'panelCopy',
    removeLines: ["  const panelCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).agentSkillsPanel;"],
  },
  {
    file: path.join(root, 'src', 'domains', 'admin', 'PlatformUsageSettingsCard.tsx'),
    prefix: 'platform_usage',
    variable: 'platformUsageCopy',
    ensureT: [
      { from: "const { i18n } = useTranslation();", to: "const { t, i18n } = useTranslation();" },
    ],
    removeLines: ["  const platformUsageCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).platformUsage;"],
  },
  {
    file: path.join(root, 'src', 'domains', 'admin', 'EnterpriseUsageSettingsTab.tsx'),
    prefix: 'enterprise_usage',
    variable: 'enterpriseUsageCopy',
    ensureT: [
      { from: "const { i18n } = useTranslation();", to: "const { t, i18n } = useTranslation();" },
    ],
    removeLines: ["  const enterpriseUsageCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).enterpriseUsage;"],
  },
  {
    file: path.join(root, 'src', 'domains', 'admin', 'AdminSettings.tsx'),
    prefix: 'admin_settings',
    variable: 'adminSettingsCopy',
    removeLines: ["  const adminSettingsCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).adminSettings;"],
  },
  {
    file: path.join(root, 'src', 'domains', 'admin', 'IMSettingsTab.tsx'),
    prefix: 'im_settings',
    variable: 'imCopy',
    removeLines: ["  const imCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).imSettings;"],
  },
  {
    file: path.join(root, 'src', 'domains', 'identity', 'Login.tsx'),
    prefix: 'login',
    variable: 'loginCopy',
    ensureT: [
      { from: "const { i18n } = useTranslation();", to: "const { t, i18n } = useTranslation();" },
    ],
    removeLines: ["  const loginCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).login;"],
  },
  {
    file: path.join(root, 'src', 'domains', 'identity', 'LoginCallback.tsx'),
    prefix: 'login',
    variable: 'loginCopy',
    ensureT: [
      { from: "const { i18n } = useTranslation();", to: "const { t, i18n } = useTranslation();" },
    ],
    removeLines: ["  const loginCopy = getUiCopy(i18n?.resolvedLanguage || i18n?.language).login;"],
  },
  {
    file: path.join(root, 'src', 'domains', 'marketing', 'ProductDocsPage.tsx'),
    prefix: 'docs',
    variable: 'docsCopy',
    removeLines: ["  const docsCopy = getUiCopy(currentLanguage).docs;"],
  },
];

function camelToSnake(input) {
  return input
    .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
    .replace(/([A-Z])([A-Z][a-z])/g, '$1_$2')
    .replace(/[\s-]+/g, '_')
    .toLowerCase();
}

function parseQuotedPairs(content) {
  const pairs = [];
  const regex = /"([^"\\]+)":\s*"((?:\\.|[^"\\])*)"/g;
  let match;
  while ((match = regex.exec(content)) !== null) {
    pairs.push([match[1], JSON.parse(`"${match[2]}"`)]);
  }
  return pairs;
}

function serializeTranslation(entries) {
  const lines = Object.entries(entries).map(([key, value], index, arr) => {
    const suffix = index === arr.length - 1 ? '' : ',';
    return `    ${JSON.stringify(key)}: ${JSON.stringify(value)}${suffix}`;
  });
  return `export const translation = {\n  \"translation\": {\n${lines.join('\n')}\n  }\n};\n`;
}

for (const localeName of fs.readdirSync(localesRoot)) {
  const localeDir = path.join(localesRoot, localeName);
  if (!fs.statSync(localeDir).isDirectory()) continue;
  const translationFile = path.join(localeDir, 'translation.ts');
  if (!fs.existsSync(translationFile)) continue;
  const translationEntries = Object.fromEntries(parseQuotedPairs(fs.readFileSync(translationFile, 'utf8')));
  for (const baseName of uiCopyFiles) {
    const filePath = path.join(localeDir, `${baseName}.ts`);
    if (!fs.existsSync(filePath)) continue;
    for (const [key, value] of parseQuotedPairs(fs.readFileSync(filePath, 'utf8'))) {
      translationEntries[`${camelToSnake(baseName)}_${camelToSnake(key)}`] = value;
    }
  }
  fs.writeFileSync(translationFile, serializeTranslation(translationEntries));
}

for (const transform of sourceTransforms) {
  let content = fs.readFileSync(transform.file, 'utf8');
  content = content.replace(/^.*getUiCopy.*\r?\n/gm, '');
  for (const step of transform.ensureT || []) {
    content = content.replace(step.from, step.to);
  }
  for (const line of transform.removeLines) {
    content = content.replace(`${line}\n`, '');
    content = content.replace(`${line}\r\n`, '');
  }
  const regex = new RegExp(`\\b${transform.variable}\\.([A-Za-z0-9_]+)`, 'g');
  content = content.replace(regex, (_, property) => `t('${transform.prefix}_${camelToSnake(property)}')`);
  fs.writeFileSync(transform.file, content);
}

const docsLibraryFile = path.join(root, 'src', 'domains', 'marketing', 'productDocsLibrary.ts');
let docsLibrary = fs.readFileSync(docsLibraryFile, 'utf8');
docsLibrary = docsLibrary.replace(
  "import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';",
  "import i18n, { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';",
);
docsLibrary = docsLibrary.replace(/^.*getUiCopy.*\r?\n/gm, '');
docsLibrary = docsLibrary.replace("  const docsCopy = getUiCopy(language).docs;", "  const docsT = i18n.getFixedT(resolveSupportedLanguage(language));");
docsLibrary = docsLibrary.replace("  const docsCopy = getUiCopy(resolved).docs;", "  const docsT = i18n.getFixedT(resolved);");
docsLibrary = docsLibrary.replace(/\bdocsCopy\.([A-Za-z0-9_]+)/g, (_, property) => `docsT('docs_${camelToSnake(property)}')`);
fs.writeFileSync(docsLibraryFile, docsLibrary);

const localesIndexFile = path.join(root, 'src', 'i18n', 'locales', 'index.ts');
let localesIndex = fs.readFileSync(localesIndexFile, 'utf8');
localesIndex = localesIndex.replace(/^import \{ uiCopy as .*\r?\n/gm, '');
localesIndex = localesIndex.replace(/^import type \{ SupportedUiLanguage, UiCopy \} from '..\/schema';\r?\n/gm, '');
localesIndex = localesIndex.replace(/\nexport const UI_COPY:[\s\S]*?};\n\n/, '\n');
fs.writeFileSync(localesIndexFile, localesIndex);

console.log('migrated-ui-copy-to-t');
