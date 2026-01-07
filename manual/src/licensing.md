# Licensing

The codebase is licensed under BSD 3-Clause. Some subdirectories have their own `LICENSE.md`. Most of the website's rendered text content and images are not licensed for reuse.

To list all license files:

```bash
find . -name "LICENSE.md" -not -path "*/node_modules/*"
```

Or with the `justfile` script:

```bash
just find_licenses
```
