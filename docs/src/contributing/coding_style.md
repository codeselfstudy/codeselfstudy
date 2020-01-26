# Coding Style

Please use [Prettier.js](https://prettier.io/) to format the JavaScript, Sass/CSS, and HTML code. Every major editor has a plugin to auto-format the files on save (including [one for vscode](https://marketplace.visualstudio.com/items?itemName=MadsKristensen.JavaScriptPrettier)).

## JavaScript, CSS, HTML

The `.prettierrc` file contains the instructions for the editor. Just use an editor plugin that formats on save, and it will be correct. You can also format it manually by running this command at the root of the project:

```
$ make format
```

Reasoning:

- Other languages that we use (like Elixir) require double quotes for strings, so the JS uses double quotes for consistency.
- 4-space indentation is easier to skim.

This project transpiles the JS so it's okay to use the latest JS features.

## Elixir

The project uses Elixir's auto-formatter. You can set up your editor to format on save or just run this command in the directory where the Elixir code is:

```
$ mix format
```

## Python

Follow PEP8 where possible. Use double quotes for the same reason as the JS. Line length can be 160 chars. (See `tox.ini`.)
