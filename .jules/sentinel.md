## 2024-05-24 - [Fix Silent Failure in controller/user.go DeleteUser]
**Vulnerability:** In `controller/user.go`'s `DeleteUser` method, if `model.HardDeleteUserById(id)` returned an error, the error was completely swallowed and the method returned `{"success": true}`.
**Learning:** Returning `success: true` on failed deletion provides misleading responses, which can be an operational or security issue (e.g. attempting to delete an admin but it fails, yet caller thinks it succeeded).
**Prevention:** In Go Gin controllers, always correctly check for errors and return `success: false` or use `common.ApiError(c, err)` instead of masking failures.
