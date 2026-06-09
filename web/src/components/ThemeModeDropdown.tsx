import React from 'react';
import { Button, Dropdown } from 'antd';
import type { MenuProps } from 'antd';
import { BgColorsOutlined, CheckOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useThemeMode } from '../theme/themeMode';
import type { ThemeMode } from '../theme/themeMode';

interface ThemeModeDropdownProps {
  buttonClassName?: string;
  compact?: boolean;
}

const ThemeModeDropdown: React.FC<ThemeModeDropdownProps> = ({ buttonClassName, compact = false }) => {
  const { t } = useTranslation();
  const { themeMode, setThemeMode } = useThemeMode();

  const items: MenuProps['items'] = (['system', 'light', 'dark'] as ThemeMode[]).map((mode) => ({
    key: mode,
    icon: themeMode === mode ? <CheckOutlined /> : <span style={{ width: 14, display: 'inline-block' }} />,
    label: t(`theme_mode_${mode}`),
  }));

  return (
    <Dropdown
      menu={{
        items,
        selectable: true,
        selectedKeys: [themeMode],
        onClick: ({ key }) => setThemeMode(key as ThemeMode),
      }}
      trigger={['click']}
    >
      <Button
        type="text"
        icon={<BgColorsOutlined />}
        className={buttonClassName}
        data-testid="theme-mode-trigger"
        aria-label={t('theme_mode_label')}
        title={t('theme_mode_label')}
      >
        {!compact && t(`theme_mode_${themeMode}`)}
      </Button>
    </Dropdown>
  );
};

export default ThemeModeDropdown;
