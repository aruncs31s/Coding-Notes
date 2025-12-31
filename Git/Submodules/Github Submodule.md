# Github Submodule

```widgets
type: quote
quote: Always commit from the deepest submodule outward.
author: GPT
```



If i update `Projects Notes.`
```bash
cd "01 Projects"
git status
git add .
git commit -m "Update project notes"
```

```bash
cd ..
git add "01 Projects"
git commit -m "Update 01 Projects submodule reference"
```

```bash
git push
cd "01 Projects"
git push
```

## Submodules Inside Submodule
```
Notes
└── 01 Projects   (submodule)
    └── ML        (submodule)
        └── Docs  (submodule)
```
