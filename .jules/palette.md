## 2024-05-02 - Ensure translation hook availability for accessibility labels
**Learning:** When adding `aria-label={t('something')}` to components, it's critical to verify that `useTranslation` is imported and `t` is defined in scope, otherwise the application will crash with a `ReferenceError` resulting in a blank screen.
**Action:** Always run a `grep` for `useTranslation` in the target file before adding translated `aria-label`s, and run a frontend build/Playwright verification immediately to catch runtime crashes.
