# Code Self Study Website

This is the new [Code Self Study](https://codeselfstudy.com/) website.

Attend a meetup to find out how to contribute. :construction:

See the documentation in [manual](./manual/src/SUMMARY.md).

## Goals:

- help people in the group find something in common to work on
- meetup activity

### Ideas so far:

- send a coding puzzle of the day into the forum, slack, and/or the browser extension so that interested people have a common task to discuss
- add new puzzles (links to other sites or original puzzles)
- mark puzzles that you've completed
- saving puzzles to do later
- use the browser extension to add new puzzles to the database?
- fetch puzzle by difficulty and type of problem
- commenting? (forum integration or separate)
- voting
- browser extension integration

## Licenses

The code is licensed under BSD 3-Clause license. Some of the subdirectories are not licensed under BSD 3-Clause license. To see the licenses of individual subdirectories, please look for the `LICENSE.md` files in subdirectories.

You can use this command to see their locations:

```bash
find . -name "LICENSE.md" -not -path "*/node_modules/*"
```

or with `just`:

```bash
just find_licenses
```

Basically, all the computer code is BSD 3-Clause licensed, except that the website's rendered text content and images are not licensed and cannot be reused, because they are unique to this website and brand (Code Self Study).
