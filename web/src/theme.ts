import type { GlobalThemeOverrides } from 'naive-ui';

export const themeOverrides: GlobalThemeOverrides = {
    common: {
        primaryColor: '#d97757',
        primaryColorHover: '#e68a6b',
        primaryColorPressed: '#c26546',
        primaryColorSuppl: '#e68a6b',

        fontFamily: '"Inter", sans-serif',
        fontFamilyMono: 'monospace',
    },
    Button: {
        borderRadiusMedium: '6px',
        textColorPrimary: '#faf9f5',
    },
    Input: {
        borderRadius: '6px',
    },
    Card: {
        borderRadius: '8px',
    }
};

export const lightPalette = {
    background: '#faf9f5',
    text: '#141413',
    secondaryText: '#b0aea5',
    accent: '#d97757',
    sidebar: '#e8e6dc'
};

export const darkPalette = {
    background: '#191919',
    text: '#faf9f5',
    secondaryText: '#888888',
    accent: '#d97757',
    sidebar: '#202020'
};
