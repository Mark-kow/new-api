## 2025-02-18 - Missing ARIA Labels on Icon-only Dropdown Toggles
**Learning:** Icon-only buttons used as dropdown toggles (e.g., `<Button icon={<IconTreeTriangleDown />} />`) frequently lack `aria-label` attributes. Since they only contain an icon, screen readers announce nothing meaningful, creating an accessibility barrier.
**Action:** Always verify that buttons containing only an icon have a descriptive `aria-label` or visually hidden text to communicate their function clearly. Use translation strings (e.g., `aria-label={t('展开选项')}`) for these labels.
