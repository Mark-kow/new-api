## 2024-05-01 - User Area aria-label
**Learning:** Icon-only buttons or dynamic content buttons like the User Area Dropdown in `UserArea.jsx` lacked a localized `aria-label`, leaving screen readers without context about the button's action. This seems to be a common missing pattern in this app's UI components.
**Action:** Next time, aggressively check custom `Button` usages that wrap visually-dominant elements (like `Avatar` and icons) instead of text labels. Ensure we apply `aria-label={t('xxx')}` to these components.
