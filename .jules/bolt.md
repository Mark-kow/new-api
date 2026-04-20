## 2024-04-20 - Layout Performance
**Learning:** Adding React.memo to deeply nested header bar components might not yield benefits if the parent component passes unstable callback references (like onThemeToggle or onLanguageChange). Ensure parent components use useCallback for these handlers to maximize memoization effectiveness.
**Action:** Always verify parent component prop stability before relying purely on React.memo for layout components.
