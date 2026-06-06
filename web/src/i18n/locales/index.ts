import { translation as enTranslation } from './en/translation';
import { translation as zhCNTranslation } from './zh-CN/translation';
import { translation as jaTranslation } from './ja/translation';
import { translation as koTranslation } from './ko/translation';
import { translation as frTranslation } from './fr/translation';
import { translation as esTranslation } from './es/translation';

function mergeWithEnglishBase<T extends { translation: Record<string, string> }>(override: T) {
  return {
    translation: {
      ...enTranslation.translation,
      ...override.translation,
    },
  };
}

export const TRANSLATION_RESOURCES = {
  en: enTranslation,
  'zh-CN': zhCNTranslation,
  ja: mergeWithEnglishBase(jaTranslation),
  ko: mergeWithEnglishBase(koTranslation),
  fr: mergeWithEnglishBase(frTranslation),
  es: mergeWithEnglishBase(esTranslation),
};
