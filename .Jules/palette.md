## 2025-04-28 - Added ARIA labels to IconDelete in JSONEditor
**Learning:** Found that `JSONEditor` component uses `@douyinfe/semi-icons`'s `IconDelete` inside a `Button` without an `aria-label` attribute. This is a common pattern that makes icon-only buttons inaccessible to screen readers.
**Action:** Always check `icon={<Icon... />}` usages inside `Button` components to ensure they have descriptive `aria-label` attributes for better accessibility.
