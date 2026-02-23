## 🎯 Description
<!-- Describe your changes in detail here. What does this Pull Request actually solve or fix? -->

## 🛠️ Changes Made
- <!-- List out the major changes... e.g. "Added new CLI command..." -->
-

## ✅ Pre-Merge Checklist
<!-- Check these boxes `[x]` before requesting a review! -->
- [ ] My code strictly adheres to the existing Go style guidelines.
- [ ] My commit messages use the Conventional Commit format (e.g. `feat:`, `fix:`, `docs:`).
- [ ] I have run `golangci-lint run ./...` and fixed any warnings.
- [ ] I have run `go test -race ./...` and my tests pass successfully.
- [ ] (If applicable) I have securely verified that any path arguments passed via MCP cannot escape the `isPathSafe` vault verification logic in `utils.go`.
- [ ] (If applicable) I have updated the Astro Documentation in `docs/src/content/docs` to reflect these features.
