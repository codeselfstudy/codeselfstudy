# Contributing to Code Self Study

Thank you for your interest in contributing! This guide explains the essentials for contributing to the project.

The project is community-driven and welcomes contributions.

Please attend a meetup to learn more about the project and group.

---

## Getting started

Start by looking at open issues, or open a new issue to suggest improvements if you notice something unclear in the documentation or code.

1. Fork the repository to your own GitHub account using GitHub’s Fork button.
2. Clone your fork to your local machine:

```bash
git clone https://github.com/your-username/codeselfstudy.git
```

3. Change directory into the project:

```bash
cd codeselfstudy
```

4. Follow the [setup steps in the manual](./manual/src/getting-started.md) to install dependencies, set environment variables, and run the dev server.
5. Create a new branch for your changes:

```bash
git checkout -b my-contribution-branch
```

6. Make your changes following the project’s style and guidelines.
7. Add tests for new functionality or bug fixes when practical.
8. If you change behavior or UI, update the manual docs in `manual/` where it makes sense.
9. Run the [quality checks from the manual](./manual/src/tooling-and-quality.md) before opening a PR.
10. Commit your changes:

```bash
git add .
```

```bash
git commit -m "Describe your changes"
```

11. Push your branch to your fork:

```bash
git push origin my-contribution-branch
```

12. Open a Pull Request to the main repository for review.

---

## Making changes

- Keep changes small and focused.
- Write clear commit messages. Use short, sentence-case summaries.
- Follow existing code style where possible.
- Unsure about something? Open an issue to discuss it.

---

## Pull Requests

- Describe what you changed and why.
- Include testing notes (for example: `bun run test`, `bun run lint`).
- Include screenshots for UI changes.
- Reference related issues if applicable.
- Be open to feedback — PRs don't need to be perfect.

---

## Issues

- To work on an existing issue, leave a comment saying you'd like to take it.
- If something is unclear or missing, open a new issue.

---

## Code of conduct

- Be respectful and constructive.
- Everyone is welcome, regardless of experience level.

---

## Contributor responsibilities

By submitting a Pull Request, you confirm that you have the right to contribute the code and that it is original or properly licensed for this project.
