import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN.json'
import enUS from './locales/en-US.json'
import jaJP from './locales/ja-JP.json'

// 支持的语言列表
export const supportedLocales = [
    { code: 'zh-CN', name: '简体中文' },
    { code: 'en-US', name: 'English' },
    { code: 'ja-JP', name: '日本語' }
] as const

export type LocaleCode = typeof supportedLocales[number]['code']

// localStorage key
const LOCALE_STORAGE_KEY = 'muxue-locale'

/**
 * 检测用户首选语言
 * 优先级: localStorage > 浏览器语言 > 默认英语
 */
function detectUserLocale(): LocaleCode {
    // 1. 检查 localStorage 中的持久化设置
    const stored = localStorage.getItem(LOCALE_STORAGE_KEY)
    if (stored && supportedLocales.some(l => l.code === stored)) {
        return stored as LocaleCode
    }

    // 2. 检测浏览器语言
    const browserLang = navigator.language || (navigator as any).userLanguage
    if (browserLang) {
        if (browserLang.startsWith('zh')) return 'zh-CN'
        if (browserLang.startsWith('ja')) return 'ja-JP'
        if (browserLang.startsWith('en')) return 'en-US'
    }

    // 3. 默认使用英语
    return 'en-US'
}

// 创建 i18n 实例
export const i18n = createI18n({
    legacy: false, // 使用 Composition API 模式
    locale: detectUserLocale(),
    fallbackLocale: 'en-US',
    messages: {
        'zh-CN': zhCN,
        'en-US': enUS,
        'ja-JP': jaJP
    }
})

/**
 * 切换语言并持久化设置
 */
export function setLocale(locale: LocaleCode) {
    i18n.global.locale.value = locale
    localStorage.setItem(LOCALE_STORAGE_KEY, locale)
}

/**
 * 获取当前语言
 */
export function getLocale(): LocaleCode {
    return i18n.global.locale.value as LocaleCode
}
