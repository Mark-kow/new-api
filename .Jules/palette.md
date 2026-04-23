## 2024-05-24 - Add ARIA label to Token Table dropdown trigger
**Learning:** Found a pattern where SplitButtonGroup contains a main action button and a dropdown trigger button with just an icon (`<IconTreeTriangleDown />`), but the trigger button lacked an `aria-label`. This makes it difficult for screen reader users to understand the purpose of the dropdown button.
**Action:** Always verify that icon-only buttons, especially those acting as dropdown triggers in button groups, have appropriate `aria-label` attributes to ensure accessibility.
