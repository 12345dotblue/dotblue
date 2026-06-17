export const ANALYTICS_EVENTS = {
  LOGIN_START: 'login_start',
  LOGIN_SUCCESS: 'login_success',
  SIGNUP_START: 'signup_start',
  SIGNUP_SUCCESS: 'signup_success',
  LOGOUT: 'logout',

  DOC_PAGE_VIEW: 'doc_page_view',
  DOC_SECTION_CLICK: 'doc_section_click',
  DOC_EXTERNAL_LINK_CLICK: 'doc_external_link_click',

  NAV_CLICK: 'nav_click',
  LANGUAGE_SWITCH: 'language_switch',

  CTA_CLICK: 'cta_click',
  PRICING_PLAN_CLICK: 'pricing_plan_click',

  ERROR_PAGE_VIEW: 'error_page_view',
} as const;

export type AnalyticsEventName = (typeof ANALYTICS_EVENTS)[keyof typeof ANALYTICS_EVENTS];
