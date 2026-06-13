import { translation as enTranslation } from './en/translation';
import { cEndChat as enCEndChat } from './en/cEndChat';
import { translation as zhCNTranslation } from './zh-CN/translation';
import { cEndChat as zhCNCEndChat } from './zh-CN/cEndChat';
import { translation as jaTranslation } from './ja/translation';
import { cEndChat as jaCEndChat } from './ja/cEndChat';
import { translation as koTranslation } from './ko/translation';
import { cEndChat as koCEndChat } from './ko/cEndChat';
import { translation as frTranslation } from './fr/translation';
import { cEndChat as frCEndChat } from './fr/cEndChat';
import { translation as esTranslation } from './es/translation';
import { cEndChat as esCEndChat } from './es/cEndChat';

function mergeResources(...resources: Array<{ translation: Record<string, string> }>) {
  return resources.reduce<{ translation: Record<string, string> }>((acc, current) => ({
    translation: {
      ...acc.translation,
      ...current.translation,
    },
  }), { translation: {} });
}

function mergeWithEnglishBase(...resources: Array<{ translation: Record<string, string> }>) {
  return {
    translation: {
      ...mergeResources(enTranslation, enCEndChat).translation,
      ...mergeResources(...resources).translation,
    },
  };
}

export const TRANSLATION_RESOURCES = {
  en: mergeResources(enTranslation, enCEndChat),
  'zh-CN': mergeResources(zhCNTranslation, zhCNCEndChat),
  ja: mergeWithEnglishBase(jaTranslation, jaCEndChat),
  ko: mergeWithEnglishBase(koTranslation, koCEndChat),
  fr: mergeWithEnglishBase(frTranslation, frCEndChat),
  es: mergeWithEnglishBase(esTranslation, esCEndChat),
};
