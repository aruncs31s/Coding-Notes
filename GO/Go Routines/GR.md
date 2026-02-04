
- [ ] St this.. latter
**improper errgroup context handling in the orchestrator.**

Changes Made:
Fixed errgroup context shadowing in o_transaction.go:

Changed g, ctx := errgroup.WithContext(ctx) to g, gCtx := errgroup.WithContext(ctx) to avoid shadowing the original ctx.
Used gCtx inside goroutines for coordinated cancellation.
Kept the original ctx for operations after g.Wait() completes.
Added context validation and logging:

Added if ctx.Err() != nil checks at the start of generateTransactionPDF and generateTransactionXLSX.
Added logging to track context state and successful operations.
Why This Fixes the Issue:
The errgroup's derived context (gCtx) gets canceled when all goroutines finish, but the original ctx remains valid.
Using the original ctx after g.Wait() prevents premature cancellation during PDF/XLSX generation.
The added checks and logs help identify if the context is still valid before proceeding with document generation.
Restart your server and test the PDF export again. The context should no longer be canceled during generation.