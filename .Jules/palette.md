## 2025-03-01 - [Missing ARIA Label on Icon-only Button]
**Learning:** Icon-only buttons used for deletion operations within dynamic form components (like JSONEditor) need explicit `aria-label` properties. Without them, screen readers cannot articulate the button's action, compromising accessibility for visually impaired users interacting with dynamic lists.
**Action:** Add `aria-label={t('删除')}` (or the equivalent translation key) to all icon-only deletion `<Button>` components, ensuring that assistive technologies provide meaningful context.
