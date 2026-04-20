## 2026-04-20 - Added ARIA labels to Icon-only Buttons
**Learning:** Found an accessibility issue pattern specific to this app's components, where `<Button>` components with only an icon (like `<IconMore />`) did not have an `aria-label`. Screen readers would announce them as an empty button or skip them.
**Action:** When adding or encountering icon-only buttons in the design system components (e.g., using Semi UI), always ensure an `aria-label` is provided to explain the action to assistive technologies.
